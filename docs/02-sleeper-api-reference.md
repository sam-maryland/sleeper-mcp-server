# Sleeper API Reference & Integration Guide

## API Overview

The Sleeper API is a read-only HTTP API that provides access to fantasy football league data without requiring authentication. All endpoints return JSON data and should be called with respect to rate limits.

**Base URL**: `https://api.sleeper.app/v1`
**Rate Limit**: Stay under 1000 API calls per minute
**Authentication**: None required
**Data Format**: JSON

## Core API Endpoints

### User Endpoints

#### Get User by Username/ID
```
GET /user/{username_or_user_id}
```
Returns user profile information including display name, avatar, and user ID.

#### Get User's Leagues
```
GET /user/{user_id}/leagues/{sport}/{season}
```
Returns all leagues a user participates in for a given sport and season.
- `sport`: Currently only "nfl" supported
- `season`: Year (e.g., "2024")

### League Endpoints

#### Get League Information
```
GET /league/{league_id}
```
Returns comprehensive league settings, scoring rules, roster requirements, and metadata.

Key fields:
- `settings`: Scoring rules, roster positions, playoff settings
- `scoring_settings`: Point values for different statistical categories
- `roster_positions`: Required roster positions (QB, RB, WR, etc.)
- `status`: League status (pre_draft, drafting, in_season, complete)

#### Get League Users
```
GET /league/{league_id}/users
```
Returns all users in the league with their display names, avatars, and team metadata.

#### Get League Rosters
```
GET /league/{league_id}/rosters
```
Returns all team rosters with current players, settings, and owner information.

Key fields:
- `players`: Array of player IDs on the roster
- `starters`: Array of player IDs in starting lineup
- `roster_id`: Unique identifier for the team
- `owner_id`: User ID of the team owner

#### Get Matchups
```
GET /league/{league_id}/matchups/{week}
```
Returns matchup data for a specific week including scores and roster information.
- `week`: Week number (1-18 for regular season)

#### Get Transactions
```
GET /league/{league_id}/transactions/{week}
```
Returns all transactions (trades, adds, drops) for a specific week.

#### Get Traded Picks
```
GET /league/{league_id}/traded_picks
```
Returns all draft pick trades in the league.

### Player Endpoints

#### Get All Players
```
GET /players/{sport}
```
Returns comprehensive player database. **Use sparingly** - intended for daily updates only.

Key fields:
- `player_id`: Unique identifier
- `full_name`: Player's full name
- `position`: Player position (QB, RB, WR, TE, etc.)
- `team`: NFL team abbreviation
- `status`: Active, Inactive, etc.
- `injury_status`: Current injury status

#### Get Trending Players
```
GET /players/{sport}/trending/{type}
```
Returns trending players for adds/drops.
- `type`: "add" or "drop"
- Query params: `lookback_hours`, `limit`

### Draft Endpoints

#### Get League Drafts
```
GET /league/{league_id}/drafts
```
Returns all drafts associated with the league.

#### Get Draft Details
```
GET /draft/{draft_id}
```
Returns detailed draft information including settings and metadata.

#### Get Draft Picks
```
GET /draft/{draft_id}/picks
```
Returns all picks made in the draft with player and roster information.

## Data Models

### League Object Structure
```json
{
  "league_id": "string",
  "name": "string",
  "status": "pre_draft|drafting|in_season|complete",
  "sport": "nfl",
  "season": "2024",
  "settings": {
    "playoff_teams": 6,
    "playoff_weeks_per_matchup": 1,
    "playoff_round_type": 0
  },
  "scoring_settings": {
    "pass_yd": 0.04,
    "pass_td": 4,
    "rush_yd": 0.1,
    "rec": 1
  },
  "roster_positions": ["QB", "RB", "RB", "WR", "WR", "TE", "FLEX", "K", "DEF", "BN", "BN"]
}
```

### Roster Object Structure
```json
{
  "roster_id": 1,
  "owner_id": "user_id",
  "players": ["player_id1", "player_id2"],
  "starters": ["player_id1", "player_id2"],
  "settings": {
    "wins": 5,
    "losses": 3,
    "ties": 0,
    "fpts": 1234.56
  }
}
```

### Player Object Structure
```json
{
  "player_id": "3086",
  "full_name": "Tom Brady",
  "position": "QB",
  "team": "TB",
  "status": "Active",
  "injury_status": null,
  "fantasy_positions": ["QB"],
  "age": 45,
  "height": "6'4\"",
  "weight": "220"
}
```

## Integration Best Practices

### Rate Limiting
- Implement client-side rate limiting to stay under 1000 calls/minute
- Cache frequently accessed data (players, league settings)
- Batch requests when possible
- Use exponential backoff for retries

### Error Handling
- Handle HTTP error codes gracefully (404, 429, 500)
- Validate league IDs and user IDs before making requests
- Provide meaningful error messages to users
- Log API errors for debugging

### Data Caching Strategy
- **Players Data**: Cache for 24 hours (updates daily)
- **League Settings**: Cache for duration of session
- **Live Scores**: Cache for 5-10 minutes during games
- **Historical Data**: Cache indefinitely (doesn't change)

### Common Use Cases

#### Getting Started Workflow
1. Get user by username → Extract user_id
2. Get user's leagues for current season → Extract league_id
3. Get league information → Understand scoring and settings
4. Get rosters → Analyze team compositions

#### Weekly Analysis Workflow
1. Get current week matchups → See who's playing whom
2. Get trending players → Identify waiver wire targets
3. Get recent transactions → See league activity
4. Analyze roster needs → Provide recommendations

#### Trade Analysis Workflow
1. Get both team rosters → Compare current players
2. Get league scoring settings → Calculate values
3. Get recent transactions → Understand market values
4. Evaluate trade fairness → Provide analysis

## MCP Tool Design Patterns

### Tool Naming Convention
Use clear, action-oriented names:
- `get_league_standings`
- `analyze_roster_strength`
- `find_trade_targets`
- `get_trending_players`

### Parameter Design
Make parameters intuitive and well-documented:
```go
type GetLeagueStandingsArgs struct {
    LeagueID string `json:"league_id" jsonschema:"required,description=The Sleeper league ID"`
}
```

### Response Format
Provide structured, AI-friendly responses:
```json
{
  "success": true,
  "data": { /* relevant data */ },
  "summary": "Human-readable summary for AI context",
  "metadata": {
    "updated_at": "2024-01-15T10:30:00Z",
    "source": "sleeper_api"
  }
}
```

This reference provides everything needed to build comprehensive MCP tools that leverage the full power of the Sleeper API while maintaining performance and reliability.