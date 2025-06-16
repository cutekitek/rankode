package middleware

import (
	"rankode/internal/services/auth"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func NewAuthMiddleware(authService *auth.AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := strings.Split(c.Get("Authorization"), " ")
		if len(header) != 2 {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		userId, err := authService.VerifyToken(header[1])
		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		c.Locals("user_id", userId)
		return c.Next()
	}
}

func AuthRequiredMiddleware(c fiber.Ctx) error {
	userId := c.Locals("user_id")
	if userId == nil {
		c.SendStatus(fiber.StatusUnauthorized)
	}
	return c.Next()
}

func UserIDFromContext(c fiber.Ctx) *int32 {
	id, ok := c.Locals("user_id").(int32)
	if !ok {
		return nil
	}
	return &id
}
