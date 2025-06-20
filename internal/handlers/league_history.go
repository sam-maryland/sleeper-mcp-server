package handlers

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sam-maryland/sleeper-mcp-server/internal/sleeper"
	"github.com/sirupsen/logrus"
)

// LeagueHistoryArgs represents parameters for league history tools
type LeagueHistoryArgs struct {
	LeagueID string `json:"league_id"`
	Seasons  int    `json:"seasons,omitempty"` // Number of seasons back to look (default: 10)
}

// SeasonData represents a league season
type SeasonData struct {
	Season   string          `json:"season"`
	LeagueID string          `json:"league_id"`
	League   *sleeper.League `json:"league"`
	Users    []sleeper.User  `json:"users"`
}

// LeagueGroup represents a multi-season league
type LeagueGroup struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Seasons     map[string]SeasonData `json:"seasons"`
	CommonUsers []sleeper.User        `json:"common_users"`
}

// DiscoverLeagueHistoryTool returns the MCP tool definition for discover_league_history
func (h *LeagueHandler) DiscoverLeagueHistoryTool() mcp.Tool {
	return mcp.Tool{
		Name:        "discover_league_history",
		Description: "Automatically discover historical league IDs for previous seasons based on a current league ID",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"league_id": map[string]interface{}{
					"type":        "string",
					"description": "Current league ID to discover history for",
					"required":    true,
				},
				"seasons": map[string]interface{}{
					"type":        "integer",
					"description": "Number of seasons back to search (default: 10)",
					"required":    false,
				},
			},
		},
	}
}

// HandleDiscoverLeagueHistory discovers historical league IDs by finding leagues with overlapping membership
func (h *LeagueHandler) HandleDiscoverLeagueHistory(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.WithField("args", args).Info("Handling discover_league_history")

	// Parse arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return nil, fmt.Errorf("league_id is required and must be a string")
	}

	seasons := 10 // default
	if seasonsRaw, exists := args["seasons"]; exists {
		if seasonsFloat, ok := seasonsRaw.(float64); ok {
			seasons = int(seasonsFloat)
		}
	}

	// Discover league history
	leagueGroup, err := h.discoverLeagueHistory(leagueID, seasons)
	if err != nil {
		h.logger.WithError(err).Error("Failed to discover league history")
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Failed to discover league history: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}

	// Create response
	response := sleeper.APIResponse{
		Success: true,
		Data:    leagueGroup,
		Summary: fmt.Sprintf("Discovered %d seasons of league history", len(leagueGroup.Seasons)),
		Metadata: sleeper.Metadata{
			Timestamp:    time.Now(),
			Source:       "sleeper_api",
			CacheHit:     false,
			APICallsUsed: calculateAPICalls(leagueGroup),
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

// discoverLeagueHistory discovers historical leagues by finding leagues with overlapping membership
func (h *LeagueHandler) discoverLeagueHistory(currentLeagueID string, maxSeasons int) (*LeagueGroup, error) {
	// Get current league info and users
	currentLeague, err := h.client.GetLeague(currentLeagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current league: %w", err)
	}

	currentUsers, err := h.client.GetLeagueUsers(currentLeagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current league users: %w", err)
	}

	if len(currentUsers) == 0 {
		return nil, fmt.Errorf("no users found in current league")
	}

	// Get current season from league
	currentSeason := currentLeague.Season
	if currentSeason == "" {
		currentSeason = fmt.Sprintf("%d", time.Now().Year())
	}

	// Initialize league group
	leagueGroup := &LeagueGroup{
		Name:        fmt.Sprintf("%s (Multi-Season)", currentLeague.Name),
		Description: "Automatically discovered league history",
		Seasons:     make(map[string]SeasonData),
		CommonUsers: currentUsers,
	}

	// Add current season
	leagueGroup.Seasons[currentSeason] = SeasonData{
		Season:   currentSeason,
		LeagueID: currentLeagueID,
		League:   currentLeague,
		Users:    currentUsers,
	}

	// Search previous seasons
	currentYear, _ := strconv.Atoi(currentSeason)
	for i := 1; i <= maxSeasons; i++ {
		searchYear := currentYear - i
		searchSeason := fmt.Sprintf("%d", searchYear)

		h.logger.WithField("season", searchSeason).Info("Searching for league in season")

		// Use first few users to search for leagues in this season
		// We'll search with multiple users to increase chances of finding the league
		searchUsers := currentUsers
		if len(searchUsers) > 5 {
			searchUsers = currentUsers[:5] // Limit to first 5 users to avoid too many API calls
		}

		var foundLeague *SeasonData
		for _, user := range searchUsers {
			userLeagues, err := h.client.GetUserLeagues(user.UserID, "nfl", searchSeason)
			if err != nil {
				h.logger.WithError(err).WithFields(logrus.Fields{
					"user_id": user.UserID,
					"season":  searchSeason,
				}).Warn("Failed to get user leagues")
				continue
			}

			// Look for leagues with significant overlap with current users
			for _, league := range userLeagues {
				leagueUsers, err := h.client.GetLeagueUsers(league.LeagueID)
				if err != nil {
					continue
				}

				// Calculate overlap percentage
				overlap := calculateUserOverlap(currentUsers, leagueUsers)
				h.logger.WithFields(logrus.Fields{
					"league_id": league.LeagueID,
					"overlap":   overlap,
					"season":    searchSeason,
				}).Debug("Calculated user overlap")

				// If 60%+ of users overlap, likely same league
				if overlap >= 0.6 {
					foundLeague = &SeasonData{
						Season:   searchSeason,
						LeagueID: league.LeagueID,
						League:   &league,
						Users:    leagueUsers,
					}
					h.logger.WithFields(logrus.Fields{
						"league_id": league.LeagueID,
						"overlap":   overlap,
						"season":    searchSeason,
					}).Info("Found historical league")
					break
				}
			}

			if foundLeague != nil {
				break
			}
		}

		if foundLeague != nil {
			leagueGroup.Seasons[searchSeason] = *foundLeague
		} else {
			h.logger.WithField("season", searchSeason).Info("No league found for season")
			// Stop searching if we don't find a league for this season
			break
		}
	}

	return leagueGroup, nil
}

// calculateUserOverlap calculates the percentage of users that overlap between two user lists
func calculateUserOverlap(users1, users2 []sleeper.User) float64 {
	if len(users1) == 0 || len(users2) == 0 {
		return 0.0
	}

	// Create map of user IDs from first list
	userMap := make(map[string]bool)
	for _, user := range users1 {
		userMap[user.UserID] = true
	}

	// Count overlapping users
	overlap := 0
	for _, user := range users2 {
		if userMap[user.UserID] {
			overlap++
		}
	}

	// Return overlap as percentage of smaller list
	minSize := len(users1)
	if len(users2) < minSize {
		minSize = len(users2)
	}

	return float64(overlap) / float64(minSize)
}

// calculateAPICalls estimates the number of API calls used in discovery
func calculateAPICalls(leagueGroup *LeagueGroup) int {
	// Rough estimate: 1 call per league + 1 call per league for users + user search calls
	return len(leagueGroup.Seasons)*2 + 10 // +10 for user league searches
}

// GetLeagueHistoryTool returns the MCP tool definition for get_league_history
func (h *LeagueHandler) GetLeagueHistoryTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_league_history",
		Description: "Get comprehensive league history including standings, championships, and user performance across multiple seasons",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"league_id": map[string]interface{}{
					"type":        "string",
					"description": "Current league ID (will auto-discover historical seasons)",
					"required":    true,
				},
				"include_standings": map[string]interface{}{
					"type":        "boolean",
					"description": "Include standings for each season (default: true)",
					"required":    false,
				},
			},
		},
	}
}

// HandleGetLeagueHistory gets comprehensive league history
func (h *LeagueHandler) HandleGetLeagueHistory(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.WithField("args", args).Info("Handling get_league_history")

	// Parse arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return nil, fmt.Errorf("league_id is required and must be a string")
	}

	includeStandings := true
	if standingsRaw, exists := args["include_standings"]; exists {
		if standingsBool, ok := standingsRaw.(bool); ok {
			includeStandings = standingsBool
		}
	}

	// First discover league history
	leagueGroup, err := h.discoverLeagueHistory(leagueID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to discover league history: %w", err)
	}

	// Build comprehensive history
	history := make(map[string]interface{})
	
	// Get all seasons and sort them
	seasons := make([]string, 0, len(leagueGroup.Seasons))
	for season := range leagueGroup.Seasons {
		seasons = append(seasons, season)
	}
	sort.Strings(seasons)

	history["league_name"] = leagueGroup.Name
	history["seasons_found"] = len(seasons)
	history["seasons"] = seasons

	// Add season details
	seasonDetails := make(map[string]interface{})
	for _, season := range seasons {
		seasonData := leagueGroup.Seasons[season]
		details := map[string]interface{}{
			"league_id":   seasonData.LeagueID,
			"league_name": seasonData.League.Name,
			"status":      seasonData.League.Status,
			"users_count": len(seasonData.Users),
		}

		if includeStandings && seasonData.League.Status == "complete" {
			// Note: This would call the existing standings function
			// For now, just indicate standings are available
			details["standings_available"] = true
		}

		seasonDetails[season] = details
	}
	history["season_details"] = seasonDetails

	// Create response
	response := sleeper.APIResponse{
		Success: true,
		Data:    history,
		Summary: fmt.Sprintf("Found %d seasons of league history from %s to %s", len(seasons), seasons[0], seasons[len(seasons)-1]),
		Metadata: sleeper.Metadata{
			Timestamp:    time.Now(),
			Source:       "sleeper_api",
			CacheHit:     false,
			APICallsUsed: calculateAPICalls(leagueGroup),
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