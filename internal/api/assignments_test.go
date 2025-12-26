package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"rankode/internal/models"
	db "rankode/internal/repository"
)

func TestAssignmentsHandlers(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	// Register and login
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("assignuser_%d", timestamp)
	email := fmt.Sprintf("assign_%d@example.com", timestamp)
	password := "password123"

	regDto := models.CreateUserDTO{Username: username, Email: email, Password: password}
	regBody, _ := json.Marshal(regDto)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	ta.App.Test(req)

	loginDto := models.AuthUserDTO{Identifier: email, Password: password}
	loginBody, _ := json.Marshal(loginDto)
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := ta.App.Test(req)
	var loginResult map[string]string
	json.NewDecoder(resp.Body).Decode(&loginResult)
	token := loginResult["token"]

	// Create a course
	courseDto := models.CreateCourseDTO{Name: "Test Course"}
	courseBody, _ := json.Marshal(courseDto)
	req, _ = http.NewRequest("POST", "/api/courses/", bytes.NewBuffer(courseBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = ta.App.Test(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var course db.Course
	json.NewDecoder(resp.Body).Decode(&course)
	courseID := course.ID

	var assignmentID int32
	maxAttempts := int32(10)

	t.Run("CreateAssignment", func(t *testing.T) {
		dto := models.CreateAssignmentDTO{
			CourseID:           courseID,
			Title:              "Test Assignment",
			Description:        "Test Description",
			MaxAttemptsPerTask: &maxAttempts,
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/api/assignments/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Response body: %s", string(body))
		}
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var assignment db.Assignment
		json.NewDecoder(resp.Body).Decode(&assignment)
		assignmentID = assignment.ID
	})

	t.Run("ListAssignments", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/assignments/?course_id=%d", courseID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetAssignmentByID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/assignments/%d", assignmentID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UpdateAssignment", func(t *testing.T) {
		dto := models.UpdateAssignmentDTO{
			Title:              "Updated Assignment",
			Description:        "Updated Description",
			MaxAttemptsPerTask: &maxAttempts,
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/assignments/%d", assignmentID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DeleteAssignment", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/assignments/%d", assignmentID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
