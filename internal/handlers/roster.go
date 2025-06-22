package handlers

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sam-maryland/sleeper-mcp-server/internal/sleeper"
	"github.com/sirupsen/logrus"
)

// RosterArgs represents the parameters for the get_roster tool
type RosterArgs struct {
	LeagueID string `json:"league_id"`
	RosterID int    `json:"roster_id"`
}

// AllRostersArgs represents the parameters for the get_all_rosters tool
type AllRostersArgs struct {
	LeagueID string `json:"league_id"`
}

// AnalyzeRosterArgs represents the parameters for the analyze_roster_strength tool
type AnalyzeRosterArgs struct {
	LeagueID string `json:"league_id"`
	RosterID int    `json:"roster_id"`
}

// CompareRostersArgs represents the parameters for the compare_rosters tool
type CompareRostersArgs struct {
	LeagueID  string `json:"league_id"`
	RosterID1 int    `json:"roster_id1"`
	RosterID2 int    `json:"roster_id2"`
}

// RosterHandler handles roster-related MCP tools
type RosterHandler struct {
	client sleeper.Client
	logger *logrus.Logger
}

// NewRosterHandler creates a new roster handler
func NewRosterHandler(client sleeper.Client, logger *logrus.Logger) *RosterHandler {
	return &RosterHandler{
		client: client,
		logger: logger,
	}
}

// GetRosterTool returns the get_roster MCP tool definition
func (h *RosterHandler) GetRosterTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_roster",
		Description: "Get detailed information about a specific team's roster including all players, starters, and bench players",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"league_id": map[string]interface{}{
					"type":        "string",
					"description": "The Sleeper league ID",
					"required":    true,
				},
				"roster_id": map[string]interface{}{
					"type":        "integer",
					"description": "The roster ID of the team to retrieve",
					"required":    true,
				},
			},
		},
	}
}

// GetAllRostersTool returns the get_all_rosters MCP tool definition
func (h *RosterHandler) GetAllRostersTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_all_rosters",
		Description: "Get all team rosters in a league for comparison and analysis",
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

// AnalyzeRosterStrengthTool returns the analyze_roster_strength MCP tool definition
func (h *RosterHandler) AnalyzeRosterStrengthTool() mcp.Tool {
	return mcp.Tool{
		Name:        "analyze_roster_strength",
		Description: "Analyze team roster strength with positional breakdown, starter quality, and injury assessment",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"league_id": map[string]interface{}{
					"type":        "string",
					"description": "The Sleeper league ID",
					"required":    true,
				},
				"roster_id": map[string]interface{}{
					"type":        "integer",
					"description": "The roster ID of the team to analyze",
					"required":    true,
				},
			},
		},
	}
}

// CompareRostersTool returns the compare_rosters MCP tool definition
func (h *RosterHandler) CompareRostersTool() mcp.Tool {
	return mcp.Tool{
		Name:        "compare_rosters",
		Description: "Compare two team rosters side-by-side with positional breakdown and strength analysis",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"league_id": map[string]interface{}{
					"type":        "string",
					"description": "The Sleeper league ID",
					"required":    true,
				},
				"roster_id1": map[string]interface{}{
					"type":        "integer",
					"description": "The first roster ID to compare",
					"required":    true,
				},
				"roster_id2": map[string]interface{}{
					"type":        "integer",
					"description": "The second roster ID to compare",
					"required":    true,
				},
			},
		},
	}
}

// HandleGetRoster handles the get_roster tool execution
func (h *RosterHandler) HandleGetRoster(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// Parse and validate arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: "Error: league_id is required and must be a string",
				},
			},
		}, nil
	}

	rosterIDFloat, ok := args["roster_id"].(float64)
	if !ok {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: "Error: roster_id is required and must be an integer",
				},
			},
		}, nil
	}
	rosterID := int(rosterIDFloat)

	h.logger.WithFields(logrus.Fields{
		"league_id": leagueID,
		"roster_id": rosterID,
	}).Info("Getting roster information")

	// Get league rosters
	rosters, err := h.client.GetLeagueRosters(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league rosters")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting rosters: %v", err),
				},
			},
		}, nil
	}

	// Find the specific roster
	var targetRoster *sleeper.Roster
	for i := range rosters {
		if rosters[i].RosterID == rosterID {
			targetRoster = &rosters[i]
			break
		}
	}

	if targetRoster == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Roster with ID %d not found in league %s", rosterID, leagueID),
				},
			},
		}, nil
	}

	// Get league users to get owner information
	users, err := h.client.GetLeagueUsers(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league users")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting league users: %v", err),
				},
			},
		}, nil
	}

	// Find owner information
	var ownerName string
	for _, user := range users {
		if user.UserID == targetRoster.OwnerID {
			if user.DisplayName != "" {
				ownerName = user.DisplayName
			} else {
				ownerName = user.Username
			}
			break
		}
	}

	// Get all players for player name lookup
	allPlayers, err := h.client.GetAllPlayers()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get all players")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting player data: %v", err),
				},
			},
		}, nil
	}

	// Build roster information with player details
	response := fmt.Sprintf("## Roster Information for Team %d\n\n", rosterID)
	if ownerName != "" {
		response += fmt.Sprintf("**Owner:** %s\n", ownerName)
	}
	response += fmt.Sprintf("**League ID:** %s\n", leagueID)
	response += fmt.Sprintf("**Record:** %d-%d-%d\n", 
		targetRoster.Settings.Wins, 
		targetRoster.Settings.Losses, 
		targetRoster.Settings.Ties)
	response += fmt.Sprintf("**Points For:** %.1f\n", targetRoster.Settings.FPTS)
	response += fmt.Sprintf("**Points Against:** %.1f\n\n", targetRoster.Settings.FPTSAgainst)

	// Starting lineup
	if len(targetRoster.Starters) > 0 {
		response += "### Starting Lineup\n"
		for i, playerID := range targetRoster.Starters {
			if playerID == "" || playerID == "0" {
				response += fmt.Sprintf("%d. Empty slot\n", i+1)
				continue
			}
			
			if player, exists := allPlayers[playerID]; exists {
				response += fmt.Sprintf("%d. %s (%s - %s)\n", i+1, player.FullName, player.Position, player.Team)
			} else {
				response += fmt.Sprintf("%d. Player ID: %s\n", i+1, playerID)
			}
		}
		response += "\n"
	}

	// Bench players
	benchPlayers := []string{}
	for _, playerID := range targetRoster.Players {
		if playerID == "" || playerID == "0" {
			continue
		}
		
		// Check if player is in starting lineup
		isStarter := false
		for _, starterID := range targetRoster.Starters {
			if playerID == starterID {
				isStarter = true
				break
			}
		}
		
		if !isStarter {
			benchPlayers = append(benchPlayers, playerID)
		}
	}

	if len(benchPlayers) > 0 {
		response += "### Bench\n"
		for i, playerID := range benchPlayers {
			if player, exists := allPlayers[playerID]; exists {
				response += fmt.Sprintf("%d. %s (%s - %s)\n", i+1, player.FullName, player.Position, player.Team)
			} else {
				response += fmt.Sprintf("%d. Player ID: %s\n", i+1, playerID)
			}
		}
	}

	return &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{
				Type: "text",
				Text: response,
			},
		},
	}, nil
}

// HandleGetAllRosters handles the get_all_rosters tool execution
func (h *RosterHandler) HandleGetAllRosters(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// Parse and validate arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: "Error: league_id is required and must be a string",
				},
			},
		}, nil
	}

	h.logger.WithFields(logrus.Fields{
		"league_id": leagueID,
	}).Info("Getting all roster information")

	// Get league rosters
	rosters, err := h.client.GetLeagueRosters(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league rosters")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting rosters: %v", err),
				},
			},
		}, nil
	}

	// Get league users to get owner information
	users, err := h.client.GetLeagueUsers(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league users")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting league users: %v", err),
				},
			},
		}, nil
	}

	// Create owner lookup map
	ownerMap := make(map[string]string)
	for _, user := range users {
		name := user.DisplayName
		if name == "" {
			name = user.Username
		}
		ownerMap[user.UserID] = name
	}

	// Build response with roster summaries
	response := fmt.Sprintf("## All Rosters in League %s\n\n", leagueID)
	
	for _, roster := range rosters {
		ownerName := ownerMap[roster.OwnerID]
		if ownerName == "" {
			ownerName = "Unknown Owner"
		}
		
		response += fmt.Sprintf("### Team %d - %s\n", roster.RosterID, ownerName)
		response += fmt.Sprintf("**Record:** %d-%d-%d\n", 
			roster.Settings.Wins, 
			roster.Settings.Losses, 
			roster.Settings.Ties)
		response += fmt.Sprintf("**Points For:** %.1f\n", roster.Settings.FPTS)
		response += fmt.Sprintf("**Points Against:** %.1f\n", roster.Settings.FPTSAgainst)
		response += fmt.Sprintf("**Total Players:** %d\n", len(roster.Players))
		response += fmt.Sprintf("**Starters:** %d\n\n", len(roster.Starters))
	}

	return &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{
				Type: "text",
				Text: response,
			},
		},
	}, nil
}

// HandleAnalyzeRosterStrength handles the analyze_roster_strength tool execution
func (h *RosterHandler) HandleAnalyzeRosterStrength(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// Parse and validate arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: "Error: league_id is required and must be a string",
				},
			},
		}, nil
	}

	rosterIDFloat, ok := args["roster_id"].(float64)
	if !ok {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: "Error: roster_id is required and must be an integer",
				},
			},
		}, nil
	}
	rosterID := int(rosterIDFloat)

	h.logger.WithFields(logrus.Fields{
		"league_id": leagueID,
		"roster_id": rosterID,
	}).Info("Analyzing roster strength")

	// Get roster data using the existing get_roster logic
	rosters, err := h.client.GetLeagueRosters(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league rosters")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting rosters: %v", err),
				},
			},
		}, nil
	}

	// Find the specific roster
	var targetRoster *sleeper.Roster
	for i := range rosters {
		if rosters[i].RosterID == rosterID {
			targetRoster = &rosters[i]
			break
		}
	}

	if targetRoster == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Roster with ID %d not found in league %s", rosterID, leagueID),
				},
			},
		}, nil
	}

	// Get owner information
	users, err := h.client.GetLeagueUsers(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league users")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting league users: %v", err),
				},
			},
		}, nil
	}

	var ownerName string
	for _, user := range users {
		if user.UserID == targetRoster.OwnerID {
			if user.DisplayName != "" {
				ownerName = user.DisplayName
			} else {
				ownerName = user.Username
			}
			break
		}
	}

	// Get all players for detailed analysis
	allPlayers, err := h.client.GetAllPlayers()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get all players")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting player data: %v", err),
				},
			},
		}, nil
	}

	// Analyze roster strength
	analysis := analyzeRosterStrength(targetRoster, allPlayers)
	
	// Build detailed analysis response
	response := fmt.Sprintf("## Roster Strength Analysis for Team %d\n\n", rosterID)
	if ownerName != "" {
		response += fmt.Sprintf("**Owner:** %s\n", ownerName)
	}
	response += fmt.Sprintf("**League ID:** %s\n", leagueID)
	response += fmt.Sprintf("**Record:** %d-%d-%d\n", 
		targetRoster.Settings.Wins, 
		targetRoster.Settings.Losses, 
		targetRoster.Settings.Ties)
	response += fmt.Sprintf("**Points For:** %.1f\n", targetRoster.Settings.FPTS)
	response += fmt.Sprintf("**Points Against:** %.1f\n\n", targetRoster.Settings.FPTSAgainst)

	// Overall strength rating
	response += fmt.Sprintf("### Overall Strength: %s\n\n", analysis.OverallRating)

	// Position breakdown
	response += "### Positional Analysis\n\n"
	for _, posAnalysis := range analysis.PositionalBreakdown {
		response += fmt.Sprintf("**%s**\n", posAnalysis.Position)
		response += fmt.Sprintf("- Starters: %d\n", posAnalysis.StarterCount)
		response += fmt.Sprintf("- Bench: %d\n", posAnalysis.BenchCount)
		response += fmt.Sprintf("- Strength: %s\n", posAnalysis.Strength)
		if posAnalysis.InjuryConcerns > 0 {
			response += fmt.Sprintf("- Injury Concerns: %d players\n", posAnalysis.InjuryConcerns)
		}
		response += "\n"
	}

	// Key insights
	if len(analysis.Insights) > 0 {
		response += "### Key Insights\n\n"
		for _, insight := range analysis.Insights {
			response += fmt.Sprintf("- %s\n", insight)
		}
	}

	return &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{
				Type: "text",
				Text: response,
			},
		},
	}, nil
}

// HandleCompareRosters handles the compare_rosters tool execution
func (h *RosterHandler) HandleCompareRosters(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// Parse and validate arguments
	leagueID, ok := args["league_id"].(string)
	if !ok || leagueID == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: "Error: league_id is required and must be a string",
				},
			},
		}, nil
	}

	rosterID1Float, ok := args["roster_id1"].(float64)
	if !ok {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: "Error: roster_id1 is required and must be an integer",
				},
			},
		}, nil
	}
	rosterID1 := int(rosterID1Float)

	rosterID2Float, ok := args["roster_id2"].(float64)
	if !ok {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: "Error: roster_id2 is required and must be an integer",
				},
			},
		}, nil
	}
	rosterID2 := int(rosterID2Float)

	h.logger.WithFields(logrus.Fields{
		"league_id":  leagueID,
		"roster_id1": rosterID1,
		"roster_id2": rosterID2,
	}).Info("Comparing rosters")

	// Get roster data
	rosters, err := h.client.GetLeagueRosters(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league rosters")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting rosters: %v", err),
				},
			},
		}, nil
	}

	// Find both rosters
	var roster1, roster2 *sleeper.Roster
	for i := range rosters {
		if rosters[i].RosterID == rosterID1 {
			roster1 = &rosters[i]
		}
		if rosters[i].RosterID == rosterID2 {
			roster2 = &rosters[i]
		}
	}

	if roster1 == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Roster with ID %d not found in league %s", rosterID1, leagueID),
				},
			},
		}, nil
	}

	if roster2 == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Roster with ID %d not found in league %s", rosterID2, leagueID),
				},
			},
		}, nil
	}

	// Get owner information
	users, err := h.client.GetLeagueUsers(leagueID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get league users")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting league users: %v", err),
				},
			},
		}, nil
	}

	// Create owner lookup
	ownerMap := make(map[string]string)
	for _, user := range users {
		name := user.DisplayName
		if name == "" {
			name = user.Username
		}
		ownerMap[user.UserID] = name
	}

	// Get all players for detailed comparison
	allPlayers, err := h.client.GetAllPlayers()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get all players")
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error getting player data: %v", err),
				},
			},
		}, nil
	}

	// Build side-by-side comparison
	response := fmt.Sprintf("## Roster Comparison\n\n")
	response += fmt.Sprintf("| Metric | Team %d (%s) | Team %d (%s) |\n", 
		rosterID1, ownerMap[roster1.OwnerID], 
		rosterID2, ownerMap[roster2.OwnerID])
	response += "|--------|-------------|-------------|\n"
	response += fmt.Sprintf("| Record | %d-%d-%d | %d-%d-%d |\n",
		roster1.Settings.Wins, roster1.Settings.Losses, roster1.Settings.Ties,
		roster2.Settings.Wins, roster2.Settings.Losses, roster2.Settings.Ties)
	response += fmt.Sprintf("| Points For | %.1f | %.1f |\n",
		roster1.Settings.FPTS, roster2.Settings.FPTS)
	response += fmt.Sprintf("| Points Against | %.1f | %.1f |\n",
		roster1.Settings.FPTSAgainst, roster2.Settings.FPTSAgainst)
	response += fmt.Sprintf("| Total Players | %d | %d |\n",
		len(roster1.Players), len(roster2.Players))
	response += fmt.Sprintf("| Starters | %d | %d |\n\n",
		len(roster1.Starters), len(roster2.Starters))

	// Positional breakdown comparison
	comparison := compareRostersByPosition(roster1, roster2, allPlayers)
	
	response += "### Positional Comparison\n\n"
	for _, posComp := range comparison.PositionalComparisons {
		response += fmt.Sprintf("**%s**\n", posComp.Position)
		response += fmt.Sprintf("- Team %d: %d starters, %d bench\n", 
			rosterID1, posComp.Team1Starters, posComp.Team1Bench)
		response += fmt.Sprintf("- Team %d: %d starters, %d bench\n", 
			rosterID2, posComp.Team2Starters, posComp.Team2Bench)
		response += fmt.Sprintf("- Advantage: %s\n\n", posComp.Advantage)
	}

	// Overall assessment
	response += fmt.Sprintf("### Overall Assessment\n\n")
	response += fmt.Sprintf("**Stronger Team:** %s\n\n", comparison.OverallAdvantage)
	
	if len(comparison.Summary) > 0 {
		response += "**Key Differences:**\n"
		for _, point := range comparison.Summary {
			response += fmt.Sprintf("- %s\n", point)
		}
	}

	return &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{
			&mcp.TextContent{
				Type: "text",
				Text: response,
			},
		},
	}, nil
}

// RosterAnalysis represents the result of analyzing a roster's strength
type RosterAnalysis struct {
	OverallRating        string                 `json:"overall_rating"`
	PositionalBreakdown  []PositionalAnalysis   `json:"positional_breakdown"`
	Insights             []string               `json:"insights"`
}

// PositionalAnalysis represents analysis for a specific position group
type PositionalAnalysis struct {
	Position        string `json:"position"`
	StarterCount    int    `json:"starter_count"`
	BenchCount      int    `json:"bench_count"`
	Strength        string `json:"strength"`
	InjuryConcerns  int    `json:"injury_concerns"`
}

// RosterComparison represents the result of comparing two rosters
type RosterComparison struct {
	PositionalComparisons []PositionalComparison `json:"positional_comparisons"`
	OverallAdvantage      string                 `json:"overall_advantage"`
	Summary               []string               `json:"summary"`
}

// PositionalComparison represents comparison for a specific position
type PositionalComparison struct {
	Position      string `json:"position"`
	Team1Starters int    `json:"team1_starters"`
	Team1Bench    int    `json:"team1_bench"`
	Team2Starters int    `json:"team2_starters"`
	Team2Bench    int    `json:"team2_bench"`
	Advantage     string `json:"advantage"`
}
