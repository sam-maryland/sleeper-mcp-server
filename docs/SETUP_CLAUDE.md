# Setting Up Claude Desktop with Sleeper MCP Server

## Step 1: Update Claude Desktop Configuration

1. **Find your Claude Desktop config file:**
   - **macOS:** `~/.config/claude/claude_desktop_config.json`
   - **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

2. **Edit the file** and add the Sleeper MCP server configuration:

```json
{
  "mcpServers": {
    "sleeper": {
      "command": "go",
      "args": ["run", "cmd/server/main.go"],
      "cwd": "/Users/sammaryland/Projects/sleeper-mcp-server"
    }
  }
}
```

**Important:** If you already have other MCP servers configured, add the `"sleeper"` entry inside the existing `"mcpServers"` object.

## Step 2: Verify Configuration

1. **Test the MCP server works locally:**
```bash
cd /Users/sammaryland/Projects/sleeper-mcp-server
go run cmd/server/main.go
```
This should start the server and wait for input.

2. **Test with a simple request** (in another terminal):
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | go run cmd/server/main.go
```

## Step 3: Restart Claude Desktop

1. **Completely quit Claude Desktop** (Cmd+Q on macOS)
2. **Restart Claude Desktop**
3. **Wait a moment** for it to load the new configuration

## Step 4: Test the Connection

1. **Start a new conversation** in Claude Desktop
2. **Ask a test question:**
   - "List the available tools for Sleeper fantasy football"
   - "Show me the league information for league 1046562655881940992"

## Troubleshooting

### If Claude doesn't recognize the MCP server:

1. **Check the config file location** - make sure you're editing the right file
2. **Verify JSON syntax** - use a JSON validator to check for syntax errors
3. **Check the working directory path** - make sure `/Users/sammaryland/Projects/sleeper-mcp-server` exists
4. **Restart Claude Desktop** completely (not just refresh)

### If you get permission errors:

1. **Make sure Go is installed** and accessible from the command line
2. **Test the server manually** using the commands in Step 2

### If tools don't work:

1. **Check if the server is running** - you should see log output when tools are called
2. **Verify your league is configured** in `configs/league_settings.json`

## Success Indicators

When working correctly, you should be able to ask Claude:

- **"Get the standings for my Sleeper league 1046562655881940992"**
- **"Show me all the users in my league"**
- **"What were the matchups for week 1?"**

And Claude will automatically apply your custom head-to-head tiebreaker rules!