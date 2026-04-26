package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"rankode/internal/models"
	db "rankode/internal/repository"
)

func TestCoursesHandlers(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	// Register and login to get a token
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("courseuser_%d", timestamp)
	email := fmt.Sprintf("course_%d@example.com", timestamp)
	password := "password123"

	// Register
	regDto := models.CreateUserDTO{Username: username, Email: email, Password: password}
	regBody, _ := json.Marshal(regDto)
	req, _ := http.NewRequest("POST", testEndpoint("/api/auth/register"), bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	ta.App.Test(req)

	// Login
	loginDto := models.AuthUserDTO{Identifier: email, Password: password}
	loginBody, _ := json.Marshal(loginDto)
	req, _ = http.NewRequest("POST", testEndpoint("/api/auth/login"), bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := ta.App.Test(req)
	var loginResult map[string]string
	json.NewDecoder(resp.Body).Decode(&loginResult)
	token := loginResult["token"]

	var courseID int32

	t.Run("CreateCourse", func(t *testing.T) {
		dto := models.CreateCourseDTO{
			Name:        "Test Course",
			Description: "Test Description",
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", testEndpoint("/api/courses/"), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var course db.Course
		json.NewDecoder(resp.Body).Decode(&course)
		courseID = course.ID
		assert.Equal(t, dto.Name, course.Name)
	})

	t.Run("ListCourses", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testEndpoint("/api/courses/"), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var courses []db.Course
		json.NewDecoder(resp.Body).Decode(&courses)
		assert.GreaterOrEqual(t, len(courses), 1)
	})

	t.Run("GetCourseByID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testEndpoint(fmt.Sprintf("/api/courses/%d", courseID)), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UpdateCourse", func(t *testing.T) {
		dto := models.UpdateCourseDTO{
			Name:        "Updated Course",
			Description: "Updated Description",
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("PUT", testEndpoint(fmt.Sprintf("/api/courses/%d", courseID)), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DeleteCourse", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", testEndpoint(fmt.Sprintf("/api/courses/%d", courseID)), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
