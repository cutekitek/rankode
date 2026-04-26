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

func TestGroupsHandlers(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	// Register and login
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("groupuser_%d", timestamp)
	email := fmt.Sprintf("group_%d@example.com", timestamp)
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

	// Create a course
	courseDto := models.CreateCourseDTO{Name: "Test Course"}
	courseBody, _ := json.Marshal(courseDto)
	req, _ = http.NewRequest("POST", testEndpoint("/api/courses/"), bytes.NewBuffer(courseBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = ta.App.Test(req)
	var course db.Course
	json.NewDecoder(resp.Body).Decode(&course)
	courseID := course.ID

	var groupID int32

	t.Run("CreateGroup", func(t *testing.T) {
		dto := models.CreateGroupDTO{
			CourseID: courseID,
			Name:     "Test Group",
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", testEndpoint("/api/groups/"), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var group db.Group
		json.NewDecoder(resp.Body).Decode(&group)
		groupID = group.ID
	})

	t.Run("ListGroups", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testEndpoint(fmt.Sprintf("/api/groups/?course_id=%d", courseID)), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DeleteGroup", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", testEndpoint(fmt.Sprintf("/api/groups/%d", groupID)), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
