package handlers

import (
	"fmt"

	"github.com/sam-maryland/sleeper-mcp-server/internal/sleeper"
)

// analyzeRosterStrength performs simple roster strength analysis
func analyzeRosterStrength(roster *sleeper.Roster, allPlayers map[string]sleeper.Player) RosterAnalysis {
	// Count players by position
	positionCounts := make(map[string]struct {
		starters int
		bench    int
		injured  int
	})

	// Analyze starters
	for _, playerID := range roster.Starters {
		if playerID == "" || playerID == "0" {
			continue
		}
		
		player, exists := allPlayers[playerID]
		if !exists {
			continue
		}
		
		pos := player.Position
		if pos == "" {
			pos = "UNKNOWN"
		}
		
		counts := positionCounts[pos]
		counts.starters++
		
		// Check injury status
		if player.InjuryStatus != "" && player.InjuryStatus != "Healthy" {
			counts.injured++
		}
		
		positionCounts[pos] = counts
	}

	// Analyze bench
	starterSet := make(map[string]bool)
	for _, starterID := range roster.Starters {
		starterSet[starterID] = true
	}

	for _, playerID := range roster.Players {
		if playerID == "" || playerID == "0" || starterSet[playerID] {
			continue
		}
		
		player, exists := allPlayers[playerID]
		if !exists {
			continue
		}
		
		pos := player.Position
		if pos == "" {
			pos = "UNKNOWN"
		}
		
		counts := positionCounts[pos]
		counts.bench++
		
		// Check injury status
		if player.InjuryStatus != "" && player.InjuryStatus != "Healthy" {
			counts.injured++
		}
		
		positionCounts[pos] = counts
	}

	// Build positional analysis
	var positionalBreakdown []PositionalAnalysis
	var insights []string
	
	positions := []string{"QB", "RB", "WR", "TE", "K", "DEF"}
	
	for _, pos := range positions {
		counts := positionCounts[pos]
		
		// Determine strength rating
		var strength string
		totalPlayers := counts.starters + counts.bench
		
		switch pos {
		case "QB":
			if totalPlayers >= 3 {
				strength = "Strong"
			} else if totalPlayers >= 2 {
				strength = "Adequate"
			} else {
				strength = "Weak"
				insights = append(insights, fmt.Sprintf("Consider adding more %s depth", pos))
			}
		case "RB", "WR":
			if totalPlayers >= 5 {
				strength = "Strong"
			} else if totalPlayers >= 3 {
				strength = "Adequate"
			} else {
				strength = "Weak"
				insights = append(insights, fmt.Sprintf("Needs more %s depth", pos))
			}
		case "TE":
			if totalPlayers >= 2 {
				strength = "Strong"
			} else if totalPlayers >= 1 {
				strength = "Adequate"
			} else {
				strength = "Weak"
				insights = append(insights, fmt.Sprintf("Missing %s coverage", pos))
			}
		default:
			if totalPlayers >= 1 {
				strength = "Adequate"
			} else {
				strength = "Weak"
			}
		}

		if counts.injured > 0 {
			insights = append(insights, fmt.Sprintf("%d %s player(s) have injury concerns", counts.injured, pos))
		}

		positionalBreakdown = append(positionalBreakdown, PositionalAnalysis{
			Position:       pos,
			StarterCount:   counts.starters,
			BenchCount:     counts.bench,
			Strength:       strength,
			InjuryConcerns: counts.injured,
		})
	}

	// Determine overall rating
	strongPositions := 0
	weakPositions := 0
	
	for _, analysis := range positionalBreakdown {
		switch analysis.Strength {
		case "Strong":
			strongPositions++
		case "Weak":
			weakPositions++
		}
	}

	var overallRating string
	if strongPositions >= 4 && weakPositions == 0 {
		overallRating = "Excellent"
	} else if strongPositions >= 3 && weakPositions <= 1 {
		overallRating = "Strong"
	} else if strongPositions >= 2 && weakPositions <= 2 {
		overallRating = "Average"
	} else {
		overallRating = "Needs Improvement"
	}

	// Add performance insights
	totalGames := roster.Settings.Wins + roster.Settings.Losses + roster.Settings.Ties
	if totalGames > 0 {
		ppg := roster.Settings.FPTS / float64(totalGames)
		if ppg > 120 {
			insights = append(insights, "High-scoring team with strong offensive production")
		} else if ppg < 90 {
			insights = append(insights, "Below-average scoring - may need offensive upgrades")
		}
	}

	return RosterAnalysis{
		OverallRating:       overallRating,
		PositionalBreakdown: positionalBreakdown,
		Insights:            insights,
	}
}

// compareRostersByPosition compares two rosters position by position
func compareRostersByPosition(roster1, roster2 *sleeper.Roster, allPlayers map[string]sleeper.Player) RosterComparison {
	// Get positional breakdowns for both teams
	analysis1 := analyzeRosterStrength(roster1, allPlayers)
	analysis2 := analyzeRosterStrength(roster2, allPlayers)

	var positionalComparisons []PositionalComparison
	var summary []string
	
	team1Advantages := 0
	team2Advantages := 0

	// Compare each position
	positions := []string{"QB", "RB", "WR", "TE", "K", "DEF"}
	
	for _, pos := range positions {
		var pos1, pos2 PositionalAnalysis
		
		// Find position data for each team
		for _, p := range analysis1.PositionalBreakdown {
			if p.Position == pos {
				pos1 = p
				break
			}
		}
		
		for _, p := range analysis2.PositionalBreakdown {
			if p.Position == pos {
				pos2 = p
				break
			}
		}

		// Determine advantage
		var advantage string
		total1 := pos1.StarterCount + pos1.BenchCount
		total2 := pos2.StarterCount + pos2.BenchCount
		
		if total1 > total2 {
			advantage = fmt.Sprintf("Team %d", roster1.RosterID)
			team1Advantages++
		} else if total2 > total1 {
			advantage = fmt.Sprintf("Team %d", roster2.RosterID)
			team2Advantages++
		} else {
			advantage = "Even"
		}

		positionalComparisons = append(positionalComparisons, PositionalComparison{
			Position:      pos,
			Team1Starters: pos1.StarterCount,
			Team1Bench:    pos1.BenchCount,
			Team2Starters: pos2.StarterCount,
			Team2Bench:    pos2.BenchCount,
			Advantage:     advantage,
		})
	}

	// Determine overall advantage
	var overallAdvantage string
	if team1Advantages > team2Advantages {
		overallAdvantage = fmt.Sprintf("Team %d", roster1.RosterID)
		summary = append(summary, fmt.Sprintf("Team %d has depth advantages in %d positions", roster1.RosterID, team1Advantages))
	} else if team2Advantages > team1Advantages {
		overallAdvantage = fmt.Sprintf("Team %d", roster2.RosterID)
		summary = append(summary, fmt.Sprintf("Team %d has depth advantages in %d positions", roster2.RosterID, team2Advantages))
	} else {
		overallAdvantage = "Even"
		summary = append(summary, "Teams are evenly matched in positional depth")
	}

	// Compare performance
	totalGames1 := roster1.Settings.Wins + roster1.Settings.Losses + roster1.Settings.Ties
	totalGames2 := roster2.Settings.Wins + roster2.Settings.Losses + roster2.Settings.Ties
	
	if totalGames1 > 0 && totalGames2 > 0 {
		ppg1 := roster1.Settings.FPTS / float64(totalGames1)
		ppg2 := roster2.Settings.FPTS / float64(totalGames2)
		
		if ppg1 > ppg2 + 5 {
			summary = append(summary, fmt.Sprintf("Team %d averages %.1f more points per game", roster1.RosterID, ppg1-ppg2))
		} else if ppg2 > ppg1 + 5 {
			summary = append(summary, fmt.Sprintf("Team %d averages %.1f more points per game", roster2.RosterID, ppg2-ppg1))
		}
	}

	return RosterComparison{
		PositionalComparisons: positionalComparisons,
		OverallAdvantage:      overallAdvantage,
		Summary:               summary,
	}
}