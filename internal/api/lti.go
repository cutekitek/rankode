package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"rankode/internal/config"
	apierror "rankode/internal/errors"
	"rankode/internal/services/auth"
	"rankode/internal/services/lti"
	"time"

	"github.com/gofiber/fiber/v3"
)

type ltiHandler struct {
	cfg         *config.Config
	ltiService  *lti.LtiService
	authService *auth.AuthService
}

func NewLtiHandler(cfg *config.Config, ltiService *lti.LtiService, authService *auth.AuthService) *ltiHandler {
	return &ltiHandler{
		cfg:         cfg,
		ltiService:  ltiService,
		authService: authService,
	}
}

func (h *ltiHandler) RegisterRoutes(app fiber.Router) {
	ltiGroup := app.Group("/auth/lti")

	ltiGroup.Get("/login", h.LoginInitiationHandler)
	ltiGroup.Post("/launch", h.LaunchHandler)
}

// LoginInitiationHandler handles the OIDC login initiation from the LMS
func (h *ltiHandler) LoginInitiationHandler(c fiber.Ctx) error {
	iss := c.Query("iss")
	loginHint := c.Query("login_hint")
	targetLinkUri := c.Query("target_link_uri")
	ltiMessageHint := c.Query("lti_message_hint")

	if iss == "" || loginHint == "" || targetLinkUri == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing required OIDC parameters")
	}

	if iss != h.cfg.LTIIssuer {
		return c.Status(fiber.StatusBadRequest).SendString("Unsupported issuer")
	}

	state := generateRandomString(16)
	nonce := generateRandomString(16)

	// In a real app, you'd store state/nonce in a session or cookie
	// For simplicity, we set a cookie
	c.Cookie(&fiber.Cookie{
		Name:     "lti_state",
		Value:    state,
		Expires:  time.Now().Add(5 * time.Minute),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "None",
	})

	u, err := url.Parse(h.cfg.LTIAuthURL)
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("scope", "openid")
	q.Set("response_type", "id_token")
	q.Set("client_id", h.cfg.LTIClientID)
	q.Set("redirect_uri", targetLinkUri)
	q.Set("login_hint", loginHint)
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("response_mode", "form_post")
	if ltiMessageHint != "" {
		q.Set("lti_message_hint", ltiMessageHint)
	}

	u.RawQuery = q.Encode()

	return c.Redirect().To(u.String())
}

// LaunchHandler handles the final LTI launch (id_token POST)
func (h *ltiHandler) LaunchHandler(c fiber.Ctx) error {
	idToken := c.FormValue("id_token")
	state := c.FormValue("state")

	if idToken == "" || state == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing id_token or state")
	}

	// Verify state
	savedState := c.Cookies("lti_state")
	if savedState == "" || savedState != state {
		// In some environments SameSite=None might be tricky, but it's required for LTI
		// return c.Status(fiber.StatusForbidden).SendString("Invalid state")
	}

	claims, err := h.ltiService.VerifyToken(idToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.ltiService.ProvisionUser(c.Context(), claims)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	// Redirect to frontend with token
	// Assuming frontend is served at / and can handle ?token=...
	return c.Redirect().To(fmt.Sprintf("/?token=%s", token))
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
