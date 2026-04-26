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

func TestGradesHandlers(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	// Register and login
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("gradeuser_%d", timestamp)
	email := fmt.Sprintf("grade_%d@example.com", timestamp)
	password := "password123"

	regDto := models.CreateUserDTO{Username: username, Email: email, Password: password}
	regBody, _ := json.Marshal(regDto)
	req, _ := http.NewRequest("POST", testEndpoint("/api/auth/register"), bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	ta.App.Test(req)

	loginDto := models.AuthUserDTO{Identifier: email, Password: password}
	loginBody, _ := json.Marshal(loginDto)
	req, _ = http.NewRequest("POST", testEndpoint("/api/auth/login"), bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := ta.App.Test(req)
	var loginResult map[string]string
	json.NewDecoder(resp.Body).Decode(&loginResult)
	token := loginResult["token"]

	id, _ := ta.AuthService.VerifyToken(token)

	// Create a course
	courseDto := models.CreateCourseDTO{Name: "Test Course"}
	courseBody, _ := json.Marshal(courseDto)
	req, _ = http.NewRequest("POST", testEndpoint("/api/courses/"), bytes.NewBuffer(courseBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = ta.App.Test(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var course db.Course
	json.NewDecoder(resp.Body).Decode(&course)

	// Enroll as student
	req, _ = http.NewRequest("POST", testEndpoint(fmt.Sprintf("/api/courses/%d/enroll", course.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = ta.App.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Create a task
	taskDto := models.CreateTaskDTO{Title: "Test Task"}
	taskBody, _ := json.Marshal(taskDto)
	req, _ = http.NewRequest("POST", testEndpoint("/api/tasks/"), bytes.NewBuffer(taskBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = ta.App.Test(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var task db.Task
	json.NewDecoder(resp.Body).Decode(&task)

	// Create an assignment
	maxAttempts := int32(10)
	assignDto := models.CreateAssignmentDTO{CourseID: course.ID, Title: "Test Assignment", MaxAttemptsPerTask: &maxAttempts}
	assignBody, _ := json.Marshal(assignDto)
	req, _ = http.NewRequest("POST", testEndpoint("/api/assignments/"), bytes.NewBuffer(assignBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = ta.App.Test(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var assign db.Assignment
	json.NewDecoder(resp.Body).Decode(&assign)

	// Add task to assignment
	addTaskDto := models.AddTaskToAssignmentDTO{TaskID: task.ID, OrderIndex: 1, Weight: 1.0}
	addTaskBody, _ := json.Marshal(addTaskDto)
	req, _ = http.NewRequest("POST", testEndpoint(fmt.Sprintf("/api/assignments/%d/tasks", assign.ID)), bytes.NewBuffer(addTaskBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = ta.App.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	t.Run("CreateOrUpdateGrade", func(t *testing.T) {
		dto := models.CreateGradeDTO{
			AssignmentID: assign.ID,
			TaskID:       task.ID,
			UserID:       id,
			Grade:        5,
			Feedback:     "Excellent",
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", testEndpoint("/api/grades/"), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Response body: %s", string(body))
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var grade db.Grade
		json.NewDecoder(resp.Body).Decode(&grade)
		assert.Equal(t, int16(5), grade.Grade)
	})

	t.Run("ListGrades", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testEndpoint(fmt.Sprintf("/api/grades/?assignment_id=%d", assign.ID)), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
