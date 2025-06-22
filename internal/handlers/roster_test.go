package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sam-maryland/sleeper-mcp-server/internal/sleeper"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestRosterHandler_GetRosterTool(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockClient := &MockSleeperClient{}
	handler := NewRosterHandler(mockClient, logger)

	tool := handler.GetRosterTool()

	// Test tool metadata
	if tool.Name != "get_roster" {
		t.Errorf("Expected tool name 'get_roster', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected tool description to be set")
	}

	// Test input schema
	if tool.InputSchema.Type != "object" {
		t.Errorf("Expected input schema type 'object', got '%s'", tool.InputSchema.Type)
	}

	// Check that league_id property exists
	if tool.InputSchema.Properties == nil {
		t.Error("Expected input schema properties to be set")
	} else {
		leagueIDProp, exists := tool.InputSchema.Properties["league_id"]
		if !exists {
			t.Error("Expected league_id property in input schema")
		} else {
			propMap, ok := leagueIDProp.(map[string]interface{})
			if !ok {
				t.Error("Expected league_id property to be a map")
			} else {
				if propMap["type"] != "string" {
					t.Errorf("Expected league_id type to be 'string', got '%v'", propMap["type"])
				}
			}
		}

		// Check that roster_id property exists
		rosterIDProp, exists := tool.InputSchema.Properties["roster_id"]
		if !exists {
			t.Error("Expected roster_id property in input schema")
		} else {
			propMap, ok := rosterIDProp.(map[string]interface{})
			if !ok {
				t.Error("Expected roster_id property to be a map")
			} else {
				if propMap["type"] != "integer" {
					t.Errorf("Expected roster_id type to be 'integer', got '%v'", propMap["type"])
				}
			}
		}
	}
}

func TestRosterHandler_GetAllRostersTool(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockClient := &MockSleeperClient{}
	handler := NewRosterHandler(mockClient, logger)

	tool := handler.GetAllRostersTool()

	// Test tool metadata
	if tool.Name != "get_all_rosters" {
		t.Errorf("Expected tool name 'get_all_rosters', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected tool description to be set")
	}

	// Test input schema
	if tool.InputSchema.Type != "object" {
		t.Errorf("Expected input schema type 'object', got '%s'", tool.InputSchema.Type)
	}
}

func TestRosterHandler_HandleGetRoster(t *testing.T) {
	tests := []struct {
		name           string
		args           map[string]interface{}
		mockRosters    []sleeper.Roster
		mockUsers      []sleeper.User
		mockPlayers    map[string]sleeper.Player
		mockError      error
		wantError      bool
		expectErrorMsg bool
	}{
		{
			name: "successful request",
			args: map[string]interface{}{
				"league_id": "123456789",
				"roster_id": float64(1),
			},
			mockRosters: []sleeper.Roster{
				{
					RosterID: 1,
					OwnerID:  "user1",
					Players:  []string{"123", "456", "789"},
					Starters: []string{"123", "456"},
					Settings: sleeper.RosterSettings{
						Wins:        10,
						Losses:      3,
						Ties:        0,
						FPTS:        1500.5,
						FPTSAgainst: 1200.2,
					},
				},
			},
			mockUsers: []sleeper.User{
				{
					UserID:      "user1",
					Username:    "testuser",
					DisplayName: "Test User",
				},
			},
			mockPlayers: map[string]sleeper.Player{
				"123": {
					PlayerID: "123",
					FullName: "Player One",
					Position: "QB",
					Team:     "KC",
				},
				"456": {
					PlayerID: "456",
					FullName: "Player Two",
					Position: "RB",
					Team:     "SF",
				},
				"789": {
					PlayerID: "789",
					FullName: "Player Three",
					Position: "WR",
					Team:     "BUF",
				},
			},
			mockError:      nil,
			wantError:      false,
			expectErrorMsg: false,
		},
		{
			name: "missing league_id",
			args: map[string]interface{}{
				"roster_id": float64(1),
			},
			mockRosters:    nil,
			mockUsers:      nil,
			mockPlayers:    nil,
			mockError:      nil,
			wantError:      false,
			expectErrorMsg: true,
		},
		{
			name: "missing roster_id",
			args: map[string]interface{}{
				"league_id": "123456789",
			},
			mockRosters:    nil,
			mockUsers:      nil,
			mockPlayers:    nil,
			mockError:      nil,
			wantError:      false,
			expectErrorMsg: true,
		},
		{
			name: "invalid roster_id type",
			args: map[string]interface{}{
				"league_id": "123456789",
				"roster_id": "invalid",
			},
			mockRosters:    nil,
			mockUsers:      nil,
			mockPlayers:    nil,
			mockError:      nil,
			wantError:      false,
			expectErrorMsg: true,
		},
		{
			name: "roster not found",
			args: map[string]interface{}{
				"league_id": "123456789",
				"roster_id": float64(999),
			},
			mockRosters: []sleeper.Roster{
				{
					RosterID: 1,
					OwnerID:  "user1",
					Players:  []string{"123"},
					Starters: []string{"123"},
					Settings: sleeper.RosterSettings{
						Wins:        5,
						Losses:      8,
						Ties:        0,
						FPTS:        1200.0,
						FPTSAgainst: 1400.0,
					},
				},
			},
			mockUsers:      []sleeper.User{},
			mockPlayers:    map[string]sleeper.Player{},
			mockError:      nil,
			wantError:      false,
			expectErrorMsg: true,
		},
		{
			name: "sleeper api error",
			args: map[string]interface{}{
				"league_id": "invalid",
				"roster_id": float64(1),
			},
			mockRosters:    nil,
			mockUsers:      nil,
			mockPlayers:    nil,
			mockError:      errors.New("league not found"),
			wantError:      false,
			expectErrorMsg: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()

			mockClient := &MockSleeperClient{
				GetLeagueRostersFunc: func(leagueID string) ([]sleeper.Roster, error) {
					return tt.mockRosters, tt.mockError
				},
				GetLeagueUsersFunc: func(leagueID string) ([]sleeper.User, error) {
					return tt.mockUsers, tt.mockError
				},
				GetAllPlayersFunc: func() (map[string]sleeper.Player, error) {
					return tt.mockPlayers, tt.mockError
				},
			}

			handler := NewRosterHandler(mockClient, logger)
			ctx := context.Background()

			result, err := handler.HandleGetRoster(ctx, tt.args)

			// Check error expectation
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result
			if !tt.wantError {
				if result == nil {
					t.Error("Expected result but got nil")
				} else {
					// Check if result indicates an error when expected
					if tt.expectErrorMsg && !result.IsError {
						t.Error("Expected result to indicate error")
					}
					if !tt.expectErrorMsg && result.IsError {
						t.Error("Expected successful result but got error")
					}

					// Check content exists
					if len(result.Content) == 0 {
						t.Error("Expected content in result")
					} else {
						textContent, ok := result.Content[0].(*mcp.TextContent)
						if !ok {
							t.Error("Expected text content")
						} else if textContent.Text == "" {
							t.Error("Expected non-empty text content")
						}
					}
				}
			}

			// Check logging for successful requests
			if tt.name == "successful request" && len(hook.Entries) == 0 {
				t.Error("Expected log entries for successful request")
			}
		})
	}
}

func TestRosterHandler_HandleGetAllRosters(t *testing.T) {
	logger, _ := test.NewNullLogger()

	mockRosters := []sleeper.Roster{
		{
			RosterID: 1,
			OwnerID:  "user1",
			Players:  []string{"123", "456"},
			Starters: []string{"123"},
			Settings: sleeper.RosterSettings{
				Wins:        10,
				Losses:      3,
				Ties:        0,
				FPTS:        1500.5,
				FPTSAgainst: 1200.2,
			},
		},
		{
			RosterID: 2,
			OwnerID:  "user2",
			Players:  []string{"789", "101"},
			Starters: []string{"789"},
			Settings: sleeper.RosterSettings{
				Wins:        8,
				Losses:      5,
				Ties:        0,
				FPTS:        1400.3,
				FPTSAgainst: 1300.1,
			},
		},
	}

	mockUsers := []sleeper.User{
		{
			UserID:      "user1",
			Username:    "player1",
			DisplayName: "Player One",
		},
		{
			UserID:      "user2",
			Username:    "player2",
			DisplayName: "Player Two",
		},
	}

	mockClient := &MockSleeperClient{
		GetLeagueRostersFunc: func(leagueID string) ([]sleeper.Roster, error) {
			return mockRosters, nil
		},
		GetLeagueUsersFunc: func(leagueID string) ([]sleeper.User, error) {
			return mockUsers, nil
		},
	}

	handler := NewRosterHandler(mockClient, logger)
	ctx := context.Background()
	args := map[string]interface{}{
		"league_id": "test123",
	}

	result, err := handler.HandleGetAllRosters(ctx, args)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected result but got nil")
	}

	if result.IsError {
		t.Error("Expected successful result but got error")
	}

	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}
}

func TestRosterHandler_HandleGetAllRosters_MissingLeagueID(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockClient := &MockSleeperClient{}
	handler := NewRosterHandler(mockClient, logger)
	ctx := context.Background()

	// Test missing league_id
	args := map[string]interface{}{}

	result, err := handler.HandleGetAllRosters(ctx, args)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected result but got nil")
	}

	if !result.IsError {
		t.Error("Expected result to indicate error")
	}
}

func TestNewRosterHandler(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockClient := &MockSleeperClient{}

	handler := NewRosterHandler(mockClient, logger)

	if handler == nil {
		t.Error("Expected handler to be created, got nil")
	}

	if handler.client != mockClient {
		t.Error("Expected handler to use provided client")
	}

	if handler.logger != logger {
		t.Error("Expected handler to use provided logger")
	}
}

func TestRosterHandler_AnalyzeRosterStrengthTool(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockClient := &MockSleeperClient{}
	handler := NewRosterHandler(mockClient, logger)

	tool := handler.AnalyzeRosterStrengthTool()

	// Test tool metadata
	if tool.Name != "analyze_roster_strength" {
		t.Errorf("Expected tool name 'analyze_roster_strength', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected tool description to be set")
	}
}

func TestRosterHandler_CompareRostersTool(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockClient := &MockSleeperClient{}
	handler := NewRosterHandler(mockClient, logger)

	tool := handler.CompareRostersTool()

	// Test tool metadata
	if tool.Name != "compare_rosters" {
		t.Errorf("Expected tool name 'compare_rosters', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected tool description to be set")
	}
}

func TestRosterHandler_HandleAnalyzeRosterStrength(t *testing.T) {
	logger, _ := test.NewNullLogger()

	mockRoster := sleeper.Roster{
		RosterID: 1,
		OwnerID:  "user1",
		Players:  []string{"123", "456", "789", "101", "202"},
		Starters: []string{"123", "456", "789"},
		Settings: sleeper.RosterSettings{
			Wins:        8,
			Losses:      5,
			Ties:        0,
			FPTS:        1400.5,
			FPTSAgainst: 1200.2,
		},
	}

	mockUsers := []sleeper.User{
		{
			UserID:      "user1",
			Username:    "testuser",
			DisplayName: "Test User",
		},
	}

	mockPlayers := map[string]sleeper.Player{
		"123": {
			PlayerID:     "123",
			FullName:     "Josh Allen",
			Position:     "QB",
			Team:         "BUF",
			InjuryStatus: "Healthy",
		},
		"456": {
			PlayerID:     "456",
			FullName:     "Derrick Henry",
			Position:     "RB",
			Team:         "TEN",
			InjuryStatus: "Healthy",
		},
		"789": {
			PlayerID:     "789",
			FullName:     "Davante Adams",
			Position:     "WR",
			Team:         "LV",
			InjuryStatus: "Questionable",
		},
		"101": {
			PlayerID:     "101",
			FullName:     "Travis Kelce",
			Position:     "TE",
			Team:         "KC",
			InjuryStatus: "Healthy",
		},
		"202": {
			PlayerID:     "202",
			FullName:     "Justin Tucker",
			Position:     "K",
			Team:         "BAL",
			InjuryStatus: "Healthy",
		},
	}

	mockClient := &MockSleeperClient{
		GetLeagueRostersFunc: func(leagueID string) ([]sleeper.Roster, error) {
			return []sleeper.Roster{mockRoster}, nil
		},
		GetLeagueUsersFunc: func(leagueID string) ([]sleeper.User, error) {
			return mockUsers, nil
		},
		GetAllPlayersFunc: func() (map[string]sleeper.Player, error) {
			return mockPlayers, nil
		},
	}

	handler := NewRosterHandler(mockClient, logger)
	ctx := context.Background()
	args := map[string]interface{}{
		"league_id": "test123",
		"roster_id": float64(1),
	}

	result, err := handler.HandleAnalyzeRosterStrength(ctx, args)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected result but got nil")
	}

	if result.IsError {
		t.Error("Expected successful result but got error")
	}

	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}
}

func TestRosterHandler_HandleCompareRosters(t *testing.T) {
	logger, _ := test.NewNullLogger()

	mockRosters := []sleeper.Roster{
		{
			RosterID: 1,
			OwnerID:  "user1",
			Players:  []string{"123", "456"},
			Starters: []string{"123"},
			Settings: sleeper.RosterSettings{
				Wins:        10,
				Losses:      3,
				Ties:        0,
				FPTS:        1500.5,
				FPTSAgainst: 1200.2,
			},
		},
		{
			RosterID: 2,
			OwnerID:  "user2",
			Players:  []string{"789", "101"},
			Starters: []string{"789"},
			Settings: sleeper.RosterSettings{
				Wins:        8,
				Losses:      5,
				Ties:        0,
				FPTS:        1400.3,
				FPTSAgainst: 1300.1,
			},
		},
	}

	mockUsers := []sleeper.User{
		{
			UserID:      "user1",
			Username:    "player1",
			DisplayName: "Player One",
		},
		{
			UserID:      "user2",
			Username:    "player2",
			DisplayName: "Player Two",
		},
	}

	mockPlayers := map[string]sleeper.Player{
		"123": {
			PlayerID: "123",
			FullName: "Josh Allen",
			Position: "QB",
			Team:     "BUF",
		},
		"456": {
			PlayerID: "456",
			FullName: "Derrick Henry",
			Position: "RB",
			Team:     "TEN",
		},
		"789": {
			PlayerID: "789",
			FullName: "Davante Adams",
			Position: "WR",
			Team:     "LV",
		},
		"101": {
			PlayerID: "101",
			FullName: "Travis Kelce",
			Position: "TE",
			Team:     "KC",
		},
	}

	mockClient := &MockSleeperClient{
		GetLeagueRostersFunc: func(leagueID string) ([]sleeper.Roster, error) {
			return mockRosters, nil
		},
		GetLeagueUsersFunc: func(leagueID string) ([]sleeper.User, error) {
			return mockUsers, nil
		},
		GetAllPlayersFunc: func() (map[string]sleeper.Player, error) {
			return mockPlayers, nil
		},
	}

	handler := NewRosterHandler(mockClient, logger)
	ctx := context.Background()
	args := map[string]interface{}{
		"league_id":  "test123",
		"roster_id1": float64(1),
		"roster_id2": float64(2),
	}

	result, err := handler.HandleCompareRosters(ctx, args)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Expected result but got nil")
	}

	if result.IsError {
		t.Error("Expected successful result but got error")
	}

	if len(result.Content) == 0 {
		t.Error("Expected content in result")
	}
}