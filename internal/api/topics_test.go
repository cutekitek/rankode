package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	db "rankode/internal/repository"
)

func TestTopicsHandler(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	t.Run("ListTopics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testEndpoint("/api/topics/"), nil)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var topics []db.Topic
		json.NewDecoder(resp.Body).Decode(&topics)
		assert.NotNil(t, topics)
	})
}
