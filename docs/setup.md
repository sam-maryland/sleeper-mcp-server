# Setup Instructions

## Prerequisites

- Go 1.21 or later
- Claude Desktop app (for local usage)

## Installation

### Option 1: Build from Source

1. Clone the repository:
```bash
git clone https://github.com/sam-maryland/sleeper-mcp-server.git
cd sleeper-mcp-server
```

2. Install dependencies:
```bash
make install
```

3. Build the server:
```bash
make build
```

### Option 2: Download Pre-built Binary

Download the appropriate binary for your platform from the [releases page](https://github.com/sam-maryland/sleeper-mcp-server/releases).

## Configuration

### Claude Desktop Setup

#### Method 1: Using Go Run (Development)

Add this to your Claude Desktop configuration file:

**macOS/Linux:** `~/.config/claude/claude_desktop_config.json`  
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "sleeper": {
      "command": "go",
      "args": ["run", "cmd/server/main.go"],
      "cwd": "/absolute/path/to/sleeper-mcp-server"
    }
  }
}
```

#### Method 2: Using Pre-built Binary (Recommended)

```json
{
  "mcpServers": {
    "sleeper": {
      "command": "/absolute/path/to/sleeper-mcp-server/sleeper-mcp-server"
    }
  }
}
```

### Other MCP Clients

The server uses STDIO transport and should work with any MCP-compatible client. Start the server and communicate via stdin/stdout using the MCP protocol.

## Usage

### Available Tools

- `get_league_info` - Get comprehensive league information including settings, scoring rules, and metadata
- `get_league_standings` - Get current standings with flexible tiebreaker support (accepts custom instructions)
- `get_league_users` - Get all league members and their team information
- `get_matchups` - Get weekly matchup data and scores

### Example Usage

Once configured, you can ask Claude:

- "Get information about my Sleeper league with ID YOUR_LEAGUE_ID"
- "Show me the current standings for league YOUR_LEAGUE_ID"
- "Get the standings for league YOUR_LEAGUE_ID using custom tiebreakers: when teams have the same wins, use head-to-head record first, then points for"
- "Who are all the users in league YOUR_LEAGUE_ID?"
- "Show me the matchups for week 1 in league YOUR_LEAGUE_ID"

### Testing Your Setup

After configuration, test the setup:

1. **Restart Claude Desktop** to load the new configuration
2. **Start a new conversation** and try a simple command:
   > "Get information about Sleeper league YOUR_LEAGUE_ID"
3. **Verify the server is working** - you should see league details returned

If you encounter issues, check the Claude Desktop logs or run the server manually to see error output.

## Development

### Running in Development Mode

```bash
make dev
```

### Building for Production

```bash
make build
```

### Running Tests

```bash
make test
```

### Cross-platform Builds

```bash
make build-release
```

This creates binaries for:
- Linux (AMD64)
- macOS (AMD64 and ARM64)
- Windows (AMD64)

## Troubleshooting

### Logging

The server outputs structured JSON logs. Set log level with environment variable:

```bash
LOG_LEVEL=debug ./sleeper-mcp-server
```

Available levels: debug, info, warn, error

## API Rate Limits

The Sleeper API has a rate limit of 1000 requests per minute. The server automatically handles rate limiting and provides appropriate error messages if limits are exceeded.