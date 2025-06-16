package api

import (
	"strconv"

	apierror "rankode/internal/errors"
	"rankode/internal/middleware"
	"rankode/internal/models"
	db "rankode/internal/repository"
	"rankode/internal/services/attempts"

	"github.com/gofiber/fiber/v3"
)

// attemptsHandler handles HTTP requests related to attempts.
type attemptsHandler struct {
	service *attempts.AttemptsService
}

// NewAttemptsHandler creates a new instance of attemptsHandler.
func NewAttemptsHandler(service *attempts.AttemptsService) *attemptsHandler {
	return &attemptsHandler{
		service: service,
	}
}

func (h *attemptsHandler) RegisterRoutes(app fiber.Router, authMiddleware fiber.Handler) {
	attemptsGroup := app.Group("/attempts")

	attemptsGroup.Post(
		"/",
		middleware.WrapJson(h.CreateAttemptHandler),
		authMiddleware,
	)

	attemptsGroup.Get(
		"/",
		h.GetUserTaskAttemptsHandler,
		authMiddleware,
	)
}

// CreateAttemptHandler godoc
// @Summary Create a new attempt
// @Description Submits a new code attempt for execution against a task’s test cases.
// @Tags Attempts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param attempt body models.CreateAttemptRequest true "Attempt creation payload"
// @Success 201 "Attempt queued successfully"
// @Failure 400 {object} apierror.ApiError "Bad request (e.g., still running or invalid input)"
// @Failure 401 {object} apierror.ApiError "Unauthorized"
// @Failure 404 {object} apierror.ApiError "Task not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /attempts [post]
func (h *attemptsHandler) CreateAttemptHandler(c fiber.Ctx, dto models.CreateAttemptRequest) error {
	rawUserID := middleware.UserIDFromContext(c)

	user := db.User{ID: *rawUserID}

	err := h.service.NewAttempt(c.Context(), user, dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	// If everything succeeded, return HTTP 201 (Created) with no body.
	return c.SendStatus(fiber.StatusCreated)
}

// GetUserTaskAttemptsHandler godoc
// @Summary List user’s attempts for a specific task
// @Description Retrieves all attempts submitted by the authenticated user for a given task.
// @Tags Attempts
// @Accept json
// @Produce json
// @Security ApiKeyAuoth
// @Param taskId query int true "Task ID"
// @Success 200 {array} models.GetUserTaskAttemptsResponse "List of attempts"
// @Failure 400 {object} apierror.ApiError "Bad request (e.g., missing or invalid taskId)"
// @Failure 401 {object} apierror.ApiError "Unauthorized"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /attempts [get]
func (h *attemptsHandler) GetUserTaskAttemptsHandler(c fiber.Ctx) error {
	taskIDStr := c.Query("taskId", "")
	if taskIDStr == "" {
		return c.
			Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Missing required query parameter: taskId"})
	}
	taskIDInt, err := strconv.Atoi(taskIDStr)
	if err != nil {
		return c.
			Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Invalid taskId; must be an integer"})
	}
	taskID := int32(taskIDInt)

	rawUserID := middleware.UserIDFromContext(c)
	user := db.User{ID: *rawUserID}

	attempts, err := h.service.GetUserTaskAttempts(c.Context(), user, taskID)

	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	response := make([]models.GetUserTaskAttemptsResponse, 0, len(attempts))
	for _, a := range attempts {
		response = append(response, 
		models.GetUserTaskAttemptsResponse{
			Id:          int(a.ID),
			Code:        a.Code,
			Language:    "",
			Status:      int(a.AttemptStatus),
			UpdatedAt:   a.UpdatedAt.Time,
			RunningTime: int(a.RunningTime.Int32),
			MemoryUsage: int(a.Memory.Int32),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
