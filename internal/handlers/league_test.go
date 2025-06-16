package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sam-maryland/sleeper-mcp-server/internal/sleeper"
	"github.com/sirupsen/logrus/hooks/test"
)

// MockSleeperClient is a mock implementation of the sleeper.Client interface for testing
type MockSleeperClient struct {
	GetLeagueFunc           func(leagueID string) (*sleeper.League, error)
	GetUserFunc             func(usernameOrID string) (*sleeper.User, error)
	GetUserLeaguesFunc      func(userID, sport, season string) ([]sleeper.League, error)
	GetLeagueUsersFunc      func(leagueID string) ([]sleeper.User, error)
	GetLeagueRostersFunc    func(leagueID string) ([]sleeper.Roster, error)
	GetMatchupsFunc         func(leagueID string, week int) ([]sleeper.Matchup, error)
	GetTransactionsFunc     func(leagueID string, week int) ([]sleeper.Transaction, error)
	GetAllPlayersFunc       func() (map[string]sleeper.Player, error)
	GetTrendingPlayersFunc  func(sport, trendType string, hours, limit int) ([]sleeper.TrendingPlayer, error)
}

func (m *MockSleeperClient) GetUser(usernameOrID string) (*sleeper.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(usernameOrID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSleeperClient) GetUserLeagues(userID, sport, season string) ([]sleeper.League, error) {
	if m.GetUserLeaguesFunc != nil {
		return m.GetUserLeaguesFunc(userID, sport, season)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSleeperClient) GetLeague(leagueID string) (*sleeper.League, error) {
	if m.GetLeagueFunc != nil {
		return m.GetLeagueFunc(leagueID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSleeperClient) GetLeagueUsers(leagueID string) ([]sleeper.User, error) {
	if m.GetLeagueUsersFunc != nil {
		return m.GetLeagueUsersFunc(leagueID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSleeperClient) GetLeagueRosters(leagueID string) ([]sleeper.Roster, error) {
	if m.GetLeagueRostersFunc != nil {
		return m.GetLeagueRostersFunc(leagueID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSleeperClient) GetMatchups(leagueID string, week int) ([]sleeper.Matchup, error) {
	if m.GetMatchupsFunc != nil {
		return m.GetMatchupsFunc(leagueID, week)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSleeperClient) GetTransactions(leagueID string, week int) ([]sleeper.Transaction, error) {
	if m.GetTransactionsFunc != nil {
		return m.GetTransactionsFunc(leagueID, week)
	}
	return nil, errors.New("not implemented")
}

func (m *MockSleeperClient) GetAllPlayers() (map[string]sleeper.Player, error) {
	if m.GetAllPlayersFunc != nil {
		return m.GetAllPlayersFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockSleeperClient) GetTrendingPlayers(sport, trendType string, hours, limit int) ([]sleeper.TrendingPlayer, error) {
	if m.GetTrendingPlayersFunc != nil {
		return m.GetTrendingPlayersFunc(sport, trendType, hours, limit)
	}
	return nil, errors.New("not implemented")
}

func TestLeagueHandler_GetLeagueInfoTool(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockClient := &MockSleeperClient{}
	handler := NewLeagueHandler(mockClient, logger)

	tool := handler.GetLeagueInfoTool()

	// Test tool metadata
	if tool.Name != "get_league_info" {
		t.Errorf("Expected tool name 'get_league_info', got '%s'", tool.Name)
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
	}
}

func TestLeagueHandler_HandleGetLeagueInfo(t *testing.T) {
	tests := []struct {
		name           string
		args           map[string]interface{}
		mockResponse   *sleeper.League
		mockError      error
		wantError      bool
		expectErrorMsg bool
	}{
		{
			name: "successful request",
			args: map[string]interface{}{
				"league_id": "123456789",
			},
			mockResponse: &sleeper.League{
				LeagueID:     "123456789",
				Name:         "Test League",
				Status:       "in_season",
				Season:       "2024",
				TotalRosters: 12,
			},
			mockError:      nil,
			wantError:      false,
			expectErrorMsg: false,
		},
		{
			name: "missing league_id",
			args: map[string]interface{}{},
			mockResponse: nil,
			mockError:    nil,
			wantError:    true,
			expectErrorMsg: false,
		},
		{
			name: "invalid league_id type",
			args: map[string]interface{}{
				"league_id": 123,
			},
			mockResponse: nil,
			mockError:    nil,
			wantError:    true,
			expectErrorMsg: false,
		},
		{
			name: "sleeper api error",
			args: map[string]interface{}{
				"league_id": "invalid",
			},
			mockResponse:   nil,
			mockError:      errors.New("league not found"),
			wantError:      false,
			expectErrorMsg: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()
			
			mockClient := &MockSleeperClient{
				GetLeagueFunc: func(leagueID string) (*sleeper.League, error) {
					return tt.mockResponse, tt.mockError
				},
			}
			
			handler := NewLeagueHandler(mockClient, logger)
			ctx := context.Background()

			result, err := handler.HandleGetLeagueInfo(ctx, tt.args)

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

			// Check logging
			if tt.name == "successful request" && len(hook.Entries) == 0 {
				t.Error("Expected log entries for successful request")
			}
		})
	}
}

func TestNewLeagueHandler(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockClient := &MockSleeperClient{}
	
	handler := NewLeagueHandler(mockClient, logger)
	
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