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
)

func TestAuthHandlers(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	timestamp := time.Now().UnixNano()
	username := fmt.Sprintf("testuser_%d", timestamp)
	email := fmt.Sprintf("test_%d@example.com", timestamp)
	password := "password123"

	t.Run("Register", func(t *testing.T) {
		dto := models.CreateUserDTO{
			Username: username,
			Email:    email,
			Password: password,
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", testEndpoint("/api/auth/register"), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Login", func(t *testing.T) {
		dto := models.AuthUserDTO{
			Identifier: email,
			Password:   password,
		}
		body, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", testEndpoint("/api/auth/login"), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Contains(t, result, "token")
	})
}
