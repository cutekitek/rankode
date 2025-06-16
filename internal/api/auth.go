package api

import (
	"fmt"
	"rankode/internal/errors"
	"rankode/internal/middleware" 
	"rankode/internal/models"  
	"rankode/internal/services/auth"
	"rankode/internal/services/users"

	"github.com/gofiber/fiber/v3"
)

type authHandler struct {
	usersService *users.UserService
	authService *auth.AuthService
}

func NewAuthHandler(users *users.UserService, auth *auth.AuthService) *authHandler {
	return &authHandler{
		usersService: users,
		authService:  auth,
	}
}

func (h *authHandler) RegisterRoutes(app fiber.Router) {
	authGroup := app.Group("/auth")

	authGroup.Post("/register", middleware.WrapJson(h.RegisterHandler))
	authGroup.Post("/login", middleware.WrapJson(h.AuthenticateHandler))


}

// RegisterHandler godoc
// @Summary Register a new user
// @Description Registers a new user with a username, email, and password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.CreateUserDTO true "User registration payload"
// @Success 201 {object} db.User "User registered successfully (sensitive data omitted)" // Adjust response type based on actual return
// @Failure 400 {object} apierror.ApiError "Bad request (e.g., invalid input, user already exists)"
// @Failure 500 {object} apierror.Api
// @Router /auth/register [post]
func (h *authHandler) RegisterHandler(c fiber.Ctx, dto models.CreateUserDTO) error {

	user, err := h.usersService.Register(c.Context(), dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}


// AuthenticateHandler godoc
// @Summary Authenticate user and get token
// @Description Authenticates a user with email/username and password and returns a JWT upon success.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body models.AuthUserDTO true "Authentication credentials payload"
// @Success 200 {object} object{token=string} "Authentication successful, returns JWT"
// @Failure 400 {object} apierror.ApiError "Bad request (e.g., invalid input)"
// @Failure 401 {object} apierror.ApiError "Unauthorized (invalid credentials)"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /auth/login [post]
func (h *authHandler) AuthenticateHandler(c fiber.Ctx, dto models.AuthUserDTO) error {
	fmt.Println("auth")
	user, err := h.usersService.Authenticate(c.Context(), dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil{
		return apierror.CheckApiErrorAndSend(err, c)
	}
	fmt.Println("auth")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token":   token,
	})
}