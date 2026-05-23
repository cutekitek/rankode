package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"rankode/internal/config"
	apierror "rankode/internal/errors"
	"rankode/internal/services/auth"
	"rankode/internal/services/oidc"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	oidcStateCookie        = "oidc_state"
	oidcNonceCookie        = "oidc_nonce"
	oidcCodeVerifierCookie = "oidc_code_verifier"
	oidcProviderCookie     = "oidc_provider"
)

type oidcHandler struct {
	cfg         *config.Config
	oidcService *oidc.Service
	authService *auth.AuthService
}

func NewOIDCHandler(cfg *config.Config, oidcService *oidc.Service, authService *auth.AuthService) *oidcHandler {
	return &oidcHandler{
		cfg:         cfg,
		oidcService: oidcService,
		authService: authService,
	}
}

func (h *oidcHandler) RegisterRoutes(app fiber.Router) {
	oidcGroup := app.Group("/auth/oidc")
	oidcGroup.Get("/login", h.LoginHandler)
	oidcGroup.Get("/:provider/login", h.LoginHandler)
	oidcGroup.Get("/callback", h.CallbackHandler)
	oidcGroup.Get("/:provider/callback", h.CallbackHandler)
}

func (h *oidcHandler) LoginHandler(c fiber.Ctx) error {
	provider, err := h.oidcService.GetProvider(c.Context(), providerNameFromRequest(c))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	state := randomURLSafeString(32)
	nonce := randomURLSafeString(32)
	codeVerifier := randomURLSafeString(64)
	codeChallenge := codeChallengeS256(codeVerifier)

	h.setOIDCCookie(c, oidcStateCookie, state)
	h.setOIDCCookie(c, oidcNonceCookie, nonce)
	h.setOIDCCookie(c, oidcCodeVerifierCookie, codeVerifier)
	h.setOIDCCookie(c, oidcProviderCookie, provider.Name)

	u, err := url.Parse(provider.AuthURL)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", provider.ClientID)
	q.Set("redirect_uri", provider.RedirectURL)
	q.Set("scope", strings.Join(provider.Scopes, " "))
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()

	return c.Redirect().To(u.String())
}

func (h *oidcHandler) CallbackHandler(c fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing OIDC code or state")
	}

	savedState := c.Cookies(oidcStateCookie)
	nonce := c.Cookies(oidcNonceCookie)
	codeVerifier := c.Cookies(oidcCodeVerifierCookie)
	providerName := c.Cookies(oidcProviderCookie)
	if providerName == "" {
		providerName = providerNameFromRequest(c)
	}
	h.clearOIDCCookies(c)

	if savedState == "" || savedState != state {
		return c.Status(fiber.StatusForbidden).SendString("Invalid OIDC state")
	}
	if nonce == "" || codeVerifier == "" {
		return c.Status(fiber.StatusForbidden).SendString("Missing OIDC session cookies")
	}

	provider, err := h.oidcService.GetProvider(c.Context(), providerName)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	user, err := h.oidcService.Authenticate(c.Context(), provider, code, codeVerifier, nonce)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	redirectURL, err := redirectURLWithToken(provider.FrontendRedirectURL, token)
	if err != nil {
		return err
	}
	return c.Redirect().To(redirectURL)
}

func (h *oidcHandler) setOIDCCookie(c fiber.Ctx, name, value string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(10 * time.Minute),
		HTTPOnly: true,
		Secure:   h.cfg.OIDCCookieSecure,
		SameSite: "Lax",
	})
}

func (h *oidcHandler) clearOIDCCookies(c fiber.Ctx) {
	for _, name := range []string{oidcStateCookie, oidcNonceCookie, oidcCodeVerifierCookie, oidcProviderCookie} {
		c.Cookie(&fiber.Cookie{
			Name:     name,
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: true,
			Secure:   h.cfg.OIDCCookieSecure,
			SameSite: "Lax",
		})
	}
}

func providerNameFromRequest(c fiber.Ctx) string {
	name := c.Params("provider")
	if name == "" {
		name = c.Query("provider")
	}
	if strings.TrimSpace(name) == "" {
		return "default"
	}
	return strings.TrimSpace(name)
}

func randomURLSafeString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func codeChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func redirectURLWithToken(rawURL, token string) (string, error) {
	if rawURL == "" {
		rawURL = "/"
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	u.Fragment = "token=" + url.QueryEscape(token)
	return u.String(), nil
}
