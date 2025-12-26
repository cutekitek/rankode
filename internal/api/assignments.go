package api

import (
	"strconv"

	apierror "rankode/internal/errors"
	"rankode/internal/middleware"
	"rankode/internal/models"
	"rankode/internal/services/assignments"

	"github.com/gofiber/fiber/v3"
)

type assignmentsHandler struct {
	service *assignments.AssignmentService
}

func NewAssignmentsHandler(service *assignments.AssignmentService) *assignmentsHandler {
	return &assignmentsHandler{
		service: service,
	}
}

func (h *assignmentsHandler) RegisterRoutes(app fiber.Router, authMiddleware fiber.Handler) {
	assignmentGroup := app.Group("/assignments")
	assignmentGroup.Use(authMiddleware)

	assignmentGroup.Post("/", middleware.WrapJson(h.CreateAssignmentHandler))
	assignmentGroup.Get("/", h.ListAssignmentsHandler)
	assignmentGroup.Get("/:id", h.GetAssignmentHandler)
	assignmentGroup.Put("/:id", middleware.WrapJson(h.UpdateAssignmentHandler))
	assignmentGroup.Delete("/:id", h.DeleteAssignmentHandler)
	assignmentGroup.Post("/:id/tasks", middleware.WrapJson(h.AddTaskToAssignmentHandler))
	assignmentGroup.Delete("/:id/tasks/:taskId", h.RemoveTaskFromAssignmentHandler)
	assignmentGroup.Get("/:id/stats", h.GetAssignmentStatsHandler)
}

// CreateAssignmentHandler godoc
// @Summary Create a new assignment
// @Description Creates a new assignment in a course (teacher only)
// @Tags Assignments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param assignment body models.CreateAssignmentDTO true "Assignment creation payload"
// @Success 201 {object} db.Assignment "Successfully created assignment"
// @Failure 400 {object} apierror.ApiError "Bad request"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /assignments [post]
func (h *assignmentsHandler) CreateAssignmentHandler(c fiber.Ctx, dto models.CreateAssignmentDTO) error {
	teacherID := *middleware.UserIDFromContext(c)
	assignment, err := h.service.CreateAssignment(c.Context(), teacherID, dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.Status(fiber.StatusCreated).JSON(assignment)
}

// GetAssignmentHandler godoc
// @Summary Get assignment by ID with tasks
// @Description Retrieves a single assignment by its ID with included tasks
// @Tags Assignments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Assignment ID"
// @Success 200 {object} models.AssignmentResponse "Assignment details"
// @Failure 404 {object} apierror.ApiError "Assignment not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /assignments/{id} [get]
func (h *assignmentsHandler) GetAssignmentHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID format"})
	}
	assignment, err := h.service.GetAssignmentWithTasks(c.Context(), int32(id))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(assignment)
}

// ListAssignmentsHandler godoc
// @Summary List assignments
// @Description Retrieves assignments for the current user (as teacher or student)
// @Tags Assignments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param course_id query int false "Filter by course ID"
// @Success 200 {array} db.Assignment "List of assignments"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /assignments [get]
func (h *assignmentsHandler) ListAssignmentsHandler(c fiber.Ctx) error {
	userID := *middleware.UserIDFromContext(c)
	courseIDStr := c.Query("course_id")
	if courseIDStr != "" {
		courseID, err := strconv.Atoi(courseIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid course ID format"})
		}
		assignments, err := h.service.ListAssignmentsByCourse(c.Context(), int32(courseID))
		if err != nil {
			return apierror.CheckApiErrorAndSend(err, c)
		}
		return c.JSON(assignments)
	}
	// List assignments for student (enrolled courses)
	assignments, err := h.service.ListAssignmentsForStudent(c.Context(), userID)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(assignments)
}

// UpdateAssignmentHandler godoc
// @Summary Update assignment
// @Description Updates an existing assignment (teacher only)
// @Tags Assignments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Assignment ID"
// @Param assignment body models.UpdateAssignmentDTO true "Assignment update payload"
// @Success 200 "Assignment updated successfully"
// @Failure 400 {object} apierror.ApiError "Bad request"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Assignment not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /assignments/{id} [put]
func (h *assignmentsHandler) UpdateAssignmentHandler(c fiber.Ctx, dto models.UpdateAssignmentDTO) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.UpdateAssignment(c.Context(), int32(id), teacherID, dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusOK)
}

// DeleteAssignmentHandler godoc
// @Summary Delete assignment
// @Description Deletes an assignment (teacher only)
// @Tags Assignments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Assignment ID"
// @Success 204 "Assignment deleted successfully"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Assignment not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /assignments/{id} [delete]
func (h *assignmentsHandler) DeleteAssignmentHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.DeleteAssignment(c.Context(), int32(id), teacherID)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// AddTaskToAssignmentHandler godoc
// @Summary Add task to assignment
// @Description Adds an existing task to an assignment (teacher only)
// @Tags Assignments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Assignment ID"
// @Param task body models.AddTaskToAssignmentDTO true "Task addition payload"
// @Success 200 "Task added to assignment successfully"
// @Failure 400 {object} apierror.ApiError "Bad request (task not in course, already added)"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Assignment or task not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /assignments/{id}/tasks [post]
func (h *assignmentsHandler) AddTaskToAssignmentHandler(c fiber.Ctx, dto models.AddTaskToAssignmentDTO) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.AddTaskToAssignment(c.Context(), int32(id), teacherID, dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusOK)
}

// RemoveTaskFromAssignmentHandler godoc
// @Summary Remove task from assignment
// @Description Removes a task from an assignment (teacher only)
// @Tags Assignments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Assignment ID"
// @Param taskId path int true "Task ID"
// @Success 200 "Task removed from assignment successfully"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Assignment or task not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /assignments/{id}/tasks/{taskId} [delete]
func (h *assignmentsHandler) RemoveTaskFromAssignmentHandler(c fiber.Ctx) error {
	assignmentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID format"})
	}
	taskID, err := strconv.Atoi(c.Params("taskId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.RemoveTaskFromAssignment(c.Context(), int32(assignmentID), int32(taskID), teacherID)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusOK)
}

// GetAssignmentStatsHandler godoc
// @Summary Get assignment statistics
// @Description Retrieves statistics for an assignment (teacher only)
// @Tags Assignments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Assignment ID"
// @Success 200 {object} models.AssignmentStats "Assignment statistics"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Assignment not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /assignments/{id}/stats [get]
func (h *assignmentsHandler) GetAssignmentStatsHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID format"})
	}
	stats, err := h.service.GetAssignmentStats(c.Context(), int32(id))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(stats)
}
