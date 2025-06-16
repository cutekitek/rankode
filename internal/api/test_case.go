package api

import (
	"strconv"

	errors "rankode/internal/errors"
	"rankode/internal/middleware"
	"rankode/internal/models"
	"rankode/internal/services/test_cases"

	"github.com/gofiber/fiber/v3"
)

// testCasesHandler handles HTTP requests related to test cases.
type testCasesHandler struct {
	service *test_cases.TestCasesService
}

// NewTestCasesHandler creates a new instance of testCasesHandler.
func NewTestCasesHandler(service *test_cases.TestCasesService) *testCasesHandler {
	return &testCasesHandler{service: service}
} 

// RegisterRoutes registers the test case-related routes with the Fiber app.
func (h *testCasesHandler) RegisterRoutes(app fiber.Router, auth fiber.Handler) {
	group := app.Group("/test-cases", auth).Use(middleware.AuthRequiredMiddleware)

	group.Post("/", middleware.WrapJson(h.NewTestCaseHandler))
	group.Post("/:id/upload", middleware.WrapQuery(h.UploadTestCaseFileHandler))
}

// NewTestCaseHandler godoc
// @Summary Create a new test case
// @Description Creates a new test case for a task
// @Tags TestCases
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param testCase body models.NewTestCaseReq true "Test case payload"
// @Success 201 {object} db.TaskTestCase "Successfully created test case"
// @Failure 400 {object} apierror.ApiError
// @Failure 401 {object} apierror.ApiError
// @Failure 404 {object} apierror.ApiError
// @Failure 500 {object} apierror.ApiError
// @Router /test-cases [post]
func (h *testCasesHandler) NewTestCaseHandler(c fiber.Ctx, req models.NewTestCaseReq) error {
	userID := middleware.UserIDFromContext(c)
	testCase, err := h.service.NewTestCase(c.Context(), int(*userID), req)
	if err != nil {
		return errors.CheckApiErrorAndSend(err, c)
	}
	return c.Status(fiber.StatusCreated).JSON(testCase)
}

// UploadTestCaseFileHandler godoc
// @Summary Upload test case file
// @Description Uploads input/output file for a test case
// @Tags TestCases
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Test Case ID"
// @Param type query string true "Upload type" 
// @Param file formData file true "Test case file"
// @Success 200 "File uploaded successfully"
// @Failure 400 {object} apierror.ApiError
// @Failure 401 {object} apierror.ApiError
// @Failure 404 {object} apierror.ApiError
// @Failure 500 {object} apierror.ApiError
// @Router /test-cases/{id}/upload [post]
func (h *testCasesHandler) UploadTestCaseFileHandler(c fiber.Ctx, req models.UploadTestCaseReq) error {
	testCaseID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid test case ID"})
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
	params := models.UploadTestCaseFileParams{
		UserID:     int(*userID),
		TestCaseID: int32(testCaseID),
		FileSize:   file.Size,
		Type: req.Type,
		Reader:     f,
	}
	err = h.service.UploadTestCaseFile(c.Context(), params)
	if err != nil {
		return errors.CheckApiErrorAndSend(err, c)
	}

	return c.SendStatus(fiber.StatusOK)
}
