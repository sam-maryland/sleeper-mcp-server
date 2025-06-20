# Sleeper MCP Server

A Model Context Protocol server that connects AI agents to Sleeper Fantasy Football, enabling intelligent league analysis, roster evaluation, and strategic insights.

## What This Does

This MCP server allows you to ask AI agents natural language questions about your Sleeper fantasy football leagues:

- **"Show me the standings for my league"** - Get current standings with custom tiebreaker support
- **"Who are the users in my league?"** - View all league members
- **"What were the matchups for week 5?"** - See weekly scoring and results
- **"Calculate standings using head-to-head tiebreakers"** - Apply custom standings rules

## Quick Start

1. **Install and configure** - See [setup instructions](docs/setup.md)
2. **Configure your league settings** - Edit `configs/league_settings.json` with your league's custom rules
3. **Start asking questions** - Use natural language to analyze your league

## Documentation

- [Setup Instructions](docs/setup.md) - **Start here**
- [Project Vision](docs/01-project-vision.md)
- [API Reference](docs/02-sleeper-api-reference.md) 
- [Tool Specifications](docs/04-tool-specifications.md)
- [Custom Standings Example](docs/05-example-custom-standings.md)
- [Development Guide](docs/DEVELOPMENT.md) - For contributors

## Key Features

- **Flexible Standings**: Supports custom tiebreaker rules not available in Sleeper
- **Natural Language**: Ask questions conversationally - no need to learn APIs
- **Agent Agnostic**: Works with Claude, ChatGPT, or any MCP-compatible AI agent
- **Head-to-Head Calculation**: Automatically calculates complex tiebreakers from matchup data
- **Multiple Leagues**: Handle different leagues with different rule sets

## Setting Up Your League

### 1. Configure Your League Settings

Edit `configs/league_settings.json` to add your league's information. Copy the template and replace with your details:

```json
{
  "leagues": {
    "1234567890123456789": {
      "name": "My Fantasy League",
      "description": "League with custom head-to-head tiebreakers",
      "custom_standings": {
        "enabled": true,
        "instructions": "When teams have the same wins, use head-to-head record first, then points for, then points against",
        "tiebreak_order": ["wins", "head_to_head", "points_for", "points_against"],
        "notes": "This league uses custom tiebreakers not supported by Sleeper natively"
      }
    }
  }
}
```

### 2. Example Conversations

Once configured, you can ask your AI agent:

- **"What are the current standings in my league?"**
- **"Show me the standings with our custom head-to-head tiebreakers"**
- **"Who beat who this week in my league?"**
- **"Compare the playoff seeding using our league rules"**

## For Developers

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for:
- Development workflow and testing
- Contributing guidelines  
- Adding new features
- Building and deployment

## How It Works

1. **You ask your AI agent** about your league in natural language
2. **The agent uses this MCP server** to fetch data from Sleeper's API
3. **Custom calculations are applied** based on your `configs/league_settings.json`
4. **You get intelligent analysis** tailored to your league's rules

The server automatically applies your league's custom settings, so you don't need to explain the rules every time.

## License
MIT License - see LICENSE file