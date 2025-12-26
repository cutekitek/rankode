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

func TestTasksHandlers(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	// Register and login to get a token
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("taskuser_%d", timestamp)
	email := fmt.Sprintf("task_%d@example.com", timestamp)
	password := "password123"

	// Register
	regDto := models.CreateUserDTO{Username: username, Email: email, Password: password}
	regBody, _ := json.Marshal(regDto)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	ta.App.Test(req)

	// Login
	loginDto := models.AuthUserDTO{Identifier: email, Password: password}
	loginBody, _ := json.Marshal(loginDto)
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := ta.App.Test(req)
	var loginResult map[string]string
	json.NewDecoder(resp.Body).Decode(&loginResult)
	token := loginResult["token"]

	var taskID int32

	t.Run("CreateTask", func(t *testing.T) {
		dto := models.CreateTaskDTO{
			Title:       "Test Task",
			Description: "Test Description",
			Difficulty:  1,
			Topics:      []int32{},
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/api/tasks/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var task db.Task
		json.NewDecoder(resp.Body).Decode(&task)
		taskID = task.ID
		assert.Equal(t, dto.Title, task.Title)
	})

	t.Run("ListTasks", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/tasks/?limit=10", nil)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tasks []db.Task
		json.NewDecoder(resp.Body).Decode(&tasks)
		assert.GreaterOrEqual(t, len(tasks), 1)
	})

	t.Run("GetTaskByID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/tasks/%d", taskID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res models.TaskByIdResponse
		json.NewDecoder(resp.Body).Decode(&res)
		assert.Equal(t, taskID, res.Task.ID)
	})

	t.Run("UpdateTask", func(t *testing.T) {
		dto := db.UpdateTaskParams{
			Title:       "Updated Task",
			Description: "Updated Description",
			Difficulty:  2,
			Topics:      []int32{},
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/tasks/%d", taskID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DeleteTask", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/tasks/%d", taskID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
