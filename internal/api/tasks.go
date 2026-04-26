package api

import (
	"strconv"

	apierror "rankode/internal/errors"
	"rankode/internal/middleware"
	"rankode/internal/models"
	db "rankode/internal/repository"
	"rankode/internal/services/tasks"
	"rankode/internal/services/test_cases"

	"github.com/gofiber/fiber/v3"
)

// tasksHandler handles HTTP requests related to tasks.
type tasksHandler struct {
	service   *tasks.TaskService
	testCases *test_cases.TestCasesService
}

// NewTasksHandler creates a new instance of tasksHandler.
func NewTasksHandler(service *tasks.TaskService, testCases *test_cases.TestCasesService) *tasksHandler {
	return &tasksHandler{
		service:   service,
		testCases: testCases,
	}
}

// RegisterRoutes registers the task-related routes with the Fiber app.
func (h *tasksHandler) RegisterRoutes(app fiber.Router, authMiddleware fiber.Handler) {
	taskGroup := app.Group("/tasks").Use(authMiddleware)

	taskGroup.Post("/", middleware.WrapJson(h.CreateTaskHandler), authMiddleware, middleware.AuthRequiredMiddleware)
	taskGroup.Get("/", middleware.WrapQuery(h.ListTasksHandler))
	taskGroup.Get("/:id", h.TaskById, authMiddleware)
	taskGroup.Post("/:id/verification-file", h.UploadVerificationFileHandler, authMiddleware, middleware.AuthRequiredMiddleware)
	taskGroup.Put("/:id", middleware.WrapJson(h.UpdateTaskHandler), authMiddleware)
	taskGroup.Delete("/:id", h.DeleteTaskHandler, authMiddleware)
}

// CreateTaskHandler godoc
// @Summary Create a new task
// @Description Creates a new task with the provided details.
// @Tags Tasks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param task body models.CreateTaskDTO true "Task creation payload"
// @Success 201 {object} db.Task "Successfully cЗreated task"
// @Failure 400 {object} apierror.ApiError "Bad request (e.g., invalid input, topic not found)"
// @Failure 401 {object} apierror.ApiError "Unauthorized"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /tasks [post]
func (h *tasksHandler) CreateTaskHandler(c fiber.Ctx, dto models.CreateTaskDTO) error {
	params := db.CreateTaskParams{
		UserID:      *middleware.UserIDFromContext(c),
		Title:       dto.Title,
		Description: dto.Description,
		Difficulty:  dto.Difficulty,
		Topics:      dto.Topics,
	}

	task, err := h.service.CreateTask(c.Context(), params)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	return c.Status(fiber.StatusCreated).JSON(task)
}

// TaskById godoc
// @Summary Get task by ID
// @Description Retrieves a single task by its ID.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} models.TaskByIdResponse "Task details"
// @Failure 400 {object} apierror.ApiError "Bad request (invalid ID format)"
// @Failure 404 {object} apierror.ApiError "Task not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /tasks/{id} [get]
func (h *tasksHandler) TaskById(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID format"})
	}

	task, err := h.service.TaskById(c.Context(), int32(id))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	resp := models.TaskByIdResponse{
		Task: task,
	}
	userID := middleware.UserIDFromContext(c)
	if userID != nil && *userID == task.UserID {
		testCases, err := h.testCases.GetTestCasesByTaskID(c.Context(), int32(id))
		if err != nil {
			return apierror.CheckApiErrorAndSend(err, c)
		}
		resp.TestCases = testCases
	}

	return c.JSON(resp)
}

// ListTasksHandler godoc
// @Summary List tasks
// @Description Retrieves a list of tasks based on optional filters.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param title query string false "Filter by title"
// @Param topics query []int false "Filter by topic IDs" collectionFormat(csv)
// @Param difficulties query []int false "Filter by difficulty levels" collectionFormat(csv)
// @Param sort query string false "Sort order (e.g., name, difficulty, score)" Enums(name, difficulty, score)
// @Param limit query int false "Number of tasks to return" default(10) minimum(10) maximum(25)
// @Param offset query int false "Number of tasks to skip" default(0) minimum(0)
// @Success 200 {array} db.Task "List of tasks"
// @Failure 400 {object} apierror.ApiError "Bad request (e.g., invalid query parameters)"
// @Failure 401 {object} apierror.ApiError "Unauthorized"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /tasks [get]
func (h *tasksHandler) ListTasksHandler(c fiber.Ctx, filter db.TaskListFilter) error {
	tasks, err := h.service.ListTasksByFilter(c.Context(), filter)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(tasks)
}

// UpdateTaskHandler godoc
// @Summary Update task
// @Description Updates an existing task's details by ID.
// @Tags Tasks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Task ID"
// @Param task body db.UpdateTaskParams true "Task update payload"
// @Success 200 "Task updated successfully"
// @Failure 400 {object} apierror.ApiError "Bad request (invalid input, ID format)"
// @Failure 401 {object} apierror.ApiError "Unauthorized"
// @Failure 404 {object} apierror.ApiError "Task not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /tasks/{id} [put]
func (h *tasksHandler) UpdateTaskHandler(c fiber.Ctx, arg db.UpdateTaskParams) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID format"})
	}

	arg.ID = int32(id)

	err = h.service.UpdateTask(c.Context(), arg)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteTaskHandler godoc
// @Summary Delete task
// @Description Deletes a task by ID.
// @Tags Tasks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Task ID"
// @Success 204 "Task deleted successfully"
// @Failure 400 {object} apierror.ApiError "Bad request (invalid ID format)"
// @Failure 401 {object} apierror.ApiError "Unauthorized"
// @Failure 404 {object} apierror.ApiError "Task not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /tasks/{id} [delete]
func (h *tasksHandler) DeleteTaskHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID format"})
	}

	err = h.service.DeleteTask(c.Context(), int32(id))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UploadVerificationFileHandler godoc
// @Summary Upload verification file
// @Description Uploads a verification file for a task.
// @Tags Tasks
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Task ID"
// @Param file formData file true "Verification file"
// @Success 200 "File uploaded successfully"
// @Failure 400 {object} apierror.ApiError
// @Failure 401 {object} apierror.ApiError
// @Failure 403 {object} apierror.ApiError
// @Failure 404 {object} apierror.ApiError
// @Failure 500 {object} apierror.ApiError
// @Router /tasks/{id}/verification-file [post]
func (h *tasksHandler) UploadVerificationFileHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID format"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing file"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file"})
	}
	defer f.Close()

	userID := middleware.UserIDFromContext(c)
	err = h.service.UploadVerificationFile(c.Context(), int32(id), *userID, f, file.Size)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}

	return c.SendStatus(fiber.StatusOK)
}
