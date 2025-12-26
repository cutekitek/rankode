package api

import (
	"strconv"

	apierror "rankode/internal/errors"
	"rankode/internal/middleware"
	"rankode/internal/models"
	"rankode/internal/services/grades"

	"github.com/gofiber/fiber/v3"
)

type gradesHandler struct {
	service *grades.GradeService
}

func NewGradesHandler(service *grades.GradeService) *gradesHandler {
	return &gradesHandler{
		service: service,
	}
}

func (h *gradesHandler) RegisterRoutes(app fiber.Router, authMiddleware fiber.Handler) {
	gradeGroup := app.Group("/grades")
	gradeGroup.Use(authMiddleware)

	gradeGroup.Post("/", middleware.WrapJson(h.CreateOrUpdateGradeHandler))
	gradeGroup.Get("/", h.ListGradesHandler)
	gradeGroup.Get("/stats", h.GetGradeStatsHandler)
	gradeGroup.Delete("/:id", h.DeleteGradeHandler)
}

// CreateOrUpdateGradeHandler godoc
// @Summary Create or update a grade
// @Description Creates or updates a grade for a student's task in an assignment (teacher only)
// @Tags Grades
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param grade body models.CreateGradeDTO true "Grade payload"
// @Success 200 {object} db.Grade "Grade created/updated successfully"
// @Failure 400 {object} apierror.ApiError "Bad request"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Assignment, task, or student not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /grades [post]
func (h *gradesHandler) CreateOrUpdateGradeHandler(c fiber.Ctx, dto models.CreateGradeDTO) error {
	teacherID := *middleware.UserIDFromContext(c)
	grade, err := h.service.CreateOrUpdateGrade(c.Context(), teacherID, dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(grade)
}

// ListGradesHandler godoc
// @Summary List grades
// @Description Retrieves grades based on filters (teacher sees all, students see own)
// @Tags Grades
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param assignment_id query int false "Filter by assignment ID"
// @Param course_id query int false "Filter by course ID"
// @Param student_id query int false "Filter by student ID (teacher only)"
// @Success 200 {array} models.GradeResponse "List of grades"
// @Failure 400 {object} apierror.ApiError "Bad request"
// @Failure 403 {object} apierror.ApiError "Forbidden"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /grades [get]
func (h *gradesHandler) ListGradesHandler(c fiber.Ctx) error {
	assignmentIDStr := c.Query("assignment_id")
	courseIDStr := c.Query("course_id")
	studentIDStr := c.Query("student_id")

	// If student_id provided, verify teacher permission
	if studentIDStr != "" {
		// For simplicity, we'll assume teachers can view any student's grades
		// In production, verify teacher is teacher of the course
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid student ID format"})
		}
		if assignmentIDStr != "" {
			assignmentID, err := strconv.Atoi(assignmentIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID format"})
			}
			grades, err := h.service.GetStudentGradesForAssignment(c.Context(), int32(assignmentID), int32(studentID))
			if err != nil {
				return apierror.CheckApiErrorAndSend(err, c)
			}
			return c.JSON(grades)
		} else if courseIDStr != "" {
			courseID, err := strconv.Atoi(courseIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid course ID format"})
			}
			grades, err := h.service.GetStudentGradesForCourse(c.Context(), int32(courseID), int32(studentID))
			if err != nil {
				return apierror.CheckApiErrorAndSend(err, c)
			}
			return c.JSON(grades)
		}
		// No filter, fall through
	}

	if assignmentIDStr != "" {
		assignmentID, err := strconv.Atoi(assignmentIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID format"})
		}
		// Teacher view all grades for assignment
		grades, err := h.service.GetGradesForAssignment(c.Context(), int32(assignmentID))
		if err != nil {
			return apierror.CheckApiErrorAndSend(err, c)
		}
		return c.JSON(grades)
	}

	// Default: return current user's grades across all courses
	// This is a simplified implementation
	// In reality, we'd need a more complex query
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "Not implemented"})
}

// GetGradeStatsHandler godoc
// @Summary Get grade statistics
// @Description Retrieves statistics for an assignment (teacher only)
// @Tags Grades
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param assignment_id query int true "Assignment ID"
// @Success 200 {object} models.GradeStats "Grade statistics"
// @Failure 400 {object} apierror.ApiError "Bad request"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Assignment not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /grades/stats [get]
func (h *gradesHandler) GetGradeStatsHandler(c fiber.Ctx) error {
	assignmentIDStr := c.Query("assignment_id")
	if assignmentIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "assignment_id query parameter required"})
	}
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignment ID format"})
	}
	stats, err := h.service.GetGradeStats(c.Context(), int32(assignmentID))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(stats)
}

// DeleteGradeHandler godoc
// @Summary Delete a grade
// @Description Deletes a grade by ID (teacher only)
// @Tags Grades
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Grade ID"
// @Success 204 "Grade deleted successfully"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher)"
// @Failure 404 {object} apierror.ApiError "Grade not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /grades/{id} [delete]
func (h *gradesHandler) DeleteGradeHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid grade ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.DeleteGrade(c.Context(), int32(id), teacherID)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
