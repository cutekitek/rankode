package api

import (
	apierror "rankode/internal/errors"
	"rankode/internal/middleware"
	"rankode/internal/models"
	"rankode/internal/services/tasks"

	"github.com/gofiber/fiber/v3"
)

type topicsHandler struct {
	service *tasks.TaskService
}

func NewtopicsHandler(service *tasks.TaskService) *topicsHandler {
	return &topicsHandler{service: service}
}

func (h *topicsHandler) RegisterRoutes(app fiber.Router, authMiddleware fiber.Handler) {
	taskGroup := app.Group("/topics")

	taskGroup.Get("/", middleware.WrapQuery(h.ListTopicsHandler))
	taskGroup.Post("/", middleware.WrapJson(h.AddTopicHandler), authMiddleware, middleware.AuthRequiredMiddleware)
}

// ListTopics godoc
// @Summary List tasks
// @Description Retrieves a list of tasks based on optional filters.
// @Tags Topics
// @Accept json
// @Produce json
// @Success 200 {array} db.Topic "List of topics"
// @Failure 400 {object} apierror.ApiError "Bad request (e.g., invalid query parameters)"
// @Failure 401 {object} apierror.ApiError "Unauthorized"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /topics [get]
func (h *topicsHandler) ListTopicsHandler(c fiber.Ctx, dto models.ListTopicsDTO) error {
	topics, err := h.service.ListTopics(c.Context(), dto)

	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.Status(fiber.StatusOK).JSON(topics)
}

// AddTopic godoc
// @Summary Add Topics
// @Description Adds new topic by its alias
// @Tags Topics
// @Accept json
// @Produce json
// @Param topic body models.AddTopicDTO true "Topic payload"
// @Success 200 {object} db.Topic "Created Topic"
// @Failure 400 {object} apierror.ApiError "Bad request (e.g., invalid query parameters)"
// @Failure 401 {object} apierror.ApiError "Unauthorized"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /topics [post]
func (h *topicsHandler) AddTopicHandler(c fiber.Ctx, dto models.AddTopicDTO) error {
	topic, err := h.service.AddTopic(c.Context(), dto.Name)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.Status(fiber.StatusOK).JSON(topic)
}
