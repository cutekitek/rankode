package api

import (
	"strconv"

	apierror "rankode/internal/errors"
	"rankode/internal/middleware"
	"rankode/internal/models"
	"rankode/internal/services/groups"

	"github.com/gofiber/fiber/v3"
)

type groupsHandler struct {
	service *groups.GroupService
}

func NewGroupsHandler(service *groups.GroupService) *groupsHandler {
	return &groupsHandler{
		service: service,
	}
}

func (h *groupsHandler) RegisterRoutes(app fiber.Router, authMiddleware fiber.Handler) {
	groupGroup := app.Group("/groups")
	groupGroup.Use(authMiddleware)

	groupGroup.Post("/", middleware.WrapJson(h.CreateGroupHandler))
	groupGroup.Get("/", h.ListGroupsHandler)
	groupGroup.Delete("/:id", h.DeleteGroupHandler)
	groupGroup.Post("/:id/students", middleware.WrapJson(h.AddStudentToGroupHandler))
	groupGroup.Delete("/:id/students/:studentId", h.RemoveStudentFromGroupHandler)
	groupGroup.Get("/:id/students", h.ListStudentsHandler)
}

func (h *groupsHandler) CreateGroupHandler(c fiber.Ctx, dto models.CreateGroupDTO) error {
	teacherID := *middleware.UserIDFromContext(c)
	group, err := h.service.CreateGroup(c.Context(), teacherID, dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.Status(fiber.StatusCreated).JSON(group)
}

func (h *groupsHandler) ListGroupsHandler(c fiber.Ctx) error {
	courseIDStr := c.Query("course_id")
	if courseIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "course_id query parameter required"})
	}
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid course ID format"})
	}
	groups, err := h.service.ListGroupsByCourse(c.Context(), int32(courseID))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(groups)
}

func (h *groupsHandler) DeleteGroupHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.DeleteGroup(c.Context(), teacherID, int32(id))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *groupsHandler) AddStudentToGroupHandler(c fiber.Ctx, dto models.AddStudentToGroupDTO) error {
	groupID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.AddStudentToGroup(c.Context(), teacherID, int32(groupID), dto.UserID)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *groupsHandler) RemoveStudentFromGroupHandler(c fiber.Ctx) error {
	groupID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID format"})
	}
	studentID, err := strconv.Atoi(c.Params("studentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.RemoveStudentFromGroup(c.Context(), teacherID, int32(groupID), int32(studentID))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *groupsHandler) ListStudentsHandler(c fiber.Ctx) error {
	groupID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID format"})
	}
	students, err := h.service.ListStudentsInGroup(c.Context(), int32(groupID))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(students)
}
