package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
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
	RosterID        int                    `json:"roster_id"`
	OwnerID         string                 `json:"owner_id"`
	DisplayName     string                 `json:"display_name"`
	TeamName        string                 `json:"team_name"`
	Wins            int                    `json:"wins"`
	Losses          int                    `json:"losses"`
	Ties            int                    `json:"ties"`
	PointsFor       float64                `json:"points_for"`
	PointsAgainst   float64                `json:"points_against"`
	Rank            int                    `json:"rank"`
	Division        int                    `json:"division,omitempty"`
	PlayoffSeed     int                    `json:"playoff_seed,omitempty"`
	CustomMetrics   map[string]interface{} `json:"custom_metrics,omitempty"`
	TiebreakerNotes string                 `json:"tiebreaker_notes,omitempty"`
	HeadToHeadWins  map[int]int            `json:"head_to_head_wins,omitempty"`
}

// LeagueHandler handles league-related MCP tools
type LeagueHandler struct {
	client sleeper.Client
	logger *logrus.Logger
}

// NewLeagueHandler creates a new league handler
func NewLeagueHandler(client sleeper.Client, logger *logrus.Logger) *LeagueHandler {
	return &LeagueHandler{
		client: client,
		logger: logger,
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
	
	// Sort standings using flexible tiebreaker system
	standings = sortStandingsWithTiebreakers(standings, effectiveTiebreakOrder, customMetrics, headToHeadMatrix)
	
	// Add notes about tiebreakers used
	tiebreakerNotes := fmt.Sprintf("Tiebreakers applied: %v", effectiveTiebreakOrder)
	if league != nil && league.Settings.PlayoffSeedType != 0 {
		tiebreakerNotes += fmt.Sprintf(" (League playoff_seed_type: %d)", league.Settings.PlayoffSeedType)
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
	
	// Sort using bubble sort with custom comparison
	for i := 0; i < len(sortedStandings); i++ {
		for j := i + 1; j < len(sortedStandings); j++ {
			if shouldSwap(sortedStandings[i], sortedStandings[j], tiebreakOrder, customMetrics, headToHeadMatrix) {
				sortedStandings[i], sortedStandings[j] = sortedStandings[j], sortedStandings[i]
			}
		}
	}
	
	return sortedStandings
}

// shouldSwap determines if two standings entries should be swapped based on tiebreaker rules
func shouldSwap(a, b StandingEntry, tiebreakOrder []string, customMetrics map[string]interface{}, headToHeadMatrix map[int]map[int]int) bool {
	for _, tiebreaker := range tiebreakOrder {
		switch tiebreaker {
		case "wins":
			if a.Wins != b.Wins {
				return a.Wins < b.Wins // Higher wins should come first
			}
		case "points_for":
			if a.PointsFor != b.PointsFor {
				return a.PointsFor < b.PointsFor // Higher points should come first
			}
		case "points_against":
			if a.PointsAgainst != b.PointsAgainst {
				return a.PointsAgainst > b.PointsAgainst // Lower points against should come first
			}
		case "division_record":
			// This would require calculating division records from matchup data
			// For now, fall back to next tiebreaker
			continue
		case "head_to_head":
			if headToHeadMatrix != nil && a.HeadToHeadWins != nil && b.HeadToHeadWins != nil {
				// Calculate head-to-head wins for each team against the other
				aH2HWins := 0
				bH2HWins := 0
				if wins, exists := a.HeadToHeadWins[b.RosterID]; exists {
					aH2HWins = wins
				}
				if wins, exists := b.HeadToHeadWins[a.RosterID]; exists {
					bH2HWins = wins
				}
				
				
				if aH2HWins != bH2HWins {
					return aH2HWins < bH2HWins // Higher head-to-head wins should come first
				}
			}
			continue
		case "custom":
			// Handle custom metrics if provided
			if customMetrics != nil {
				// This is where custom tiebreaker logic would be applied
				// Implementation depends on the specific custom metrics provided
			}
			continue
		}
	}
	
	// If all tiebreakers are equal, maintain current order
	return false
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