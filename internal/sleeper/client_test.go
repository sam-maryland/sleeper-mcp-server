package sleeper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestHTTPClient_GetLeague(t *testing.T) {
	tests := []struct {
		name           string
		leagueID       string
		serverResponse string
		serverStatus   int
		wantError      bool
		wantLeague     *League
	}{
		{
			name:         "successful request",
			leagueID:     "123456789",
			serverStatus: http.StatusOK,
			serverResponse: `{
				"league_id": "123456789",
				"name": "Test League",
				"status": "in_season",
				"sport": "nfl",
				"season": "2024",
				"total_rosters": 12,
				"settings": {},
				"scoring_settings": {},
				"roster_positions": ["QB", "RB", "WR", "TE", "FLEX", "K", "DEF"]
			}`,
			wantError: false,
			wantLeague: &League{
				LeagueID:        "123456789",
				Name:            "Test League",
				Status:          "in_season",
				Sport:           "nfl",
				Season:          "2024",
				TotalRosters:    12,
				Settings:        LeagueSettings{},
				ScoringSettings: map[string]float64{},
				RosterPositions: []string{"QB", "RB", "WR", "TE", "FLEX", "K", "DEF"},
			},
		},
		{
			name:           "league not found",
			leagueID:       "invalid",
			serverStatus:   http.StatusNotFound,
			serverResponse: "null",
			wantError:      true,
			wantLeague:     nil,
		},
		{
			name:           "server error",
			leagueID:       "123456789",
			serverStatus:   http.StatusInternalServerError,
			serverResponse: "Internal Server Error",
			wantError:      true,
			wantLeague:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/league/"+tt.leagueID {
					t.Errorf("Expected path /league/%s, got %s", tt.leagueID, r.URL.Path)
				}
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create client with test server URL
			logger, _ := test.NewNullLogger()
			client := &HTTPClient{
				baseURL:    server.URL,
				httpClient: &http.Client{},
				logger:     logger,
			}

			// Call the method
			league, err := client.GetLeague(tt.leagueID)

			// Check error expectation
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check league result
			if tt.wantLeague != nil {
				if league == nil {
					t.Error("Expected league but got nil")
				} else {
					if league.LeagueID != tt.wantLeague.LeagueID {
						t.Errorf("Expected league ID %s, got %s", tt.wantLeague.LeagueID, league.LeagueID)
					}
					if league.Name != tt.wantLeague.Name {
						t.Errorf("Expected league name %s, got %s", tt.wantLeague.Name, league.Name)
					}
				}
			} else if league != nil {
				t.Error("Expected nil league but got result")
			}
		})
	}
}

func TestHTTPClient_GetTrendingPlayers(t *testing.T) {
	tests := []struct {
		name           string
		sport          string
		trendType      string
		hours          int
		limit          int
		serverResponse string
		serverStatus   int
		wantError      bool
		wantCount      int
	}{
		{
			name:         "successful trending adds",
			sport:        "nfl",
			trendType:    "add",
			hours:        24,
			limit:        5,
			serverStatus: http.StatusOK,
			serverResponse: `[
				{"player_id": "4171", "count": 150},
				{"player_id": "4172", "count": 120},
				{"player_id": "4173", "count": 100}
			]`,
			wantError: false,
			wantCount: 3,
		},
		{
			name:           "server error",
			sport:          "nfl",
			trendType:      "add",
			hours:          24,
			limit:          5,
			serverStatus:   http.StatusInternalServerError,
			serverResponse: "Internal Server Error",
			wantError:      true,
			wantCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/players/" + tt.sport + "/trending/" + tt.trendType
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}
				
				// Check query parameters
				query := r.URL.Query()
				if tt.hours > 0 {
					if query.Get("lookback_hours") != "24" {
						t.Errorf("Expected lookback_hours=24, got %s", query.Get("lookback_hours"))
					}
				}
				if tt.limit > 0 {
					if query.Get("limit") != "5" {
						t.Errorf("Expected limit=5, got %s", query.Get("limit"))
					}
				}
				
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create client with test server URL
			logger, _ := test.NewNullLogger()
			client := &HTTPClient{
				baseURL:    server.URL,
				httpClient: &http.Client{},
				logger:     logger,
			}

			// Call the method
			trending, err := client.GetTrendingPlayers(tt.sport, tt.trendType, tt.hours, tt.limit)

			// Check error expectation
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result count
			if len(trending) != tt.wantCount {
				t.Errorf("Expected %d trending players, got %d", tt.wantCount, len(trending))
			}

			// Check first player if we have results
			if tt.wantCount > 0 && len(trending) > 0 {
				if trending[0].PlayerID != "4171" {
					t.Errorf("Expected first player ID 4171, got %s", trending[0].PlayerID)
				}
				if trending[0].Count != 150 {
					t.Errorf("Expected first player count 150, got %d", trending[0].Count)
				}
			}
		})
	}
}

func TestSleeperError_Error(t *testing.T) {
	err := &SleeperError{
		Type:    "api_error",
		Message: "League not found",
	}

	expected := "League not found"
	if err.Error() != expected {
		t.Errorf("Expected error message %s, got %s", expected, err.Error())
	}
}

func TestNewHTTPClient(t *testing.T) {
	logger := logrus.New()
	client := NewHTTPClient(logger)

	if client == nil {
		t.Error("Expected client to be created, got nil")
	}

	// Ensure it implements the Client interface
	var _ Client = client
}