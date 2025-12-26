package api

import (
	"strconv"

	apierror "rankode/internal/errors"
	"rankode/internal/middleware"
	"rankode/internal/models"
	"rankode/internal/services/courses"

	"github.com/gofiber/fiber/v3"
)

type coursesHandler struct {
	service *courses.CourseService
}

func NewCoursesHandler(service *courses.CourseService) *coursesHandler {
	return &coursesHandler{
		service: service,
	}
}

func (h *coursesHandler) RegisterRoutes(app fiber.Router, authMiddleware fiber.Handler) {
	courseGroup := app.Group("/courses")
	courseGroup.Use(authMiddleware)

	courseGroup.Post("/", middleware.WrapJson(h.CreateCourseHandler))
	courseGroup.Get("/", h.ListCoursesHandler)
	courseGroup.Get("/:id", h.GetCourseHandler)
	courseGroup.Put("/:id", middleware.WrapJson(h.UpdateCourseHandler))
	courseGroup.Delete("/:id", h.DeleteCourseHandler)
	courseGroup.Post("/enroll", middleware.WrapJson(h.EnrollByJoinCodeHandler))
	courseGroup.Post("/:id/enroll", h.EnrollInCourseHandler)
	courseGroup.Delete("/:id/enroll", h.UnenrollFromCourseHandler)
	courseGroup.Get("/:id/students", h.ListStudentsHandler)
}

// CreateCourseHandler godoc
// @Summary Create a new course
// @Description Creates a new course (teacher only)
// @Tags Courses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param course body models.CreateCourseDTO true "Course creation payload"
// @Success 201 {object} db.Course "Successfully created course"
// @Failure 400 {object} apierror.ApiError "Bad request"
// @Failure 403 {object} apierror.ApiError "Forbidden (not a teacher)"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /courses [post]
func (h *coursesHandler) CreateCourseHandler(c fiber.Ctx, dto models.CreateCourseDTO) error {
	teacherID := *middleware.UserIDFromContext(c)
	course, err := h.service.CreateCourse(c.Context(), teacherID, dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.Status(fiber.StatusCreated).JSON(course)
}

// GetCourseHandler godoc
// @Summary Get course by ID
// @Description Retrieves a single course by its ID
// @Tags Courses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Course ID"
// @Success 200 {object} models.CourseResponse "Course details"
// @Failure 404 {object} apierror.ApiError "Course not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /courses/{id} [get]
func (h *coursesHandler) GetCourseHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid course ID format"})
	}

	course, err := h.service.GetCourseWithStats(c.Context(), int32(id))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(course)
}

// ListCoursesHandler godoc
// @Summary List courses
// @Description Retrieves courses for the current user (as teacher or student)
// @Tags Courses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} db.Course "List of courses"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /courses [get]
func (h *coursesHandler) ListCoursesHandler(c fiber.Ctx) error {
	userID := *middleware.UserIDFromContext(c)
	// Determine role? For now list both teaching and enrolled courses
	courses, err := h.service.ListCoursesByTeacher(c.Context(), userID)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	// Also get enrolled courses
	enrolled, err := h.service.ListCoursesForStudent(c.Context(), userID)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	// Merge unique courses (simple approach)
	allCourses := append(courses, enrolled...)
	return c.JSON(allCourses)
}

// UpdateCourseHandler godoc
// @Summary Update course
// @Description Updates an existing course (teacher only)
// @Tags Courses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Course ID"
// @Param course body models.UpdateCourseDTO true "Course update payload"
// @Success 200 "Course updated successfully"
// @Failure 400 {object} apierror.ApiError "Bad request"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Course not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /courses/{id} [put]
func (h *coursesHandler) UpdateCourseHandler(c fiber.Ctx, dto models.UpdateCourseDTO) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid course ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.UpdateCourse(c.Context(), int32(id), teacherID, dto)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusOK)
}

// DeleteCourseHandler godoc
// @Summary Delete course
// @Description Deletes a course (teacher only)
// @Tags Courses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Course ID"
// @Success 204 "Course deleted successfully"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Course not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /courses/{id} [delete]
func (h *coursesHandler) DeleteCourseHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid course ID format"})
	}
	teacherID := *middleware.UserIDFromContext(c)
	err = h.service.DeleteCourse(c.Context(), int32(id), teacherID)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// EnrollByJoinCodeHandler godoc
// @Summary Enroll in a course using join code
// @Description Enrolls the current user in a course using join code
// @Tags Courses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body models.JoinCourseDTO true "Join code"
// @Success 200 {object} db.Course "Successfully enrolled"
// @Failure 400 {object} apierror.ApiError "Already enrolled or invalid join code"
// @Failure 404 {object} apierror.ApiError "Course not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /courses/enroll [post]
func (h *coursesHandler) EnrollByJoinCodeHandler(c fiber.Ctx, dto models.JoinCourseDTO) error {
	userID := *middleware.UserIDFromContext(c)
	course, err := h.service.EnrollByJoinCode(c.Context(), userID, dto.JoinCode)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(course)
}

// EnrollInCourseHandler godoc
// @Summary Enroll in a course (direct)
// @Description Enrolls the current user in a course (requires teacher permission?)
// @Tags Courses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Course ID"
// @Success 200 "Successfully enrolled"
// @Failure 400 {object} apierror.ApiError "Already enrolled"
// @Failure 403 {object} apierror.ApiError "Forbidden"
// @Failure 404 {object} apierror.ApiError "Course not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /courses/{id}/enroll [post]
func (h *coursesHandler) EnrollInCourseHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid course ID format"})
	}
	userID := *middleware.UserIDFromContext(c)
	err = h.service.EnrollStudent(c.Context(), int32(id), userID, "student")
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusOK)
}

// UnenrollFromCourseHandler godoc
// @Summary Unenroll from a course
// @Description Removes the current user from a course
// @Tags Courses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Course ID"
// @Success 200 "Successfully unenrolled"
// @Failure 400 {object} apierror.ApiError "Not enrolled"
// @Failure 404 {object} apierror.ApiError "Course not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /courses/{id}/enroll [delete]
func (h *coursesHandler) UnenrollFromCourseHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid course ID format"})
	}
	userID := *middleware.UserIDFromContext(c)
	err = h.service.UnenrollStudent(c.Context(), int32(id), userID)
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.SendStatus(fiber.StatusOK)
}

// ListStudentsHandler godoc
// @Summary List students in course
// @Description Lists all students enrolled in a course (teacher only)
// @Tags Courses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Course ID"
// @Success 200 {array} models.StudentInCourse "List of students"
// @Failure 403 {object} apierror.ApiError "Forbidden (not teacher of course)"
// @Failure 404 {object} apierror.ApiError "Course not found"
// @Failure 500 {object} apierror.ApiError "Internal server error"
// @Router /courses/{id}/students [get]
func (h *coursesHandler) ListStudentsHandler(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid course ID format"})
	}
	students, err := h.service.ListStudentsInCourse(c.Context(), int32(id))
	if err != nil {
		return apierror.CheckApiErrorAndSend(err, c)
	}
	return c.JSON(students)
}
