# Sleeper MCP Server Makefile

.PHONY: build clean test install dev

# Build the server binary
build:
	mkdir -p .bin
	go build -o .bin/sleeper-mcp-server ./cmd/server

# Clean build artifacts
clean:
	rm -rf .bin
	go clean

# Run tests (both unit and integration)
test:
	go test -count=1 ./...
	go test -count=1 -tags=integration ./...

# Install dependencies
install:
	go mod download
	go mod tidy

# Run in development mode
dev:
	go run ./cmd/server/main.go

# Build for release
build-release:
	mkdir -p .bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o .bin/sleeper-mcp-server-linux-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o .bin/sleeper-mcp-server-darwin-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o .bin/sleeper-mcp-server-darwin-arm64 ./cmd/server
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o .bin/sleeper-mcp-server-windows-amd64.exe ./cmd/server

# Test via MCP client (requires Claude Desktop configuration)
test-mcp:
	@echo "To test the MCP server:"
	@echo "1. Configure Claude Desktop with the server (see configs/claude_desktop_config.json)"
	@echo "2. Start a conversation and test tools naturally:"
	@echo "   - 'Get standings for league YOUR_LEAGUE_ID'"
	@echo "   - 'Show me the users in league YOUR_LEAGUE_ID'"
	@echo "   - 'Get matchups for week 1 in league YOUR_LEAGUE_ID'"
	@echo ""
	@echo "This tests both MCP protocol integration and tool functionality."

# Display help
help:
	@echo "Available targets:"
	@echo "  build          - Build the server binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests (REQUIRED after changes)"
	@echo "  install        - Install dependencies"
	@echo "  dev            - Run in development mode"
	@echo "  build-release  - Build for multiple platforms"
	@echo "  test-mcp       - Show MCP testing instructions"
	@echo "  help           - Show this help message"