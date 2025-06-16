package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sam-maryland/sleeper-mcp-server/internal/sleeper"
	"github.com/sirupsen/logrus"
)

// LeagueInfoArgs represents the parameters for the get_league_info tool
type LeagueInfoArgs struct {
	LeagueID string `json:"league_id"`
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