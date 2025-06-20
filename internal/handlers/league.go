package handlers

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sam-maryland/sleeper-mcp-server/internal/config"
	"github.com/sam-maryland/sleeper-mcp-server/internal/sleeper"
	"github.com/sirupsen/logrus"
)

// LeagueInfoArgs represents the parameters for the get_league_info tool
type LeagueInfoArgs struct {
	LeagueID string `json:"league_id"`
}

// LeagueStandingsArgs represents the parameters for the get_league_standings tool
type LeagueStandingsArgs struct {
	LeagueID       string   `json:"league_id"`
	TiebreakOrder  []string `json:"tiebreak_order,omitempty"`
	CustomMetrics  map[string]interface{} `json:"custom_metrics,omitempty"`
	Instructions   string   `json:"instructions,omitempty"`
	Mode           string   `json:"mode,omitempty"`
}

// TiebreakerType represents different tiebreaker methods
type TiebreakerType string

const (
	TiebreakerWins          TiebreakerType = "wins"
	TiebreakerPointsFor     TiebreakerType = "points_for" 
	TiebreakerPointsAgainst TiebreakerType = "points_against"
	TiebreakerHeadToHead    TiebreakerType = "head_to_head"
	TiebreakerDivisionRecord TiebreakerType = "division_record"
	TiebreakerCustom        TiebreakerType = "custom"
)

// LeagueUsersArgs represents the parameters for the get_league_users tool
type LeagueUsersArgs struct {
	LeagueID string `json:"league_id"`
}

// MatchupsArgs represents the parameters for the get_matchups tool
type MatchupsArgs struct {
	LeagueID string `json:"league_id"`
	Week     int    `json:"week"`
}

// StandingEntry represents a team's standing in the league
type StandingEntry struct {
	RosterID           int                    `json:"roster_id"`
	OwnerID            string                 `json:"owner_id"`
	DisplayName        string                 `json:"display_name"`
	TeamName           string                 `json:"team_name"`
	Wins               int                    `json:"wins"`
	Losses             int                    `json:"losses"`
	Ties               int                    `json:"ties"`
	PointsFor          float64                `json:"points_for"`
	PointsAgainst      float64                `json:"points_against"`
	Rank               int                    `json:"rank"`
	Division           int                    `json:"division,omitempty"`
	PlayoffSeed        int                    `json:"playoff_seed,omitempty"`
	PlayoffOutcome     string                 `json:"playoff_outcome,omitempty"`     // "champion", "runner_up", "third_place", "fourth_place", "quarterfinal_loss", "no_playoffs"
	RegularSeasonRank  int                    `json:"regular_season_rank,omitempty"` // Original regular season ranking
	CustomMetrics      map[string]interface{} `json:"custom_metrics,omitempty"`
	TiebreakerNotes    string                 `json:"tiebreaker_notes,omitempty"`
	HeadToHeadWins     map[int]int            `json:"head_to_head_wins,omitempty"`
	RandomTiebreakerID string                 `json:"random_tiebreaker_id,omitempty"` // For random tiebreaker consistency
}

// PlayoffBracket represents the playoff bracket structure and results
type PlayoffBracket struct {
	QuarterfinalsWeek int                        `json:"quarterfinals_week"`
	SemifinalsWeek    int                        `json:"semifinals_week"`
	ChampionshipWeek  int                        `json:"championship_week"`
	ThirdPlaceWeek    int                        `json:"third_place_week"`
	HasThirdPlace     bool                       `json:"has_third_place"`
	QuarterFinals     []PlayoffMatchup           `json:"quarterfinals"`
	SemiFinals        []PlayoffMatchup           `json:"semifinals"`
	Championship      *PlayoffMatchup            `json:"championship"`
	ThirdPlace        *PlayoffMatchup            `json:"third_place"`
	PlayoffTeams      map[int]int                `json:"playoff_teams"` // RosterID -> PlayoffSeed
}

// PlayoffMatchup represents a single playoff game
type PlayoffMatchup struct {
	Week        int     `json:"week"`
	MatchupID   int     `json:"matchup_id"`
	Team1       int     `json:"team1_roster_id"`
	Team2       int     `json:"team2_roster_id"`
	Team1Points float64 `json:"team1_points"`
	Team2Points float64 `json:"team2_points"`
	Winner      int     `json:"winner_roster_id"`
	Loser       int     `json:"loser_roster_id"`
	GameType    string  `json:"game_type"` // "quarterfinal", "semifinal", "championship", "third_place"
}

// PlayoffStructure represents the detected playoff schedule and format
type PlayoffStructure struct {
	QuarterfinalsWeek int `json:"quarterfinals_week"`
	SemifinalsWeek    int `json:"semifinals_week"`
	ChampionshipWeek  int `json:"championship_week"`
	ThirdPlaceWeek    int `json:"third_place_week"`
	Format            string `json:"format"` // "6team", "4team", "8team", etc.
	HasThirdPlace     bool `json:"has_third_place"`
}

// LeagueHandler handles league-related MCP tools
type LeagueHandler struct {
	client sleeper.Client
	logger *logrus.Logger
	config *config.LeagueConfig
}

// NewLeagueHandler creates a new league handler
func NewLeagueHandler(client sleeper.Client, logger *logrus.Logger) *LeagueHandler {
	// Load league configuration
	leagueConfig, err := config.LoadLeagueSettings()
	if err != nil {
		logger.WithError(err).Warn("Failed to load league settings, using defaults")
		// Create default config
		leagueConfig = &config.LeagueConfig{
			Leagues: make(map[string]config.LeagueSettings),
			DefaultSettings: config.LeagueSettings{
				Name:        "Default League",
				Description: "League with standard Sleeper tiebreakers",
				Custom: config.CustomStandings{
					Enabled:       false,
					Instructions:  "Use Sleeper default tiebreakers",
					TiebreakOrder: []string{"wins", "points_for"},
					Notes:         "Standard Sleeper tiebreaker rules apply",
				},
			},
		}
	}
	
	return &LeagueHandler{
		client: client,
		logger: logger,
		config: leagueConfig,
	}
}

// GetLeagueInfoTool returns the MCP tool definition for get_league_info
func (h *LeagueHandler) GetLeagueInfoTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_league_info",
		Description: "Get comprehensive league information including settings, scoring rules, and metadata",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"league_id": map[string]interface{}{
					"type":        "string",
					"description": "The Sleeper league ID",
					"required":    true,
				},
			},
		},
	}
}

// HandleGetLeagueInfo handles the get_league_info tool call
func (h *LeagueHandler) HandleGetLeagueInfo(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.WithField("args", args).Info("Handling get_league_info")
	
	// Parse arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return nil, fmt.Errorf("league_id is required and must be a string")
	}
	
	// Get league information from Sleeper API
	league, err := h.client.GetLeague(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league info")
		
		// Return error as MCP response
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to get league information: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}
	
	// Create response
	response := sleeper.APIResponse{
		Success: true,
		Data:    league,
		Summary: fmt.Sprintf("League '%s' (%s) - %s season, %d teams, %s status", 
			league.Name, league.LeagueID, league.Season, league.TotalRosters, league.Status),
		Metadata: sleeper.Metadata{
			Timestamp:    time.Now(),
			Source:       "sleeper_api",
			CacheHit:     false,
			APICallsUsed: 1,
			LeagueID:     leagueID,
		},
	}
	
	// Convert to JSON string for MCP response
	jsonResponse, err := formatJSONResponse(response)
	if err != nil {
		h.logger.WithError(err).Error("Failed to format response")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error formatting response: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Type: "text",
				Text: jsonResponse,
			},
		},
	}, nil
}

// GetLeagueStandingsTool returns the MCP tool definition for get_league_standings
func (h *LeagueHandler) GetLeagueStandingsTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_league_standings",
		Description: "Get current league standings with wins, losses, points scored, and playoff positioning. Supports custom tiebreaker rules.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"league_id": map[string]interface{}{
					"type":        "string",
					"description": "The Sleeper league ID",
					"required":    true,
				},
				"tiebreak_order": map[string]interface{}{
					"type":        "array",
					"description": "Custom tiebreaker order. Options: wins, points_for, points_against, head_to_head, division_record, custom",
					"items": map[string]interface{}{
						"type": "string",
					},
					"required": false,
				},
				"custom_metrics": map[string]interface{}{
					"type":        "object",
					"description": "Custom metrics for tiebreakers (e.g., all_play_record, bench_points, etc.)",
					"required":    false,
				},
				"instructions": map[string]interface{}{
					"type":        "string",
					"description": "Natural language instructions for how standings should be calculated (e.g., 'Use head-to-head record as first tiebreaker, then points for')",
					"required":    false,
				},
				"mode": map[string]interface{}{
					"type":        "string",
					"description": "Standings mode: 'regular_season' or 'final' (default: regular_season)",
					"required":    false,
				},
			},
		},
	}
}

// HandleGetLeagueStandings handles the get_league_standings tool call
func (h *LeagueHandler) HandleGetLeagueStandings(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.WithField("args", args).Info("Handling get_league_standings")
	
	// Parse arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return nil, fmt.Errorf("league_id is required and must be a string")
	}
	
	// Parse optional tiebreaker order
	var tiebreakOrder []string
	if tiebreakOrderRaw, exists := args["tiebreak_order"]; exists {
		if tiebreakArray, ok := tiebreakOrderRaw.([]interface{}); ok {
			for _, item := range tiebreakArray {
				if str, ok := item.(string); ok {
					tiebreakOrder = append(tiebreakOrder, str)
				}
			}
		}
	}
	
	// Parse optional custom metrics
	var customMetrics map[string]interface{}
	if customMetricsRaw, exists := args["custom_metrics"]; exists {
		if metrics, ok := customMetricsRaw.(map[string]interface{}); ok {
			customMetrics = metrics
		}
	}
	
	// Parse optional instructions
	var instructions string
	if instructionsRaw, exists := args["instructions"]; exists {
		if str, ok := instructionsRaw.(string); ok {
			instructions = str
		}
	}
	
	// Parse optional mode
	mode := "regular_season" // default
	if modeRaw, exists := args["mode"]; exists {
		if str, ok := modeRaw.(string); ok && str != "" {
			mode = str
		}
	}
	
	// Check for league-specific configuration
	if h.config != nil && h.config.HasCustomStandings(leagueID) {
		leagueSettings := h.config.GetLeagueSettings(leagueID)
		
		// Apply league configuration if not explicitly overridden
		if instructions == "" && leagueSettings.Custom.Instructions != "" {
			instructions = leagueSettings.Custom.Instructions
			h.logger.WithField("league_id", leagueID).Info("Applied league-specific custom instructions")
		}
		
		if len(tiebreakOrder) == 0 && len(leagueSettings.Custom.TiebreakOrder) > 0 {
			tiebreakOrder = leagueSettings.Custom.TiebreakOrder
			h.logger.WithField("league_id", leagueID).Info("Applied league-specific tiebreaker order")
		}
	}
	
	// Get league information to check playoff seed type
	league, err := h.client.GetLeague(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league info")
	}
	
	// Get league rosters for standings calculation
	rosters, err := h.client.GetLeagueRosters(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league rosters")
		
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to get league standings: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}

	// Get league users for team owner information
	users, err := h.client.GetLeagueUsers(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league users")
		
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to get league users: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}
	
	// Create user map for quick lookup
	userMap := make(map[string]*sleeper.User)
	for i := range users {
		userMap[users[i].UserID] = &users[i]
	}
	
	var standings []StandingEntry
	for _, roster := range rosters {
		entry := StandingEntry{
			RosterID:      roster.RosterID,
			OwnerID:       roster.OwnerID,
			Wins:          roster.Settings.Wins,
			Losses:        roster.Settings.Losses,
			Ties:          roster.Settings.Ties,
			PointsFor:     roster.Settings.FPTS,
			PointsAgainst: roster.Settings.FPTSAgainst,
			Division:      roster.Settings.Division,
			PlayoffSeed:   roster.Settings.PlayoffSeed,
		}
		
		// Add user information if available
		if user, exists := userMap[roster.OwnerID]; exists {
			entry.DisplayName = user.DisplayName
		}
		
		// Add custom metrics if provided
		if customMetrics != nil {
			entry.CustomMetrics = make(map[string]interface{})
			for key, value := range customMetrics {
				// This would be populated with actual calculated metrics
				// For now, just pass through the provided custom metrics
				entry.CustomMetrics[key] = value
			}
		}
		
		standings = append(standings, entry)
	}
	
	// Determine tiebreaker order from instructions, explicit order, or league settings
	effectiveTiebreakOrder := getTiebreakOrder(tiebreakOrder, league)
	if instructions != "" {
		effectiveTiebreakOrder = parseInstructionsToTiebreakOrder(instructions, effectiveTiebreakOrder)
	}
	
	// Calculate head-to-head records if needed
	var headToHeadMatrix map[int]map[int]int
	needsHeadToHead := containsTiebreaker(effectiveTiebreakOrder, "head_to_head")
	if needsHeadToHead {
		headToHeadMatrix, err = h.calculateHeadToHeadMatrix(leagueID, mode)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to calculate head-to-head matrix, skipping head-to-head tiebreaker")
			needsHeadToHead = false
		}
	}
	
	// Add head-to-head data to standings entries
	if needsHeadToHead && headToHeadMatrix != nil {
		for i := range standings {
			if h2hRecord, exists := headToHeadMatrix[standings[i].RosterID]; exists {
				standings[i].HeadToHeadWins = h2hRecord
			}
		}
	}
	
	// Apply random tiebreaker IDs if needed
	h.applyRandomTiebreaker(standings, effectiveTiebreakOrder)
	
	// Sort standings using flexible tiebreaker system
	standings = sortStandingsWithTiebreakers(standings, effectiveTiebreakOrder, customMetrics, headToHeadMatrix)
	
	// For final mode, process playoff bracket and calculate final standings
	if mode == "final" && league != nil {
		// First, get regular season standings (preserve current standings as regular season)
		regularSeasonStandings := make([]StandingEntry, len(standings))
		copy(regularSeasonStandings, standings)
		
		// Process playoff bracket
		bracket, err := h.processPlayoffBracket(leagueID, league, regularSeasonStandings)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to process playoff bracket, using regular season standings")
		} else {
			// Validate playoff bracket completeness
			if validationErr := h.validatePlayoffBracket(bracket); validationErr != nil {
				h.logger.WithError(validationErr).Warn("Playoff bracket validation failed, using regular season standings")
			} else {
				// Calculate final standings based on playoff results
				standings = h.calculateFinalStandings(regularSeasonStandings, bracket)
				h.logger.Info("Successfully calculated final standings based on playoff results")
			}
		}
	}
	
	// Add notes about tiebreakers used
	tiebreakerNotes := fmt.Sprintf("Tiebreakers applied: %v", effectiveTiebreakOrder)
	if league != nil && league.Settings.PlayoffSeedType != 0 {
		tiebreakerNotes += fmt.Sprintf(" (League playoff_seed_type: %d)", league.Settings.PlayoffSeedType)
	}
	if mode == "final" {
		tiebreakerNotes += " (Final standings based on playoff results)"
	}
	
	// Add rank to each entry and tiebreaker notes
	for i := range standings {
		standings[i].Rank = i + 1
		standings[i].TiebreakerNotes = tiebreakerNotes
	}
	
	// Create response
	response := sleeper.APIResponse{
		Success: true,
		Data:    standings,
		Summary: fmt.Sprintf("League standings for %d teams - Leader: %s (%d-%d, %.1f pts)", 
			len(standings), standings[0].DisplayName, standings[0].Wins, standings[0].Losses, standings[0].PointsFor),
		Metadata: sleeper.Metadata{
			Timestamp:    time.Now(),
			Source:       "sleeper_api",
			CacheHit:     false,
			APICallsUsed: 3,
			LeagueID:     leagueID,
		},
	}
	
	// Convert to JSON string for MCP response
	jsonResponse, err := formatJSONResponse(response)
	if err != nil {
		h.logger.WithError(err).Error("Failed to format response")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error formatting response: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Type: "text",
				Text: jsonResponse,
			},
		},
	}, nil
}

// GetLeagueUsersTool returns the MCP tool definition for get_league_users  
func (h *LeagueHandler) GetLeagueUsersTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_league_users", 
		Description: "Get all league members and their team information",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"league_id": map[string]interface{}{
					"type":        "string",
					"description": "The Sleeper league ID",
					"required":    true,
				},
			},
		},
	}
}

// HandleGetLeagueUsers handles the get_league_users tool call
func (h *LeagueHandler) HandleGetLeagueUsers(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.WithField("args", args).Info("Handling get_league_users")
	
	// Parse arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return nil, fmt.Errorf("league_id is required and must be a string")
	}
	
	// Get league users from Sleeper API
	users, err := h.client.GetLeagueUsers(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league users")
		
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to get league users: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}
	
	// Create response
	response := sleeper.APIResponse{
		Success: true,
		Data:    users,
		Summary: fmt.Sprintf("Found %d league members", len(users)),
		Metadata: sleeper.Metadata{
			Timestamp:    time.Now(),
			Source:       "sleeper_api",
			CacheHit:     false,
			APICallsUsed: 1,
			LeagueID:     leagueID,
		},
	}
	
	// Convert to JSON string for MCP response
	jsonResponse, err := formatJSONResponse(response)
	if err != nil {
		h.logger.WithError(err).Error("Failed to format response")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error formatting response: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Type: "text",
				Text: jsonResponse,
			},
		},
	}, nil
}

// GetMatchupsTool returns the MCP tool definition for get_matchups
func (h *LeagueHandler) GetMatchupsTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_matchups",
		Description: "Get matchup information for a specific week including scores and rosters",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"league_id": map[string]interface{}{
					"type":        "string",
					"description": "The Sleeper league ID",
					"required":    true,
				},
				"week": map[string]interface{}{
					"type":        "integer",
					"description": "Week number (1-18)",
					"required":    true,
				},
			},
		},
	}
}

// HandleGetMatchups handles the get_matchups tool call
func (h *LeagueHandler) HandleGetMatchups(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.WithField("args", args).Info("Handling get_matchups")
	
	// Parse arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return nil, fmt.Errorf("league_id is required and must be a string")
	}
	
	weekFloat, ok := args["week"].(float64)
	if !ok {
		return nil, fmt.Errorf("week is required and must be a number")
	}
	week := int(weekFloat)
	
	if week < 1 || week > 18 {
		return nil, fmt.Errorf("week must be between 1 and 18")
	}
	
	// Get matchups from Sleeper API
	matchups, err := h.client.GetMatchups(leagueID, week)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get matchups")
		
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to get matchups: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}
	
	// Create response
	response := sleeper.APIResponse{
		Success: true,
		Data:    matchups,
		Summary: fmt.Sprintf("Found %d matchups for week %d", len(matchups), week),
		Metadata: sleeper.Metadata{
			Timestamp:    time.Now(),
			Source:       "sleeper_api",
			CacheHit:     false,
			APICallsUsed: 1,
			LeagueID:     leagueID,
		},
	}
	
	// Convert to JSON string for MCP response
	jsonResponse, err := formatJSONResponse(response)
	if err != nil {
		h.logger.WithError(err).Error("Failed to format response")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error formatting response: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Type: "text",
				Text: jsonResponse,
			},
		},
	}, nil
}

// getTiebreakOrder determines the effective tiebreaker order based on user input and league settings
func getTiebreakOrder(userTiebreakOrder []string, league *sleeper.League) []string {
	// If user provided custom tiebreaker order, use it
	if len(userTiebreakOrder) > 0 {
		return userTiebreakOrder
	}
	
	// Check league's playoff seed type setting
	if league != nil {
		switch league.Settings.PlayoffSeedType {
		case 0: // Standard: wins, then points for, then points against
			return []string{"wins", "points_for", "points_against"}
		case 1: // Points-based tiebreaker
			return []string{"wins", "points_for"}
		case 2: // Head-to-head record (if supported)
			return []string{"wins", "head_to_head", "points_for", "points_against"}
		default:
			// Fall back to standard
			return []string{"wins", "points_for", "points_against"}
		}
	}
	
	// Default tiebreaker order (Sleeper standard)
	return []string{"wins", "points_for", "points_against"}
}

// sortStandingsWithTiebreakers sorts standings using the specified tiebreaker order
func sortStandingsWithTiebreakers(standings []StandingEntry, tiebreakOrder []string, customMetrics map[string]interface{}, headToHeadMatrix map[int]map[int]int) []StandingEntry {
	// Create a copy to sort
	sortedStandings := make([]StandingEntry, len(standings))
	copy(sortedStandings, standings)
	
	// Group teams by their current tiebreaker values and sort within groups
	return sortStandingsRecursive(sortedStandings, tiebreakOrder, 0, customMetrics, headToHeadMatrix)
}

// sortStandingsRecursive recursively sorts standings by applying tiebreakers level by level
func sortStandingsRecursive(standings []StandingEntry, tiebreakOrder []string, tiebreakerIndex int, customMetrics map[string]interface{}, headToHeadMatrix map[int]map[int]int) []StandingEntry {
	if len(standings) <= 1 || tiebreakerIndex >= len(tiebreakOrder) {
		return standings
	}
	
	currentTiebreaker := tiebreakOrder[tiebreakerIndex]
	
	// Special handling for head-to-head tiebreaker
	if currentTiebreaker == "head_to_head" {
		return sortByHeadToHeadMiniLeague(standings, tiebreakOrder, tiebreakerIndex, customMetrics, headToHeadMatrix)
	}
	
	// Group teams by current tiebreaker value
	groups := groupByTiebreaker(standings, currentTiebreaker, customMetrics)
	
	var result []StandingEntry
	for _, group := range groups {
		// Sort each group by the next tiebreaker
		sortedGroup := sortStandingsRecursive(group, tiebreakOrder, tiebreakerIndex+1, customMetrics, headToHeadMatrix)
		result = append(result, sortedGroup...)
	}
	
	return result
}

// groupByTiebreaker groups teams that are tied on the current tiebreaker
func groupByTiebreaker(standings []StandingEntry, tiebreaker string, customMetrics map[string]interface{}) [][]StandingEntry {
	// Map to group teams by tiebreaker value
	valueGroups := make(map[interface{}][]StandingEntry)
	var orderedValues []interface{}
	
	for _, entry := range standings {
		var value interface{}
		
		switch tiebreaker {
		case "wins":
			value = entry.Wins
		case "points_for":
			value = entry.PointsFor
		case "points_against":
			value = -entry.PointsAgainst // Negative for ascending sort (lower is better)
		case "losses":
			value = -entry.Losses // Negative for ascending sort (fewer losses is better)
		case "custom":
			if customMetrics != nil {
				if val, exists := customMetrics[fmt.Sprintf("roster_%d", entry.RosterID)]; exists {
					value = val
				} else {
					value = 0
				}
			} else {
				value = 0
			}
		case "random":
			// Use random tiebreaker ID for sorting
			value = entry.RandomTiebreakerID
		default:
			value = 0
		}
		
		if _, exists := valueGroups[value]; !exists {
			orderedValues = append(orderedValues, value)
		}
		valueGroups[value] = append(valueGroups[value], entry)
	}
	
	// Sort values in descending order (higher is better for most tiebreakers)
	sort.Slice(orderedValues, func(i, j int) bool {
		return compareValues(orderedValues[i], orderedValues[j]) > 0
	})
	
	// Return groups in sorted order
	var result [][]StandingEntry
	for _, value := range orderedValues {
		result = append(result, valueGroups[value])
	}
	
	return result
}

// sortByHeadToHeadMiniLeague handles the head-to-head tiebreaker using mini-league approach
func sortByHeadToHeadMiniLeague(standings []StandingEntry, tiebreakOrder []string, tiebreakerIndex int, customMetrics map[string]interface{}, headToHeadMatrix map[int]map[int]int) []StandingEntry {
	if headToHeadMatrix == nil {
		// No head-to-head data available, skip to next tiebreaker
		return sortStandingsRecursive(standings, tiebreakOrder, tiebreakerIndex+1, customMetrics, headToHeadMatrix)
	}
	
	// Validate that we have complete head-to-head data for all tied teams
	teamIDs := make([]int, len(standings))
	for i, entry := range standings {
		teamIDs[i] = entry.RosterID
	}
	
	// Check for incomplete matchup data
	if !hasCompleteHeadToHeadData(teamIDs, headToHeadMatrix) {
		// Incomplete data, skip to next tiebreaker
		return sortStandingsRecursive(standings, tiebreakOrder, tiebreakerIndex+1, customMetrics, headToHeadMatrix)
	}
	
	// Calculate mini-league records for all teams
	miniLeagueRecords := make(map[int]MiniLeagueRecord)
	
	for _, entry := range standings {
		miniLeagueRecords[entry.RosterID] = MiniLeagueRecord{
			RosterID: entry.RosterID,
			Wins:     0,
			Losses:   0,
		}
	}
	
	// Calculate head-to-head records within this group (mini-league)
	for _, teamA := range teamIDs {
		for _, teamB := range teamIDs {
			if teamA == teamB {
				continue
			}
			
			// Get wins by teamA against teamB
			winsA := 0
			if teamAData, exists := headToHeadMatrix[teamA]; exists {
				if wins, exists := teamAData[teamB]; exists {
					winsA = wins
				}
			}
			
			// Get wins by teamB against teamA
			winsB := 0
			if teamBData, exists := headToHeadMatrix[teamB]; exists {
				if wins, exists := teamBData[teamA]; exists {
					winsB = wins
				}
			}
			
			// Update mini-league records
			record := miniLeagueRecords[teamA]
			record.Wins += winsA
			record.Losses += winsB
			miniLeagueRecords[teamA] = record
		}
	}
	
	// Group teams by their mini-league records
	recordGroups := make(map[string][]StandingEntry)
	var recordKeys []string
	
	for _, entry := range standings {
		record := miniLeagueRecords[entry.RosterID]
		key := fmt.Sprintf("%d-%d", record.Wins, record.Losses)
		
		if _, exists := recordGroups[key]; !exists {
			recordKeys = append(recordKeys, key)
		}
		recordGroups[key] = append(recordGroups[key], entry)
	}
	
	// Sort record keys by wins (descending), then losses (ascending)
	sort.Slice(recordKeys, func(i, j int) bool {
		var winsI, lossesI, winsJ, lossesJ int
		fmt.Sscanf(recordKeys[i], "%d-%d", &winsI, &lossesI)
		fmt.Sscanf(recordKeys[j], "%d-%d", &winsJ, &lossesJ)
		
		if winsI != winsJ {
			return winsI > winsJ // More wins is better
		}
		return lossesI < lossesJ // Fewer losses is better
	})
	
	// Process each group recursively with the next tiebreaker
	var result []StandingEntry
	for _, key := range recordKeys {
		group := recordGroups[key]
		sortedGroup := sortStandingsRecursive(group, tiebreakOrder, tiebreakerIndex+1, customMetrics, headToHeadMatrix)
		result = append(result, sortedGroup...)
	}
	
	return result
}

// MiniLeagueRecord represents a team's record within a mini-league (tied teams only)
type MiniLeagueRecord struct {
	RosterID int
	Wins     int
	Losses   int
}

// compareValues compares two interface{} values for sorting
func compareValues(a, b interface{}) int {
	switch v1 := a.(type) {
	case int:
		if v2, ok := b.(int); ok {
			if v1 > v2 {
				return 1
			} else if v1 < v2 {
				return -1
			}
			return 0
		}
	case float64:
		if v2, ok := b.(float64); ok {
			if v1 > v2 {
				return 1
			} else if v1 < v2 {
				return -1
			}
			return 0
		}
	}
	return 0
}

// hasCompleteHeadToHeadData validates that all teams have played each other at least once
func hasCompleteHeadToHeadData(teamIDs []int, headToHeadMatrix map[int]map[int]int) bool {
	// For a valid head-to-head tiebreaker, each pair of teams should have played at least once
	for _, teamA := range teamIDs {
		for _, teamB := range teamIDs {
			if teamA == teamB {
				continue
			}
			
			// Check if teamA and teamB have played each other
			winsA := 0
			winsB := 0
			
			if teamAData, exists := headToHeadMatrix[teamA]; exists {
				if wins, exists := teamAData[teamB]; exists {
					winsA = wins
				}
			}
			
			if teamBData, exists := headToHeadMatrix[teamB]; exists {
				if wins, exists := teamBData[teamA]; exists {
					winsB = wins
				}
			}
			
			// If neither team has any wins against the other, they haven't played
			if winsA == 0 && winsB == 0 {
				return false
			}
		}
	}
	
	return true
}

// parseInstructionsToTiebreakOrder extracts tiebreaker order from natural language instructions
func parseInstructionsToTiebreakOrder(instructions string, fallback []string) []string {
	instructions = strings.ToLower(instructions)
	var order []string
	
	// Look for specific tiebreaker mentions in likely order
	patterns := map[string]string{
		"head.to.head|head-to-head|h2h":           "head_to_head",
		"points? for|total points|points scored":  "points_for", 
		"points? against|points allowed":          "points_against",
		"division":                                "division_record",
		"custom":                                  "custom",
	}
	
	// Always start with wins unless explicitly stated otherwise
	if !strings.Contains(instructions, "not wins") && !strings.Contains(instructions, "ignore wins") {
		order = append(order, "wins")
	}
	
	// Find mentions of tiebreakers in order of appearance
	for pattern, tiebreaker := range patterns {
		if matched, _ := regexp.MatchString(pattern, instructions); matched {
			// Avoid duplicates
			exists := false
			for _, existing := range order {
				if existing == tiebreaker {
					exists = true
					break
				}
			}
			if !exists {
				order = append(order, tiebreaker)
			}
		}
	}
	
	// If no specific instructions found, use fallback
	if len(order) <= 1 { // Only wins or nothing
		return fallback
	}
	
	return order
}

// containsTiebreaker checks if a tiebreaker type exists in the order
func containsTiebreaker(tiebreakOrder []string, tiebreaker string) bool {
	for _, tb := range tiebreakOrder {
		if tb == tiebreaker {
			return true
		}
	}
	return false
}

// calculateHeadToHeadMatrix calculates head-to-head records from matchup data
func (h *LeagueHandler) calculateHeadToHeadMatrix(leagueID, mode string) (map[int]map[int]int, error) {
	matrix := make(map[int]map[int]int)
	
	// Determine which weeks to include based on mode
	startWeek := 1
	endWeek := 14 // Standard regular season
	if mode == "final" {
		endWeek = 18 // Include all weeks for final standings
	}
	
	// For each week, get matchups and calculate head-to-head
	for week := startWeek; week <= endWeek; week++ {
		matchups, err := h.client.GetMatchups(leagueID, week)
		if err != nil {
			h.logger.WithError(err).WithField("week", week).Warn("Failed to get matchups for week")
			continue
		}
		
		// Group matchups by matchup_id to find opponents
		matchupGroups := make(map[int][]sleeper.Matchup)
		for _, matchup := range matchups {
			matchupGroups[matchup.MatchupID] = append(matchupGroups[matchup.MatchupID], matchup)
		}
		
		// Process each matchup pair
		for _, matchupPair := range matchupGroups {
			if len(matchupPair) == 2 {
				team1 := matchupPair[0]
				team2 := matchupPair[1]
				
				// Initialize matrix entries if needed
				if matrix[team1.RosterID] == nil {
					matrix[team1.RosterID] = make(map[int]int)
				}
				if matrix[team2.RosterID] == nil {
					matrix[team2.RosterID] = make(map[int]int)
				}
				
				// Determine winner and update head-to-head record
				if team1.Points > team2.Points {
					matrix[team1.RosterID][team2.RosterID]++
				} else if team2.Points > team1.Points {
					matrix[team2.RosterID][team1.RosterID]++
				}
				// Ties don't count as wins for either team
			}
		}
	}
	
	return matrix, nil
}

// processPlayoffBracket analyzes playoff bracket results and determines final standings
func (h *LeagueHandler) processPlayoffBracket(leagueID string, league *sleeper.League, regularSeasonStandings []StandingEntry) (*PlayoffBracket, error) {
	bracket := &PlayoffBracket{
		PlayoffTeams:  make(map[int]int),
		HasThirdPlace: true, // Default assumption
	}
	
	// Determine playoff teams based on league settings and regular season standings
	playoffTeams := 6 // Default
	if league.Settings.PlayoffTeams > 0 {
		playoffTeams = league.Settings.PlayoffTeams
	}
	
	// Map top teams to playoff seeds based on regular season standings
	for i, team := range regularSeasonStandings {
		if i < playoffTeams {
			bracket.PlayoffTeams[team.RosterID] = i + 1 // 1-indexed seeds
		}
	}
	
	// Get authoritative bracket data from Sleeper API
	winnersBracket, err := h.client.GetWinnersBracket(leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get winners bracket: %w", err)
	}
	
	// Process winners bracket to identify championship and other games
	if err := h.processBracketFromAPI(bracket, winnersBracket); err != nil {
		return nil, fmt.Errorf("failed to process bracket from API: %w", err)
	}
	
	return bracket, nil
}

// processBracketFromAPI processes the authoritative bracket data from Sleeper's API
func (h *LeagueHandler) processBracketFromAPI(bracket *PlayoffBracket, winnersBracket []sleeper.BracketMatchup) error {
	// Find the championship game (highest round)
	var championshipGame *sleeper.BracketMatchup
	var thirdPlaceGame *sleeper.BracketMatchup
	maxRound := 0
	
	// Identify the highest round (championship)
	for _, game := range winnersBracket {
		if game.Round > maxRound {
			maxRound = game.Round
		}
	}
	
	// Find championship and third place games in the highest round
	for i, game := range winnersBracket {
		if game.Round == maxRound {
			// In the final round, there are typically 2 games:
			// - One championship game (usually has PlayoffWeek=1 or lower matchup ID)
			// - One third place game (usually has PlayoffWeek=3 or higher matchup ID)
			if game.PlayoffWeek != nil && *game.PlayoffWeek == 1 {
				championshipGame = &winnersBracket[i]
			} else if game.PlayoffWeek != nil && *game.PlayoffWeek == 3 {
				thirdPlaceGame = &winnersBracket[i]
			} else if championshipGame == nil {
				// Fallback: first game in final round is championship
				championshipGame = &winnersBracket[i]
			} else if thirdPlaceGame == nil {
				// Fallback: second game in final round is third place
				thirdPlaceGame = &winnersBracket[i]
			}
		}
	}
	
	if championshipGame == nil {
		return fmt.Errorf("championship game not found in bracket data")
	}
	
	// Convert championship game to our format
	bracket.Championship = &PlayoffMatchup{
		Week:        17, // Assume championship weeks
		MatchupID:   championshipGame.MatchupID,
		Team1:       championshipGame.Team1,
		Team2:       championshipGame.Team2,
		Winner:      championshipGame.Winner,
		Loser:       championshipGame.Loser,
		GameType:    "championship",
	}
	
	// Convert third place game if found
	if thirdPlaceGame != nil {
		bracket.ThirdPlace = &PlayoffMatchup{
			Week:        17, // Assume same weeks as championship
			MatchupID:   thirdPlaceGame.MatchupID,
			Team1:       thirdPlaceGame.Team1,
			Team2:       thirdPlaceGame.Team2,
			Winner:      thirdPlaceGame.Winner,
			Loser:       thirdPlaceGame.Loser,
			GameType:    "third_place",
		}
		bracket.HasThirdPlace = true
	}
	
	// Process other rounds for quarterfinals and semifinals
	for _, game := range winnersBracket {
		playoffMatchup := PlayoffMatchup{
			Week:        15 + game.Round - 1, // Rough mapping: Round 1=Week 15, etc.
			MatchupID:   game.MatchupID,
			Team1:       game.Team1,
			Team2:       game.Team2,
			Winner:      game.Winner,
			Loser:       game.Loser,
		}
		
		switch game.Round {
		case 1:
			playoffMatchup.GameType = "quarterfinal"
			bracket.QuarterFinals = append(bracket.QuarterFinals, playoffMatchup)
		case 2:
			playoffMatchup.GameType = "semifinal"
			bracket.SemiFinals = append(bracket.SemiFinals, playoffMatchup)
		}
	}
	
	h.logger.WithFields(map[string]interface{}{
		"championship_winner": bracket.Championship.Winner,
		"championship_loser":  bracket.Championship.Loser,
		"third_place_winner":  func() int { if bracket.ThirdPlace != nil { return bracket.ThirdPlace.Winner }; return 0 }(),
		"quarterfinals_count": len(bracket.QuarterFinals),
		"semifinals_count":    len(bracket.SemiFinals),
	}).Info("Processed bracket from authoritative Sleeper API data")
	
	return nil
}

// processPlayoffWeek processes all playoff games for a specific week
func (h *LeagueHandler) processPlayoffWeek(leagueID string, bracket *PlayoffBracket, week int, gameType string) error {
	matchups, err := h.client.GetMatchups(leagueID, week)
	if err != nil {
		return fmt.Errorf("failed to get matchups for week %d: %w", week, err)
	}
	
	// Group matchups by matchup_id to find playoff games
	matchupGroups := make(map[int][]sleeper.Matchup)
	for _, matchup := range matchups {
		// Only consider teams that made playoffs
		if _, isPlayoffTeam := bracket.PlayoffTeams[matchup.RosterID]; isPlayoffTeam {
			matchupGroups[matchup.MatchupID] = append(matchupGroups[matchup.MatchupID], matchup)
		}
	}
	
	// Process each playoff matchup
	for matchupID, teams := range matchupGroups {
		if len(teams) != 2 {
			continue // Skip invalid matchups
		}
		
		team1, team2 := teams[0], teams[1]
		
		// Determine winner
		var winner, loser int
		if team1.Points > team2.Points {
			winner, loser = team1.RosterID, team2.RosterID
		} else {
			winner, loser = team2.RosterID, team1.RosterID
		}
		
		playoffMatchup := PlayoffMatchup{
			Week:        week,
			MatchupID:   matchupID,
			Team1:       team1.RosterID,
			Team2:       team2.RosterID,
			Team1Points: team1.Points,
			Team2Points: team2.Points,
			Winner:      winner,
			Loser:       loser,
			GameType:    gameType,
		}
		
		// Classify and store the game
		switch gameType {
		case "quarterfinal":
			bracket.QuarterFinals = append(bracket.QuarterFinals, playoffMatchup)
		case "semifinal":
			bracket.SemiFinals = append(bracket.SemiFinals, playoffMatchup)
		case "championship":
			// Determine if this is championship or third place game
			if h.isChampionshipGame(bracket, team1.RosterID, team2.RosterID) {
				bracket.Championship = &playoffMatchup
				bracket.Championship.GameType = "championship"
			} else {
				bracket.ThirdPlace = &playoffMatchup
				bracket.ThirdPlace.GameType = "third_place"
			}
		}
	}
	
	return nil
}

// isChampionshipGame determines if a matchup is the championship game vs third place game
func (h *LeagueHandler) isChampionshipGame(bracket *PlayoffBracket, team1, team2 int) bool {
	// Championship game is between the two semifinal winners
	if len(bracket.SemiFinals) != 2 {
		return false // Can't determine without complete semifinal data
	}
	
	semifinalWinners := make(map[int]bool)
	for _, semifinal := range bracket.SemiFinals {
		semifinalWinners[semifinal.Winner] = true
	}
	
	// If both teams are semifinal winners, this is the championship
	return semifinalWinners[team1] && semifinalWinners[team2]
}

// calculateFinalStandings calculates final standings based on playoff bracket results
func (h *LeagueHandler) calculateFinalStandings(regularSeasonStandings []StandingEntry, bracket *PlayoffBracket) []StandingEntry {
	finalStandings := make([]StandingEntry, len(regularSeasonStandings))
	
	// Preserve regular season data and add regular season rank
	for i, team := range regularSeasonStandings {
		finalStandings[i] = team
		finalStandings[i].RegularSeasonRank = i + 1
	}
	
	// Assign playoff outcomes and calculate final positions
	h.assignPlayoffOutcomes(finalStandings, bracket)
	
	// Sort by final standings rules
	h.sortFinalStandings(finalStandings)
	
	// Update final ranks
	for i := range finalStandings {
		finalStandings[i].Rank = i + 1
	}
	
	return finalStandings
}

// assignPlayoffOutcomes assigns playoff outcomes to teams based on bracket results
func (h *LeagueHandler) assignPlayoffOutcomes(standings []StandingEntry, bracket *PlayoffBracket) {
	// Create roster ID to standing entry map for quick lookup
	standingMap := make(map[int]*StandingEntry)
	for i := range standings {
		standingMap[standings[i].RosterID] = &standings[i]
	}
	
	// Assign outcomes based on playoff results
	if bracket.Championship != nil {
		if entry, exists := standingMap[bracket.Championship.Winner]; exists {
			entry.PlayoffOutcome = "champion"
		}
		if entry, exists := standingMap[bracket.Championship.Loser]; exists {
			entry.PlayoffOutcome = "runner_up"
		}
	}
	
	if bracket.ThirdPlace != nil {
		if entry, exists := standingMap[bracket.ThirdPlace.Winner]; exists {
			entry.PlayoffOutcome = "third_place"
		}
		if entry, exists := standingMap[bracket.ThirdPlace.Loser]; exists {
			entry.PlayoffOutcome = "fourth_place"
		}
	}
	
	// Assign quarterfinal losers
	quarterfinalLosers := make(map[int]bool)
	for _, qf := range bracket.QuarterFinals {
		quarterfinalLosers[qf.Loser] = true
	}
	
	for rosterID := range quarterfinalLosers {
		if entry, exists := standingMap[rosterID]; exists {
			entry.PlayoffOutcome = "quarterfinal_loss"
		}
	}
	
	// Assign non-playoff teams
	for i := range standings {
		if standings[i].PlayoffOutcome == "" {
			if _, madePlayoffs := bracket.PlayoffTeams[standings[i].RosterID]; !madePlayoffs {
				standings[i].PlayoffOutcome = "no_playoffs"
			}
		}
	}
}

// sortFinalStandings sorts standings according to final standings rules
func (h *LeagueHandler) sortFinalStandings(standings []StandingEntry) {
	sort.Slice(standings, func(i, j int) bool {
		return h.shouldSwapFinalStandings(standings[i], standings[j])
	})
}

// shouldSwapFinalStandings determines final standings order based on playoff outcomes
func (h *LeagueHandler) shouldSwapFinalStandings(a, b StandingEntry) bool {
	// Define playoff outcome priority (lower number = better finish)
	outcomePriority := map[string]int{
		"champion":          1,
		"runner_up":         2,
		"third_place":       3,
		"fourth_place":      4,
		"quarterfinal_loss": 5,
		"no_playoffs":       6,
	}
	
	priorityA := outcomePriority[a.PlayoffOutcome]
	priorityB := outcomePriority[b.PlayoffOutcome]
	
	// If different playoff outcomes, use playoff outcome priority
	if priorityA != priorityB {
		return priorityA > priorityB // Higher priority (worse outcome) should come later
	}
	
	// Same playoff outcome - use regular season tiebreakers
	switch a.PlayoffOutcome {
	case "quarterfinal_loss":
		// Quarterfinal losers ranked by regular season record
		return a.RegularSeasonRank > b.RegularSeasonRank
	case "no_playoffs":
		// Non-playoff teams ranked by regular season record
		return a.RegularSeasonRank > b.RegularSeasonRank
	default:
		// Championship, runner-up, third place, fourth place are determined by games
		// No additional sorting needed within these categories
		return false
	}
}

// validatePlayoffBracket validates that the playoff bracket has complete and consistent data
func (h *LeagueHandler) validatePlayoffBracket(bracket *PlayoffBracket) error {
	if bracket == nil {
		return fmt.Errorf("bracket is nil")
	}
	
	// Check if we have the expected number of playoff teams (6 for standard format)
	expectedPlayoffTeams := 6
	if len(bracket.PlayoffTeams) != expectedPlayoffTeams {
		return fmt.Errorf("expected %d playoff teams, got %d", expectedPlayoffTeams, len(bracket.PlayoffTeams))
	}
	
	// For a complete bracket, we should have:
	// - 2 quarterfinal games (seeds 3v6, 4v5)
	// - 2 semifinal games
	// - 1 championship game
	// - 1 third place game
	
	if len(bracket.QuarterFinals) < 2 {
		return fmt.Errorf("incomplete quarterfinals: expected 2 games, got %d", len(bracket.QuarterFinals))
	}
	
	if len(bracket.SemiFinals) < 2 {
		return fmt.Errorf("incomplete semifinals: expected 2 games, got %d", len(bracket.SemiFinals))
	}
	
	if bracket.Championship == nil {
		return fmt.Errorf("championship game not found")
	}
	
	if bracket.ThirdPlace == nil && bracket.HasThirdPlace {
		return fmt.Errorf("third place game not found")
	}
	
	// For 2-week championship format, we don't validate semifinal winners in championship
	// because all 4 semifinal participants compete in the 4-team finals
	if bracket.ChampionshipWeek != 17 || !bracket.HasThirdPlace {
		// Only validate traditional format
		semifinalWinners := make([]int, 0, 2)
		for _, semifinal := range bracket.SemiFinals {
			semifinalWinners = append(semifinalWinners, semifinal.Winner)
		}
		
		championshipParticipants := []int{bracket.Championship.Team1, bracket.Championship.Team2}
		if !containsAllTeams(championshipParticipants, semifinalWinners) {
			return fmt.Errorf("championship game participants don't match semifinal winners")
		}
	}
	
	return nil
}

// containsAllTeams checks if all teams in the expected list are present in the actual list
func containsAllTeams(actual, expected []int) bool {
	if len(actual) != len(expected) {
		return false
	}
	
	actualMap := make(map[int]bool)
	for _, team := range actual {
		actualMap[team] = true
	}
	
	for _, team := range expected {
		if !actualMap[team] {
			return false
		}
	}
	
	return true
}

// detectPlayoffStructure analyzes matchup patterns to determine playoff structure
func (h *LeagueHandler) detectPlayoffStructure(leagueID string, playoffTeams map[int]int) (*PlayoffStructure, error) {
	structure := &PlayoffStructure{
		Format: "6team", // Default assumption
	}
	
	// Analyze weeks 15-18 to detect playoff patterns
	weekAnalysis := make(map[int]int) // week -> number of playoff teams playing
	
	for week := 15; week <= 18; week++ {
		matchups, err := h.client.GetMatchups(leagueID, week)
		if err != nil {
			h.logger.WithField("week", week).Warn("Failed to get matchups for playoff detection")
			continue
		}
		
		// Count playoff teams in actual playoff bracket games (not consolation)
		// Group by matchup to find games between playoff teams
		matchupGroups := make(map[int][]sleeper.Matchup)
		for _, matchup := range matchups {
			matchupGroups[matchup.MatchupID] = append(matchupGroups[matchup.MatchupID], matchup)
		}
		
		playoffTeamsInBracket := make(map[int]bool)
		for _, teams := range matchupGroups {
			if len(teams) != 2 {
				continue
			}
			
			team1, team2 := teams[0], teams[1]
			_, isPlayoff1 := playoffTeams[team1.RosterID]
			_, isPlayoff2 := playoffTeams[team2.RosterID]
			
			// Only count if BOTH teams are playoff teams (bracket game, not consolation)
			if isPlayoff1 && isPlayoff2 {
				playoffTeamsInBracket[team1.RosterID] = true
				playoffTeamsInBracket[team2.RosterID] = true
			}
		}
		
		weekAnalysis[week] = len(playoffTeamsInBracket)
		h.logger.WithFields(map[string]interface{}{
			"week": week,
			"playoff_teams_in_bracket": len(playoffTeamsInBracket),
		}).Info("Playoff week analysis")
	}
	
	// Determine playoff structure based on patterns
	// Check for different playoff formats:
	
	// Pattern 1: Standard format (4 teams QF, 4 teams SF, 2 teams championship)
	// Pattern 2: 6-team format with byes (4 teams QF, 4 teams SF, 2 teams in weeks 17-18)
	
	// Look for 4 teams playing in consecutive weeks (QF then SF)
	if weekAnalysis[15] == 4 && weekAnalysis[16] == 4 {
		structure.QuarterfinalsWeek = 15
		structure.SemifinalsWeek = 16
		
		// For 2-week championship, need to check manually for semifinal winners
		// since the bracket logic changes after semifinals
		if weekAnalysis[17] == 0 && weekAnalysis[18] == 0 {
			// This suggests 2-week championship format
			// Verify by checking if we can find the two semifinal winners playing in weeks 17-18
			if h.verifyTwoWeekChampionship(leagueID, playoffTeams) {
				structure.ChampionshipWeek = 17
				structure.ThirdPlaceWeek = 18
				structure.HasThirdPlace = false // 2-week championship, no separate 3rd place
			}
		} else if weekAnalysis[17] == 2 {
			structure.ChampionshipWeek = 17
			structure.ThirdPlaceWeek = 17
			structure.HasThirdPlace = true
		}
	} else {
		// Try to detect other patterns
		for week, teamsPlaying := range weekAnalysis {
			switch teamsPlaying {
			case 4:
				if structure.QuarterfinalsWeek == 0 {
					structure.QuarterfinalsWeek = week
				} else if structure.SemifinalsWeek == 0 {
					structure.SemifinalsWeek = week
				}
			case 2:
				if structure.ChampionshipWeek == 0 {
					structure.ChampionshipWeek = week
					structure.ThirdPlaceWeek = week
					structure.HasThirdPlace = true
				}
			}
		}
	}
	
	// Validate detected structure
	if structure.ChampionshipWeek == 0 {
		return nil, fmt.Errorf("could not detect championship week")
	}
	
	h.logger.WithFields(map[string]interface{}{
		"quarterfinals_week": structure.QuarterfinalsWeek,
		"semifinals_week":    structure.SemifinalsWeek,
		"championship_week":  structure.ChampionshipWeek,
		"third_place_week":   structure.ThirdPlaceWeek,
		"format":             "6team_with_2week_championship",
	}).Info("Detected playoff structure")
	
	return structure, nil
}

// processPlayoffWeekImproved processes playoff games with improved team detection
func (h *LeagueHandler) processPlayoffWeekImproved(leagueID string, bracket *PlayoffBracket, week int, gameType string) error {
	matchups, err := h.client.GetMatchups(leagueID, week)
	if err != nil {
		return fmt.Errorf("failed to get matchups for week %d: %w", week, err)
	}
	
	// Group matchups by matchup_id
	matchupGroups := make(map[int][]sleeper.Matchup)
	for _, matchup := range matchups {
		matchupGroups[matchup.MatchupID] = append(matchupGroups[matchup.MatchupID], matchup)
	}
	
	// Process each matchup group
	for matchupID, teams := range matchupGroups {
		if len(teams) != 2 {
			continue // Skip invalid matchups
		}
		
		team1, team2 := teams[0], teams[1]
		
		// Check if both teams are playoff teams
		_, team1IsPlayoff := bracket.PlayoffTeams[team1.RosterID]
		_, team2IsPlayoff := bracket.PlayoffTeams[team2.RosterID]
		
		if !team1IsPlayoff || !team2IsPlayoff {
			// This might be a consolation game or regular season game
			h.logger.WithFields(map[string]interface{}{
				"week":       week,
				"matchup_id": matchupID,
				"team1":      team1.RosterID,
				"team2":      team2.RosterID,
			}).Debug("Skipping non-playoff matchup")
			continue
		}
		
		// Determine winner
		var winner, loser int
		if team1.Points > team2.Points {
			winner, loser = team1.RosterID, team2.RosterID
		} else {
			winner, loser = team2.RosterID, team1.RosterID
		}
		
		playoffMatchup := PlayoffMatchup{
			Week:        week,
			MatchupID:   matchupID,
			Team1:       team1.RosterID,
			Team2:       team2.RosterID,
			Team1Points: team1.Points,
			Team2Points: team2.Points,
			Winner:      winner,
			Loser:       loser,
			GameType:    gameType,
		}
		
		// Store the game in the appropriate bracket section
		switch gameType {
		case "quarterfinal":
			bracket.QuarterFinals = append(bracket.QuarterFinals, playoffMatchup)
		case "semifinal":
			bracket.SemiFinals = append(bracket.SemiFinals, playoffMatchup)
		case "championship":
			// For championship week, we need to distinguish between championship and third place
			if h.isChampionshipGameImproved(bracket, team1.RosterID, team2.RosterID) {
				bracket.Championship = &playoffMatchup
				bracket.Championship.GameType = "championship"
			} else {
				bracket.ThirdPlace = &playoffMatchup
				bracket.ThirdPlace.GameType = "third_place"
			}
		}
		
		h.logger.WithFields(map[string]interface{}{
			"week":      week,
			"game_type": gameType,
			"winner":    winner,
			"loser":     loser,
		}).Info("Processed playoff game")
	}
	
	return nil
}

// isChampionshipGameImproved determines if a matchup is the championship game vs third place game
func (h *LeagueHandler) isChampionshipGameImproved(bracket *PlayoffBracket, team1, team2 int) bool {
	// If we have semifinal data, use it to determine championship participants
	if len(bracket.SemiFinals) >= 2 {
		semifinalWinners := make(map[int]bool)
		for _, semifinal := range bracket.SemiFinals {
			semifinalWinners[semifinal.Winner] = true
		}
		
		// If both teams are semifinal winners, this is the championship
		return semifinalWinners[team1] && semifinalWinners[team2]
	}
	
	// Fallback: assume the matchup with higher-seeded teams is the championship
	seed1, exists1 := bracket.PlayoffTeams[team1]
	seed2, exists2 := bracket.PlayoffTeams[team2]
	
	if exists1 && exists2 {
		// Lower seed numbers = higher seeds (1 is better than 2)
		// Championship likely involves the top seeds
		averageSeed := (seed1 + seed2) / 2
		return averageSeed <= 2 // Top 2 average seeds likely championship
	}
	
	// Default to championship if we can't determine
	return true
}

// processTwoWeekChampionship handles 2-week championship format (weeks 17-18)
// In this format: semifinal winners play H2H for 2 weeks, semifinal losers play H2H for 3rd place
func (h *LeagueHandler) processTwoWeekChampionship(leagueID string, bracket *PlayoffBracket) error {
	if len(bracket.SemiFinals) < 2 {
		return fmt.Errorf("need semifinal results to process championship")
	}
	
	// Get semifinal winners and losers
	semifinalWinners := make([]int, 0, 2)
	semifinalLosers := make([]int, 0, 2)
	for _, semifinal := range bracket.SemiFinals {
		semifinalWinners = append(semifinalWinners, semifinal.Winner)
		semifinalLosers = append(semifinalLosers, semifinal.Loser)
	}
	
	if len(semifinalWinners) != 2 {
		return fmt.Errorf("expected 2 semifinal winners, got %d", len(semifinalWinners))
	}
	if len(semifinalLosers) != 2 {
		return fmt.Errorf("expected 2 semifinal losers, got %d", len(semifinalLosers))
	}
	
	// Track 2-week totals for championship game (semifinal winners)
	championshipTotals := make(map[int]float64)
	thirdPlaceTotals := make(map[int]float64)
	
	// Process weeks 17 and 18
	for week := 17; week <= 18; week++ {
		matchups, err := h.client.GetMatchups(leagueID, week)
		if err != nil {
			return fmt.Errorf("failed to get matchups for week %d: %w", week, err)
		}
		
		// Find championship participants (semifinal winners) and third place participants (semifinal losers)
		for _, matchup := range matchups {
			if contains(semifinalWinners, matchup.RosterID) {
				championshipTotals[matchup.RosterID] += matchup.Points
			} else if contains(semifinalLosers, matchup.RosterID) {
				thirdPlaceTotals[matchup.RosterID] += matchup.Points
			}
		}
	}
	
	// Determine championship winner (higher 2-week total between semifinal winners)
	var champion, runnerUp int
	var championTotal, runnerUpTotal float64
	
	for rosterID, total := range championshipTotals {
		if total > championTotal {
			runnerUp = champion
			runnerUpTotal = championTotal
			champion = rosterID
			championTotal = total
		} else if total > runnerUpTotal {
			runnerUp = rosterID
			runnerUpTotal = total
		}
	}
	
	if champion == 0 || runnerUp == 0 {
		return fmt.Errorf("could not determine championship results from semifinal winners")
	}
	
	// Determine third place winner (higher 2-week total between semifinal losers)
	var thirdPlace, fourthPlace int
	var thirdPlaceTotal, fourthPlaceTotal float64
	
	for rosterID, total := range thirdPlaceTotals {
		if total > thirdPlaceTotal {
			fourthPlace = thirdPlace
			fourthPlaceTotal = thirdPlaceTotal
			thirdPlace = rosterID
			thirdPlaceTotal = total
		} else if total > fourthPlaceTotal {
			fourthPlace = rosterID
			fourthPlaceTotal = total
		}
	}
	
	if thirdPlace == 0 || fourthPlace == 0 {
		return fmt.Errorf("could not determine third place results from semifinal losers")
	}
	
	// Create championship result (2-week head-to-head between semifinal winners)
	bracket.Championship = &PlayoffMatchup{
		Week:        17, // Start week
		MatchupID:   1,  // Combined matchup
		Team1:       champion,
		Team2:       runnerUp,
		Team1Points: championTotal,
		Team2Points: runnerUpTotal,
		Winner:      champion,
		Loser:       runnerUp,
		GameType:    "championship",
	}
	
	// Create third place result (2-week head-to-head between semifinal losers)
	bracket.ThirdPlace = &PlayoffMatchup{
		Week:        17, // Start week
		MatchupID:   2,  // Different from championship
		Team1:       thirdPlace,
		Team2:       fourthPlace,
		Team1Points: thirdPlaceTotal,
		Team2Points: fourthPlaceTotal,
		Winner:      thirdPlace,
		Loser:       fourthPlace,
		GameType:    "third_place",
	}
	bracket.HasThirdPlace = true
	
	h.logger.WithFields(map[string]interface{}{
		"champion":           champion,
		"champion_total":     championTotal,
		"runner_up":          runnerUp,
		"runner_up_total":    runnerUpTotal,
		"third_place":        thirdPlace,
		"third_place_total":  thirdPlaceTotal,
		"fourth_place":       fourthPlace,
		"fourth_place_total": fourthPlaceTotal,
	}).Info("Processed 2-week championship (winners H2H) and third place game (losers H2H)")
	
	return nil
}

// contains checks if a slice contains a specific item
func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// verifyTwoWeekChampionship checks if weeks 17-18 contain a 2-week championship
func (h *LeagueHandler) verifyTwoWeekChampionship(leagueID string, playoffTeams map[int]int) bool {
	// Check if the same 2 teams are playing in both weeks 17 and 18
	var week17Teams, week18Teams []int
	
	for week := 17; week <= 18; week++ {
		matchups, err := h.client.GetMatchups(leagueID, week)
		if err != nil {
			h.logger.WithField("week", week).Warn("Failed to get matchups for championship verification")
			continue
		}
		
		teamsThisWeek := make(map[int]bool)
		for _, matchup := range matchups {
			// Look for any teams that could be playing championship
			// (not necessarily in original top 6, could be semifinal winners)
			teamsThisWeek[matchup.RosterID] = true
		}
		
		var teamList []int
		for teamID := range teamsThisWeek {
			teamList = append(teamList, teamID)
		}
		
		if week == 17 {
			week17Teams = teamList
		} else {
			week18Teams = teamList
		}
	}
	
	// Check if we have exactly 2 teams playing in both weeks
	if len(week17Teams) >= 2 && len(week18Teams) >= 2 {
		// Find common teams between weeks 17 and 18
		commonTeams := 0
		for _, team17 := range week17Teams {
			for _, team18 := range week18Teams {
				if team17 == team18 {
					commonTeams++
					break
				}
			}
		}
		
		// If we have at least 2 common teams, this is likely a 2-week championship
		if commonTeams >= 2 {
			h.logger.WithFields(map[string]interface{}{
				"week_17_teams": week17Teams,
				"week_18_teams": week18Teams,
				"common_teams":  commonTeams,
			}).Info("Verified 2-week championship format")
			return true
		}
	}
	
	return false
}

// generateRandomTiebreakerID generates a consistent random ID for tiebreaker purposes
func generateRandomTiebreakerID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("tb_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("tb_%x", bytes)
}

// applyRandomTiebreaker assigns random tiebreaker IDs to teams that need them
func (h *LeagueHandler) applyRandomTiebreaker(standings []StandingEntry, tiebreakOrder []string) {
	// Only apply random tiebreaker if it's in the tiebreak order
	hasRandomTiebreaker := false
	for _, tb := range tiebreakOrder {
		if tb == "random" {
			hasRandomTiebreaker = true
			break
		}
	}
	
	if !hasRandomTiebreaker {
		return
	}
	
	// Assign random IDs to teams that don't have them
	for i := range standings {
		if standings[i].RandomTiebreakerID == "" {
			standings[i].RandomTiebreakerID = generateRandomTiebreakerID()
		}
	}
}