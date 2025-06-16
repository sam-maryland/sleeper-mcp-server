//go:build integration
// +build integration

package main

import (
	"context"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sam-maryland/sleeper-mcp-server/internal/handlers"
	"github.com/sam-maryland/sleeper-mcp-server/internal/sleeper"
	"github.com/sirupsen/logrus/hooks/test"
)

// Integration tests that actually call the Sleeper API
// Run with: go test -tags=integration ./...

func TestIntegration_SleeperAPI_GetTrendingPlayers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger, _ := test.NewNullLogger()
	client := sleeper.NewHTTPClient(logger)

	trending, err := client.GetTrendingPlayers("nfl", "add", 24, 5)
	if err != nil {
		t.Errorf("Failed to get trending players: %v", err)
	}

	if len(trending) == 0 {
		t.Error("Expected at least some trending players")
	}

	// Check first trending player has valid data
	if len(trending) > 0 {
		if trending[0].PlayerID == "" {
			t.Error("Expected player ID to be set")
		}
		if trending[0].Count <= 0 {
			t.Error("Expected player count to be positive")
		}
	}
}

func TestIntegration_LeagueHandler_WithRealLeague(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use environment variable for league ID to avoid hardcoding
	leagueID := os.Getenv("TEST_LEAGUE_ID")
	if leagueID == "" {
		t.Skip("TEST_LEAGUE_ID environment variable not set, skipping integration test")
	}

	logger, _ := test.NewNullLogger()
	client := sleeper.NewHTTPClient(logger)
	handler := handlers.NewLeagueHandler(client, logger)

	// Test get_league_info with real league
	args := map[string]interface{}{
		"league_id": leagueID,
	}

	result, err := handler.HandleGetLeagueInfo(context.Background(), args)
	if err != nil {
		t.Errorf("Failed to handle get_league_info: %v", err)
	}

	if result == nil {
		t.Error("Expected result but got nil")
	}

	if result.IsError {
		t.Errorf("Expected successful result but got error: %v", result.Content)
	}

	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}

	// Verify the response contains league data
	textContent := result.Content[0].(*mcp.TextContent)
	if textContent.Text == "" {
		t.Error("Expected non-empty response text")
	}

	// Basic validation that it looks like JSON
	if textContent.Text[0] != '{' {
		t.Error("Expected JSON response to start with '{'")
	}
}

func TestIntegration_SleeperAPI_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger, _ := test.NewNullLogger()
	client := sleeper.NewHTTPClient(logger)

	// Test with invalid league ID
	_, err := client.GetLeague("invalid_league_id")
	if err == nil {
		t.Error("Expected error for invalid league ID")
	}

	// Check that the error contains information about the failure
	if err.Error() == "" {
		t.Error("Expected error message to be non-empty")
	}
	
	// Log the error type for debugging (this shows it's wrapped)
	t.Logf("Error type: %T, message: %s", err, err.Error())
}
