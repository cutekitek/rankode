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

func TestTestCasesHandlers(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	// Register and login
	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("tcuser_%d", timestamp)
	email := fmt.Sprintf("tc_%d@example.com", timestamp)
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

	// Create a task
	taskDto := models.CreateTaskDTO{Title: "Test Task"}
	taskBody, _ := json.Marshal(taskDto)
	req, _ = http.NewRequest("POST", testEndpoint("/api/tasks/"), bytes.NewBuffer(taskBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ = ta.App.Test(req)
	var task db.Task
	json.NewDecoder(resp.Body).Decode(&task)
	taskID := task.ID

	t.Run("CreateTestCase", func(t *testing.T) {
		dto := models.NewTestCaseReq{
			TaskID: taskID,
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", testEndpoint("/api/test-cases/"), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var tc db.TaskTestCase
		json.NewDecoder(resp.Body).Decode(&tc)
		assert.Equal(t, taskID, tc.TaskID)
	})
}
