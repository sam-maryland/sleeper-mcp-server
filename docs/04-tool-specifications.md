# Initial Tool Specifications - MVP Implementation

## Phase 1: Core Foundation Tools

Start with these essential tools that provide immediate value and demonstrate core functionality. These tools cover the most common fantasy football use cases.

## Tool Category 1: League Information

### Tool: `get_league_info`
**Purpose**: Get comprehensive league information and settings

**Parameters**:
```json
{
  "league_id": "string (required) - The Sleeper league ID"
}
```

**Response Data**:
- League name and basic info
- Scoring settings (PPR, standard, etc.)
- Roster requirements (positions, bench size)
- Playoff structure
- Current season status

**AI Context**: "Provides essential league context for all other analysis. Use this first to understand scoring rules and roster structure."

### Tool: `get_league_standings`
**Purpose**: Get current league standings with wins, losses, points scored, and playoff positioning. Supports custom tiebreaker rules.

**Parameters**:
```json
{
  "league_id": "string (required) - The Sleeper league ID",
  "tiebreak_order": "array (optional) - Custom tiebreaker order. Options: wins, points_for, points_against, head_to_head, division_record, custom",
  "custom_metrics": "object (optional) - Custom metrics for tiebreakers (e.g., all_play_record, bench_points, etc.)",
  "instructions": "string (optional) - Natural language instructions for standings calculation (e.g., 'When teams have the same wins, use head-to-head record first, then points for')",
  "mode": "string (optional) - 'regular_season' (default) or 'final' - determines which weeks to include in calculations"
}
```

**Response Data**:
- Team rankings by wins/losses with flexible tiebreaker support
- Total points scored and allowed by each team
- Playoff positioning and division information
- Team owner information
- Head-to-head records (when applicable)
- Tiebreaker notes showing which rules were applied
- Custom metrics (if provided)

**Tiebreaker Options**:
- `wins` - Win-loss record (standard)
- `points_for` - Total points scored
- `points_against` - Total points allowed (lower is better)
- `head_to_head` - Head-to-head record calculated from matchup data
- `division_record` - Division win-loss record (future enhancement)
- `custom` - User-defined custom metrics

**Natural Language Processing**: The tool can parse instructions like:
- "When teams have the same wins, use head-to-head record first, then points for"
- "Use head-to-head record as first tiebreaker, then total points scored"
- "Tiebreakers: H2H, points for, points against"

**Modes**:
- `regular_season` (default): Uses weeks 1-14 for calculations
- `final`: Uses all weeks 1-18, including playoff results

**League Integration**: Automatically detects and uses the league's `playoff_seed_type` setting from Sleeper if no custom tiebreaker order is provided.

**AI Context**: "Shows competitive landscape with flexible ranking rules that can be described in natural language. Perfect for leagues with complex custom tiebreaker systems like the example in 05-example-custom-standings.md. Can handle sophisticated instructions about how standings should be calculated, including head-to-head records and different modes for regular season vs final standings."

## Tool Category 2: Roster Management

### Tool: `get_roster`
**Purpose**: Get a specific team's current roster and lineup

**Parameters**:
```json
{
  "league_id": "string (required) - The Sleeper league ID",
  "roster_id": "integer (required) - The roster ID to retrieve"
}
```

**Response Data**:
- All players on the roster
- Current starting lineup
- Bench players
- Player positions and team affiliations
- Basic stats (if available)

**AI Context**: "Essential for analyzing individual teams. Use for start/sit advice, roster construction analysis, and trade evaluation."

### Tool: `get_all_rosters`
**Purpose**: Get all rosters in the league for comparative analysis

**Parameters**:
```json
{
  "league_id": "string (required) - The Sleeper league ID"
}
```

**Response Data**:
- All team rosters in the league
- Owner mappings
- Roster strengths by position

**AI Context**: "Enables league-wide analysis, trade target identification, and competitive balance assessment."

## Tool Category 3: Player Information

### Tool: `get_trending_players`
**Purpose**: Get currently trending players for waiver wire insights

**Parameters**:
```json
{
  "type": "string (required, enum: add|drop) - Type of trending to retrieve",
  "hours": "integer (optional, default: 24) - Lookback period in hours",
  "limit": "integer (optional, default: 25) - Number of players to return"
}
```

**Response Data**:
- List of trending players
- Player positions and teams
- Trending percentage/volume
- Basic player information

**AI Context**: "Critical for waiver wire strategy. Shows what players the fantasy community is targeting or dropping."

### Tool: `search_players`
**Purpose**: Find players by name, position, or team

**Parameters**:
```json
{
  "query": "string (required) - Search term (name, position, or team)",
  "limit": "integer (optional, default: 10) - Maximum results to return"
}
```

**Response Data**:
- Matching players with full details
- Position, team, status information
- Fantasy relevance indicators

**AI Context**: "Helps find specific players for analysis. Use when users mention player names or want position-specific research."

## Tool Category 4: Weekly Matchups

### Tool: `get_matchups`
**Purpose**: Get matchup information for a specific week

**Parameters**:
```json
{
  "league_id": "string (required) - The Sleeper league ID",
  "week": "integer (required) - Week number (1-18)"
}
```

**Response Data**:
- All matchups for the week
- Team scores (if week is complete)
- Roster information for each matchup
- Projected vs actual scoring

**AI Context**: "Essential for weekly analysis, playoff implications, and historical performance review."

## Tool Category 5: League Activity

### Tool: `get_recent_transactions`
**Purpose**: Get recent trades, adds, and drops in the league

**Parameters**:
```json
{
  "league_id": "string (required) - The Sleeper league ID",
  "week": "integer (optional) - Specific week to analyze, defaults to current",
  "limit": "integer (optional, default: 20) - Number of transactions to return"
}
```

**Response Data**:
- Recent trades with player details
- Waiver wire adds and drops
- Transaction timestamps
- Teams involved

**AI Context**: "Shows league activity and market dynamics. Useful for understanding player values and league engagement."

## Advanced Analysis Tools (Phase 1.5)

### Tool: `analyze_roster_strength`
**Purpose**: Provide intelligent analysis of a team's roster composition

**Parameters**:
```json
{
  "league_id": "string (required) - The Sleeper league ID",
  "roster_id": "integer (required) - The roster ID to analyze"
}
```

**Response Data**:
- Position-by-position strength analysis
- Depth assessment
- Potential weak spots
- Improvement recommendations

**AI Context**: "Provides detailed roster evaluation. Use for comprehensive team analysis and improvement suggestions."

### Tool: `suggest_trade_targets`
**Purpose**: Identify potential trade partners and scenarios

**Parameters**:
```json
{
  "league_id": "string (required) - The Sleeper league ID",
  "roster_id": "integer (required) - The roster seeking trades",
  "position_need": "string (optional) - Specific position to target (QB, RB, WR, TE)"
}
```

**Response Data**:
- Potential trade partners
- Players that make sense to target
- Teams with surplus at needed positions
- Fair trade value assessments

**AI Context**: "Advanced trade intelligence. Helps identify realistic trade opportunities based on roster construction and needs."

## Implementation Priority

### Week 1: Foundation
1. `get_league_info` - Essential for context
2. `get_roster` - Core functionality  
3. `get_all_rosters` - League overview

### Week 2: Player Data
4. `search_players` - Player lookup
5. `get_trending_players` - Waiver wire intel

### Week 3: Activity & Matchups
6. `get_matchups` - Weekly analysis
7. `get_recent_transactions` - League activity
8. `get_league_standings` - Competitive context

### Week 4: Intelligence
9. `analyze_roster_strength` - Smart analysis
10. `suggest_trade_targets` - Advanced recommendations

## Tool Response Standards

### Success Response Format
```json
{
  "success": true,
  "data": {
    // Tool-specific data structure
  },
  "summary": "Human-readable summary for AI context",
  "metadata": {
    "timestamp": "2024-01-15T10:30:00Z",
    "source": "sleeper_api",
    "cache_hit": false,
    "league_id": "123456789"
  }
}
```

### Error Response Format
```json
{
  "success": false,
  "error": "Clear error message for the user",
  "error_code": "LEAGUE_NOT_FOUND",
  "suggestions": [
    "Verify the league ID is correct",
    "Ensure the league is public or you have access"
  ],
  "metadata": {
    "timestamp": "2024-01-15T10: