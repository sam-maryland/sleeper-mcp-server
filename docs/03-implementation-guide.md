# MCP Server Implementation Guide - Technical Specifications

## Project Structure

```
sleeper-mcp-server/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── mcp/
│   │   ├── server.go           # MCP server setup
│   │   └── tools.go            # Tool registration
│   ├── sleeper/
│   │   ├── client.go           # Sleeper API client
│   │   ├── models.go           # Data structures
│   │   └── cache.go            # Caching layer
│   ├── handlers/
│   │   ├── league.go           # League-related tools
│   │   ├── roster.go           # Roster analysis tools
│   │   ├── player.go           # Player data tools
│   │   └── analysis.go         # Advanced analysis tools
│   └── utils/
│       ├── cache.go            # Utility functions
│       └── validation.go       # Input validation
├── configs/
│   └── claude_desktop_config.json  # Example config
├── docs/
│   ├── setup.md                # Setup instructions
│   └── tools.md                # Tool documentation
├── go.mod
├── go.sum
├── README.md
└── Makefile
```

## Technical Stack

### Dependencies
```go
module github.com/sam-maryland/sleeper-mcp-server

go 1.21

require (
    github.com/mark3labs/mcp-go v0.1.0
    github.com/patrickmn/go-cache v2.1.0+incompatible
    github.com/sirupsen/logrus v1.9.3
)
```

### Core Technologies
- **MCP SDK**: mark3labs/mcp-go for MCP protocol implementation
- **HTTP Client**: Standard net/http with custom retry logic
- **Caching**: patrickmn/go-cache for in-memory caching
- **Logging**: sirupsen/logrus for structured logging
- **Transport**: STDIO for maximum client compatibility

## MCP Server Architecture

### Server Initialization
```go
func NewSleeperMCPServer() *server.MCPServer {
    s := server.NewMCPServer(
        "Sleeper Fantasy Football",
        "1.0.0",
        server.WithToolCapabilities(false),
    )
    
    // Register all tools
    registerLeagueTools(s)
    registerRosterTools(s)
    registerPlayerTools(s)
    registerAnalysisTools(s)
    
    return s
}
```

### Tool Categories

#### 1. League Information Tools
- `get_league_info` - Basic league settings and metadata
- `get_league_standings` - Current standings with win/loss records
- `get_league_users` - All league members and team names
- `get_matchups` - Weekly matchups and scores

#### 2. Roster Management Tools  
- `get_roster` - Specific team's current roster
- `get_all_rosters` - All rosters in the league
- `analyze_roster_strength` - Evaluate team composition
- `compare_rosters` - Head-to-head team comparison

#### 3. Player Data Tools
- `get_trending_players` - Hot waiver wire targets
- `get_player_stats` - Individual player information
- `search_players` - Find players by name/position

#### 4. Transaction Tools
- `get_transactions` - Recent trades, adds, drops
- `get_traded_picks` - Draft pick transactions
- `analyze_trade` - Evaluate trade fairness

#### 5. Analysis Tools
- `suggest_trade_targets` - AI-powered trade recommendations
- `predict_matchup` - Week outcome predictions
- `identify_sleepers` - Undervalued players
- `generate_league_report` - Comprehensive league summary

## Data Structures

### Core Models
```go
// League represents a Sleeper fantasy league
type League struct {
    LeagueID         string                 `json:"league_id"`
    Name             string                 `json:"name"`
    Status           string                 `json:"status"`
    Sport            string                 `json:"sport"`
    Season           string                 `json:"season"`
    Settings         LeagueSettings         `json:"settings"`
    ScoringSettings  map[string]float64     `json:"scoring_settings"`
    RosterPositions  []string               `json:"roster_positions"`
}

// Roster represents a team's roster
type Roster struct {
    RosterID    int      `json:"roster_id"`
    OwnerID     string   `json:"owner_id"`
    Players     []string `json:"players"`
    Starters    []string `json:"starters"`
    Settings    RosterSettings `json:"settings"`
}

// Player represents an NFL player
type Player struct {
    PlayerID          string   `json:"player_id"`
    FullName          string   `json:"full_name"`
    Position          string   `json:"position"`
    Team              string   `json:"team"`
    Status            string   `json:"status"`
    InjuryStatus      string   `json:"injury_status"`
    FantasyPositions  []string `json:"fantasy_positions"`
}
```

### Tool Parameter Structures
```go
// GetLeagueInfoArgs parameters for league info tool
type GetLeagueInfoArgs struct {
    LeagueID string `json:"league_id" jsonschema:"required,description=The Sleeper league ID"`
}

// AnalyzeRosterArgs parameters for roster analysis
type AnalyzeRosterArgs struct {
    LeagueID string `json:"league_id" jsonschema:"required,description=The Sleeper league ID"`
    RosterID int    `json:"roster_id" jsonschema:"required,description=The roster ID to analyze"`
}

// GetTrendingPlayersArgs parameters for trending players
type GetTrendingPlayersArgs struct {
    Type         string `json:"type" jsonschema:"required,enum=add|drop,description=Type of trending (add or drop)"`
    Hours        int    `json:"hours" jsonschema:"description=Lookback hours (default: 24)"`
    Limit        int    `json:"limit" jsonschema:"description=Number of players to return (default: 25)"`
}
```

## Sleeper API Client

### Client Interface
```go
type SleeperClient interface {
    // User methods
    GetUser(usernameOrID string) (*User, error)
    GetUserLeagues(userID, sport, season string) ([]League, error)
    
    // League methods
    GetLeague(leagueID string) (*League, error)
    GetLeagueUsers(leagueID string) ([]User, error)
    GetLeagueRosters(leagueID string) ([]Roster, error)
    GetMatchups(leagueID string, week int) ([]Matchup, error)
    GetTransactions(leagueID string, week int) ([]Transaction, error)
    
    // Player methods
    GetAllPlayers() (map[string]Player, error)
    GetTrendingPlayers(sport, trendType string, hours, limit int) ([]Player, error)
}
```

### HTTP Client Implementation
```go
type HTTPClient struct {
    baseURL    string
    httpClient *http.Client
    cache      *cache.Cache
    logger     *logrus.Logger
}

func NewHTTPClient() *HTTPClient {
    return &HTTPClient{
        baseURL: "https://api.sleeper.app/v1",
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
        cache:  cache.New(5*time.Minute, 10*time.Minute),
        logger: logrus.New(),
    }
}
```

## Caching Strategy

### Cache Configuration
```go
type CacheConfig struct {
    // Players data - updates daily
    PlayersDataTTL    time.Duration // 24 hours
    
    // League settings - rarely change
    LeagueSettingsTTL time.Duration // 1 hour
    
    // Live scores - frequent updates during games
    LiveScoresTTL     time.Duration // 5 minutes
    
    // Historical data - never changes
    HistoricalDataTTL time.Duration // 7 days
}
```

### Cache Keys
- `players:all` - All NFL players
- `league:{league_id}:info` - League information
- `league:{league_id}:rosters` - League rosters
- `league:{league_id}:matchups:{week}` - Weekly matchups
- `trending:players:{type}:{hours}` - Trending players

## Error Handling

### Error Types
```go
type SleeperError struct {
    Type       string `json:"type"`
    Message    string `json:"message"`
    StatusCode int    `json:"status_code,omitempty"`
    LeagueID   string `json:"league_id,omitempty"`
}

func (e *SleeperError) Error() string {
    return fmt.Sprintf("sleeper api error: %s", e.Message)
}
```

### Error Handling Patterns
- **404 Not Found**: Invalid league/user ID
- **429 Rate Limited**: Implement exponential backoff
- **500 Server Error**: Retry with backoff, fallback to cached data
- **Network Errors**: Retry up to 3 times with increasing delays

## Tool Response Format

### Standard Response Structure
```go
type ToolResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Summary   string      `json:"summary"`
    Error     string      `json:"error,omitempty"`
    Metadata  Metadata    `json:"metadata"`
}

type Metadata struct {
    Timestamp  time.Time `json:"timestamp"`
    Source     string    `json:"source"`
    CacheHit   bool      `json:"cache_hit"`
    APICallsUsed int     `json:"api_calls_used"`
}
```

## Testing Strategy

### Unit Tests
- Test each tool handler with mock API responses
- Validate parameter parsing and validation
- Test error handling scenarios
- Cache behavior verification

### Integration Tests
- Test against live Sleeper API (rate-limited)
- End-to-end MCP protocol testing
- Performance testing with realistic data loads

### Test Data
- Mock league with diverse team configurations
- Sample player database for testing
- Various transaction scenarios

### Test Validation Requirements

**CRITICAL**: After every body of changes, run the full test suite to validate:

```bash
make test
```

This command runs:
1. **Unit tests**: `go test -count=1 ./...`
2. **Integration tests**: `go test -count=1 -tags=integration ./...`

**Required for all changes**:
- New feature implementations
- Bug fixes
- Refactoring
- Documentation updates that affect code
- Configuration changes

**Test failure policy**: All tests must pass before committing changes or considering work complete.

## Performance Considerations

### Optimization Strategies
1. **Aggressive Caching**: Cache all API responses appropriately
2. **Batch Requests**: Combine related API calls when possible
3. **Lazy Loading**: Only fetch data when requested
4. **Rate Limiting**: Implement client-side rate limiting
5. **Connection Pooling**: Reuse HTTP connections

### Monitoring
- Log API response times
- Track cache hit rates
- Monitor rate limit usage
- Alert on error rates

## Configuration Management

### Environment Variables
```bash
SLEEPER_API_BASE_URL=https://api.sleeper.app/v1
LOG_LEVEL=info
CACHE_DEFAULT_TTL=5m
RATE_LIMIT_PER_MINUTE=900
```

### Configuration File
```json
{
  "server": {
    "name": "Sleeper Fantasy Football",
    "version": "1.0.0"
  },
  "sleeper_api": {
    "base_url": "https://api.sleeper.app/v1",
    "timeout": "10s",
    "rate_limit": 900
  },
  "cache": {
    "default_ttl": "5m",
    "cleanup_interval": "10m"
  }
}
```

This implementation guide provides a comprehensive foundation for building a robust, performant, and maintainable Sleeper MCP server that can handle real-world usage while providing valuable insights to fantasy football players.