package sleeper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	BaseURL    = "https://api.sleeper.app/v1"
	DefaultTimeout = 10 * time.Second
)

// Client defines the interface for interacting with the Sleeper API
type Client interface {
	// User methods
	GetUser(usernameOrID string) (*User, error)
	GetUserLeagues(userID, sport, season string) ([]League, error)
	
	// League methods
	GetLeague(leagueID string) (*League, error)
	GetLeagueUsers(leagueID string) ([]User, error)
	GetLeagueRosters(leagueID string) ([]Roster, error)
	GetMatchups(leagueID string, week int) ([]Matchup, error)
	GetTransactions(leagueID string, week int) ([]Transaction, error)
	GetWinnersBracket(leagueID string) ([]BracketMatchup, error)
	GetLosersBracket(leagueID string) ([]BracketMatchup, error)
	
	// Player methods
	GetAllPlayers() (map[string]Player, error)
	GetTrendingPlayers(sport, trendType string, hours, limit int) ([]TrendingPlayer, error)
}

// HTTPClient implements the Client interface using HTTP requests
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewHTTPClient creates a new HTTP client for the Sleeper API
func NewHTTPClient(logger *logrus.Logger) Client {
	return &HTTPClient{
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		logger: logger,
	}
}

// makeRequest performs an HTTP GET request to the Sleeper API
func (c *HTTPClient) makeRequest(endpoint string, result interface{}) error {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	
	c.logger.WithField("url", url).Debug("Making API request")
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		c.logger.WithError(err).Error("HTTP request failed")
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.WithError(err).Error("Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("API request failed")
		
		return &SleeperError{
			Type:       "api_error",
			Message:    fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}
	
	if err := json.Unmarshal(body, result); err != nil {
		c.logger.WithError(err).WithField("body", string(body)).Error("Failed to unmarshal response")
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	c.logger.Debug("API request completed successfully")
	return nil
}

// GetUser retrieves a user by username or user ID
func (c *HTTPClient) GetUser(usernameOrID string) (*User, error) {
	endpoint := fmt.Sprintf("/user/%s", usernameOrID)
	var user User
	
	if err := c.makeRequest(endpoint, &user); err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", usernameOrID, err)
	}
	
	return &user, nil
}

// GetUserLeagues retrieves all leagues for a user in a given sport and season
func (c *HTTPClient) GetUserLeagues(userID, sport, season string) ([]League, error) {
	endpoint := fmt.Sprintf("/user/%s/leagues/%s/%s", userID, sport, season)
	var leagues []League
	
	if err := c.makeRequest(endpoint, &leagues); err != nil {
		return nil, fmt.Errorf("failed to get leagues for user %s: %w", userID, err)
	}
	
	return leagues, nil
}

// GetLeague retrieves comprehensive league information
func (c *HTTPClient) GetLeague(leagueID string) (*League, error) {
	endpoint := fmt.Sprintf("/league/%s", leagueID)
	var league League
	
	if err := c.makeRequest(endpoint, &league); err != nil {
		return nil, fmt.Errorf("failed to get league %s: %w", leagueID, err)
	}
	
	return &league, nil
}

// GetLeagueUsers retrieves all users in a league
func (c *HTTPClient) GetLeagueUsers(leagueID string) ([]User, error) {
	endpoint := fmt.Sprintf("/league/%s/users", leagueID)
	var users []User
	
	if err := c.makeRequest(endpoint, &users); err != nil {
		return nil, fmt.Errorf("failed to get users for league %s: %w", leagueID, err)
	}
	
	return users, nil
}

// GetLeagueRosters retrieves all rosters in a league
func (c *HTTPClient) GetLeagueRosters(leagueID string) ([]Roster, error) {
	endpoint := fmt.Sprintf("/league/%s/rosters", leagueID)
	var rosters []Roster
	
	if err := c.makeRequest(endpoint, &rosters); err != nil {
		return nil, fmt.Errorf("failed to get rosters for league %s: %w", leagueID, err)
	}
	
	return rosters, nil
}

// GetMatchups retrieves matchups for a specific week
func (c *HTTPClient) GetMatchups(leagueID string, week int) ([]Matchup, error) {
	endpoint := fmt.Sprintf("/league/%s/matchups/%d", leagueID, week)
	var matchups []Matchup
	
	if err := c.makeRequest(endpoint, &matchups); err != nil {
		return nil, fmt.Errorf("failed to get matchups for league %s week %d: %w", leagueID, week, err)
	}
	
	return matchups, nil
}

// GetTransactions retrieves transactions for a specific week
func (c *HTTPClient) GetTransactions(leagueID string, week int) ([]Transaction, error) {
	endpoint := fmt.Sprintf("/league/%s/transactions/%d", leagueID, week)
	var transactions []Transaction
	
	if err := c.makeRequest(endpoint, &transactions); err != nil {
		return nil, fmt.Errorf("failed to get transactions for league %s week %d: %w", leagueID, week, err)
	}
	
	return transactions, nil
}

// GetAllPlayers retrieves all NFL players (use sparingly)
func (c *HTTPClient) GetAllPlayers() (map[string]Player, error) {
	endpoint := "/players/nfl"
	var players map[string]Player
	
	if err := c.makeRequest(endpoint, &players); err != nil {
		return nil, fmt.Errorf("failed to get all players: %w", err)
	}
	
	return players, nil
}

// GetTrendingPlayers retrieves trending players for adds/drops
func (c *HTTPClient) GetTrendingPlayers(sport, trendType string, hours, limit int) ([]TrendingPlayer, error) {
	endpoint := fmt.Sprintf("/players/%s/trending/%s", sport, trendType)
	
	// Add query parameters if provided
	if hours > 0 || limit > 0 {
		endpoint += "?"
		if hours > 0 {
			endpoint += fmt.Sprintf("lookback_hours=%d", hours)
		}
		if limit > 0 {
			if hours > 0 {
				endpoint += "&"
			}
			endpoint += fmt.Sprintf("limit=%d", limit)
		}
	}
	
	var trending []TrendingPlayer
	
	if err := c.makeRequest(endpoint, &trending); err != nil {
		return nil, fmt.Errorf("failed to get trending players: %w", err)
	}
	
	return trending, nil
}

// GetWinnersBracket retrieves the winners bracket for a league
func (c *HTTPClient) GetWinnersBracket(leagueID string) ([]BracketMatchup, error) {
	endpoint := fmt.Sprintf("/league/%s/winners_bracket", leagueID)
	var bracket []BracketMatchup
	
	if err := c.makeRequest(endpoint, &bracket); err != nil {
		return nil, fmt.Errorf("failed to get winners bracket: %w", err)
	}
	
	return bracket, nil
}

// GetLosersBracket retrieves the losers bracket for a league
func (c *HTTPClient) GetLosersBracket(leagueID string) ([]BracketMatchup, error) {
	endpoint := fmt.Sprintf("/league/%s/losers_bracket", leagueID)
	var bracket []BracketMatchup
	
	if err := c.makeRequest(endpoint, &bracket); err != nil {
		return nil, fmt.Errorf("failed to get losers bracket: %w", err)
	}
	
	return bracket, nil
}