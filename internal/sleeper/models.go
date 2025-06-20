package sleeper

import "time"

// League represents a Sleeper fantasy league
type League struct {
	LeagueID        string                 `json:"league_id"`
	Name            string                 `json:"name"`
	Status          string                 `json:"status"`
	Sport           string                 `json:"sport"`
	Season          string                 `json:"season"`
	Settings        LeagueSettings         `json:"settings"`
	ScoringSettings map[string]float64     `json:"scoring_settings"`
	RosterPositions []string               `json:"roster_positions"`
	TotalRosters    int                    `json:"total_rosters"`
	DraftID         string                 `json:"draft_id"`
	Avatar          string                 `json:"avatar"`
}

// LeagueSettings contains league configuration
type LeagueSettings struct {
	PlayoffTeams         int `json:"playoff_teams"`
	PlayoffWeeksPerMatch int `json:"playoff_weeks_per_matchup"`
	PlayoffRoundType     int `json:"playoff_round_type"`
	PlayoffSeedType      int `json:"playoff_seed_type"`
	NumTeams             int `json:"num_teams"`
	LeagueAverageMatch   int `json:"league_average_match"`
	StartWeek            int `json:"start_week"`
	LastScoredLeg        int `json:"last_scored_leg"`
	Leg                  int `json:"leg"`
	MaxKeepers           int `json:"max_keepers"`
	DraftRounds          int `json:"draft_rounds"`
	TradeDeadline        int `json:"trade_deadline"`
	ReserveAllowCov      int `json:"reserve_allow_cov"`
	ReserveSlots         int `json:"reserve_slots"`
	PlayoffType          int `json:"playoff_type"`
	DailyWaivers         int `json:"daily_waivers"`
	WaiverDayOfWeek      int `json:"waiver_day_of_week"`
}

// User represents a Sleeper user
type User struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
}

// Roster represents a team's roster
type Roster struct {
	RosterID int      `json:"roster_id"`
	OwnerID  string   `json:"owner_id"`
	Players  []string `json:"players"`
	Starters []string `json:"starters"`
	Reserve  []string `json:"reserve"`
	Taxi     []string `json:"taxi"`
	Settings RosterSettings `json:"settings"`
}

// RosterSettings contains team performance data
type RosterSettings struct {
	Wins                    int     `json:"wins"`
	Losses                  int     `json:"losses"`
	Ties                    int     `json:"ties"`
	FPTS                    float64 `json:"fpts"`
	FPTSDecimal             float64 `json:"fpts_decimal"`
	FPTSAgainst             float64 `json:"fpts_against"`
	FPTSAgainstDecimal      float64 `json:"fpts_against_decimal"`
	TotalMoves              int     `json:"total_moves"`
	WaiverPosition          int     `json:"waiver_position"`
	WaiverBudgetUsed        int     `json:"waiver_budget_used"`
	Division                int     `json:"division,omitempty"`
	PlayoffSeed             int     `json:"playoff_seed,omitempty"`
}

// Player represents an NFL player
type Player struct {
	PlayerID         string   `json:"player_id"`
	FullName         string   `json:"full_name"`
	FirstName        string   `json:"first_name"`
	LastName         string   `json:"last_name"`
	Position         string   `json:"position"`
	Team             string   `json:"team"`
	Status           string   `json:"status"`
	InjuryStatus     string   `json:"injury_status"`
	FantasyPositions []string `json:"fantasy_positions"`
	Age              int      `json:"age"`
	Height           string   `json:"height"`
	Weight           string   `json:"weight"`
	YearsExp         int      `json:"years_exp"`
	College          string   `json:"college"`
}

// TrendingPlayer represents a trending player for adds/drops
type TrendingPlayer struct {
	PlayerID string `json:"player_id"`
	Count    int    `json:"count"`
}

// Matchup represents a weekly matchup
type Matchup struct {
	RosterID       int                    `json:"roster_id"`
	MatchupID      int                    `json:"matchup_id"`
	Points         float64                `json:"points"`
	Starters       []string               `json:"starters"`
	StartersPoints []float64              `json:"starters_points"`
	Players        []string               `json:"players"`
	PlayersPoints  map[string]float64     `json:"players_points"`
	CustomPoints   map[string]float64     `json:"custom_points"`
}

// BracketMatchup represents a playoff bracket matchup from Sleeper's bracket API
type BracketMatchup struct {
	MatchupID    int                    `json:"m"`     // matchup number
	Round        int                    `json:"r"`     // round (1=quarterfinals, 2=semifinals, 3=finals)
	Winner       int                    `json:"w"`     // winner roster ID
	Loser        int                    `json:"l"`     // loser roster ID
	Team1        int                    `json:"t1"`    // team 1 roster ID
	Team2        int                    `json:"t2"`    // team 2 roster ID
	PlayoffWeek  *int                   `json:"p,omitempty"` // playoff week (when specified)
	Team1From    map[string]interface{} `json:"t1_from,omitempty"` // where team1 came from
	Team2From    map[string]interface{} `json:"t2_from,omitempty"` // where team2 came from
}

// Transaction represents a league transaction
type Transaction struct {
	TransactionID string                 `json:"transaction_id"`
	Type          string                 `json:"type"`
	Status        string                 `json:"status"`
	StatusUpdated int64                  `json:"status_updated"`
	Created       int64                  `json:"created"`
	RosterIDs     []int                  `json:"roster_ids"`
	Settings      map[string]interface{} `json:"settings"`
	Metadata      map[string]interface{} `json:"metadata"`
	Adds          map[string]int         `json:"adds"`
	Drops         map[string]int         `json:"drops"`
	DraftPicks    []DraftPick            `json:"draft_picks"`
	WaiverBudget  []WaiverBudget         `json:"waiver_budget"`
}

// DraftPick represents a draft pick in a transaction
type DraftPick struct {
	Season   string `json:"season"`
	Round    int    `json:"round"`
	RosterID int    `json:"roster_id"`
	Previous int    `json:"previous_owner_id"`
	OwnerID  int    `json:"owner_id"`
}

// WaiverBudget represents waiver budget changes
type WaiverBudget struct {
	Sender   int `json:"sender"`
	Receiver int `json:"receiver"`
	Amount   int `json:"amount"`
}

// APIResponse represents the standard response format for our tools
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Summary   string      `json:"summary"`
	Error     string      `json:"error,omitempty"`
	Metadata  Metadata    `json:"metadata"`
}

// Metadata contains response metadata
type Metadata struct {
	Timestamp    time.Time `json:"timestamp"`
	Source       string    `json:"source"`
	CacheHit     bool      `json:"cache_hit"`
	APICallsUsed int       `json:"api_calls_used"`
	LeagueID     string    `json:"league_id,omitempty"`
}

// SleeperError represents an error from the Sleeper API
type SleeperError struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
	LeagueID   string `json:"league_id,omitempty"`
}

func (e *SleeperError) Error() string {
	return e.Message
}