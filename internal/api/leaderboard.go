package api

import (
	"rankode/internal/errors"
	"rankode/internal/services/users"

	"github.com/gofiber/fiber/v3"
)

type leaderboardHandler struct {
	usersService *users.UserService
}

func NewLeaderboardHandler(users *users.UserService) *leaderboardHandler {
	return &leaderboardHandler{
		usersService: users,
	}
}

func (h *leaderboardHandler) RegisterRoutes(app fiber.Router) {
	app.Get("/leaderboard", h.handler)
}

// Leaderboard godoc
// @Summary Users leaderboard
// @Description Users leaderboard
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {array} db.GetUsersLeaderboardRow "List of users"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /leaderboard [get]
func (h *leaderboardHandler) handler(c fiber.Ctx) error {
	users, err := h.usersService.GetLeaderboard(c.Context())
	if err != nil {
		apierror.CheckApiErrorAndSend(err, c)
	}
	return c.Status(fiber.StatusOK).JSON(users)
}
