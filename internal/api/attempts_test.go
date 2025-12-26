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

func TestAttemptsHandlers(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	// Register and login
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("attemptuser_%d", timestamp)
	email := fmt.Sprintf("attempt_%d@example.com", timestamp)
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

	// Create a task
	taskDto := models.CreateTaskDTO{Title: "Test Task for Attempt"}
	taskBody, _ := json.Marshal(taskDto)
	req, _ = http.NewRequest("POST", "/api/tasks/", bytes.NewBuffer(taskBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = ta.App.Test(req)
	var task db.Task
	json.NewDecoder(resp.Body).Decode(&task)
	taskID := task.ID

	// Create a test case (required by NewAttempt)
	tcDto := models.NewTestCaseReq{TaskID: taskID}
	tcBody, _ := json.Marshal(tcDto)
	req, _ = http.NewRequest("POST", "/api/test-cases/", bytes.NewBuffer(tcBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	ta.App.Test(req)

	t.Run("CreateAttempt", func(t *testing.T) {
		dto := models.CreateAttemptRequest{
			TaskID:   int(taskID),
			Language: "python",
			Code:     "print('hello')",
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/api/attempts/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("GetAttempts", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/attempts/?taskId=%d", taskID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var attempts []models.GetUserTaskAttemptsResponse
		json.NewDecoder(resp.Body).Decode(&attempts)
		assert.GreaterOrEqual(t, len(attempts), 1)
	})
}
