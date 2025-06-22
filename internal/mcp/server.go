package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sam-maryland/sleeper-mcp-server/internal/handlers"
	"github.com/sam-maryland/sleeper-mcp-server/internal/sleeper"
	"github.com/sirupsen/logrus"
)

type SleeperMCPServer struct {
	server        *server.DefaultServer
	logger        *logrus.Logger
	sleeperClient sleeper.Client
	leagueHandler *handlers.LeagueHandler
	rosterHandler *handlers.RosterHandler
}

func NewSleeperMCPServer(logger *logrus.Logger) *server.DefaultServer {
	// Create Sleeper API client
	sleeperClient := sleeper.NewHTTPClient(logger)
	
	// Create handlers
	leagueHandler := handlers.NewLeagueHandler(sleeperClient, logger)
	rosterHandler := handlers.NewRosterHandler(sleeperClient, logger)
	
	// Create MCP server
	s := server.NewDefaultServer("Sleeper Fantasy Football", "1.0.0")
	
	if s == nil {
		logger.Error("Failed to create MCP server instance")
		return nil
	}

	logger.Info("MCP server instance created successfully")
	
	// Set up list tools handler
	s.HandleListTools(func(ctx context.Context, cursor *string) (*mcp.ListToolsResult, error) {
		tools := []mcp.Tool{
			leagueHandler.GetLeagueInfoTool(),
			leagueHandler.GetLeagueStandingsTool(),
			leagueHandler.GetLeagueUsersTool(),
			leagueHandler.GetMatchupsTool(),
			leagueHandler.DiscoverLeagueHistoryTool(),
			leagueHandler.GetLeagueHistoryTool(),
			rosterHandler.GetRosterTool(),
			rosterHandler.GetAllRostersTool(),
			rosterHandler.AnalyzeRosterStrengthTool(),
			rosterHandler.CompareRostersTool(),
		}
		
		logger.WithField("tools_count", len(tools)).Info("Listing available tools")
		
		return &mcp.ListToolsResult{
			Tools: tools,
		}, nil
	})
	
	// Set up call tool handler
	s.HandleCallTool(func(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		logger.WithFields(logrus.Fields{
			"tool": name,
			"args": arguments,
		}).Info("Tool called")
		
		// Route to specific tool handlers
		switch name {
		case "get_league_info":
			return leagueHandler.HandleGetLeagueInfo(ctx, arguments)
		case "get_league_standings":
			return leagueHandler.HandleGetLeagueStandings(ctx, arguments)
		case "get_league_users":
			return leagueHandler.HandleGetLeagueUsers(ctx, arguments)
		case "get_matchups":
			return leagueHandler.HandleGetMatchups(ctx, arguments)
		case "discover_league_history":
			return leagueHandler.HandleDiscoverLeagueHistory(ctx, arguments)
		case "get_league_history":
			return leagueHandler.HandleGetLeagueHistory(ctx, arguments)
		case "get_roster":
			return rosterHandler.HandleGetRoster(ctx, arguments)
		case "get_all_rosters":
			return rosterHandler.HandleGetAllRosters(ctx, arguments)
		case "analyze_roster_strength":
			return rosterHandler.HandleAnalyzeRosterStrength(ctx, arguments)
		case "compare_rosters":
			return rosterHandler.HandleCompareRosters(ctx, arguments)
		default:
			logger.WithField("tool", name).Warn("Unknown tool called")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Type: "text",
						Text: "Unknown tool: " + name,
					},
				},
				IsError: true,
			}, nil
		}
	})
	
	logger.Info("All tools registered successfully")
	return s
}