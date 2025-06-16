package main

import (
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sam-maryland/sleeper-mcp-server/internal/mcp"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	mcpServer := mcp.NewSleeperMCPServer(logger)
	if mcpServer == nil {
		logger.Fatal("Failed to create MCP server")
	}

	logger.Info("Starting Sleeper MCP Server...")
	
	if err := server.ServeStdio(mcpServer); err != nil {
		logger.WithError(err).Fatal("Server failed to start")
		os.Exit(1)
	}
}