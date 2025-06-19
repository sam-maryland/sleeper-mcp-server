# Sleeper MCP Server

A Model Context Protocol server for Sleeper Fantasy Football that enables AI agents to analyze leagues, manage rosters, and provide strategic insights.

## Quick Start
See [docs/setup.md](docs/setup.md) for detailed setup instructions.

## Documentation
- [Project Vision](docs/01-project-vision.md)
- [API Reference](docs/02-sleeper-api-reference.md) 
- [Implementation Guide](docs/03-implementation-guide.md)
- [Tool Specifications](docs/04-tool-specifications.md)
- [Custom Standings Example](docs/05-example-custom-standings.md)
- [Development Workflow](docs/DEVELOPMENT.md) - **Required reading for contributors**

## Features
- League analysis and standings with flexible tiebreaker support
- Roster management and evaluation
- Player trends and waiver wire insights
- Trade analysis and suggestions
- Natural language instruction parsing for custom standings

## Development

### Testing
**Critical**: Always run tests after making changes:

```bash
make test
```

This runs both unit and integration tests. All tests must pass before committing.

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for complete development workflow.

### Building
```bash
make build
```

### Running
```bash
make dev
```

### Testing via MCP Protocol
Test the server through Claude Desktop or other MCP clients:

```bash
# See MCP testing setup instructions
make test-mcp
```

1. Configure Claude Desktop with the server (see `configs/claude_desktop_config.json`)
2. Test with natural language commands like:
   - "Get standings for league YOUR_LEAGUE_ID"  
   - "Show users in league YOUR_LEAGUE_ID"

This tests both MCP integration and tool functionality as intended.

## License
MIT License - see LICENSE file