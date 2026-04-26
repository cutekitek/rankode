package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	db "rankode/internal/repository"
)

func TestLeaderboardHandler(t *testing.T) {
	ctx := context.Background()
	ta, err := setupTestApp(ctx)
	if err != nil {
		t.Fatalf("failed to setup test app: %v", err)
	}
	defer ta.DB.Close()

	t.Run("GetLeaderboard", func(t *testing.T) {
		req, _ := http.NewRequest("GET", testEndpoint("/api/leaderboard"), nil)
		resp, err := ta.App.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var leaderboard []db.GetUsersLeaderboardRow
		json.NewDecoder(resp.Body).Decode(&leaderboard)
		assert.NotNil(t, leaderboard)
	})
}
