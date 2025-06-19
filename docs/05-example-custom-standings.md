# Fantasy Football Standings Calculation

This document explains how team standings are calculated in our fantasy football league system. Understanding these calculations helps explain why teams are ranked in their positions and what determines playoff seeding.

## Overview

The standings calculation differs depending on whether the league is currently in progress (regular season) or completed (post-season). Regular season standings determine playoff seeding, while post-season standings reflect final league placement based on playoff performance.

## Regular Season Standings

During the regular season, teams are ranked based on their performance in regular season matchups only. Playoff games are excluded from these calculations.

### Primary Ranking Criteria

Teams are primarily ranked by **wins** in descending order (most wins first).

### Tiebreaker System

When teams have the same number of wins, the following tiebreakers are applied in order:

#### 1. Head-to-Head Record
- Among tied teams, calculate total head-to-head wins against other teams in the tie
- The team with the most head-to-head wins against the other tied teams is ranked higher
- Example: If Teams A, B, and C all have 8 wins, but Team A beat both B and C during the season (2 H2H wins), Team B beat C but lost to A (1 H2H win), and Team C lost to both A and B (0 H2H wins), the order would be: A, B, C

#### 2. Points For (Total Points Scored)
- If head-to-head records are tied, the team with more total points scored throughout the season is ranked higher
- This rewards consistent offensive performance

#### 3. Points Against (Total Points Allowed)
- If Points For are also tied, the team that allowed fewer points is ranked higher
- This accounts for both strength of schedule and defensive performance (though defense is less relevant in fantasy)

#### 4. Random Tiebreaker (Coin Flip)
- As a final tiebreaker, a random determination is made
- This is extremely rare but ensures no two teams can have identical standings

### Calculation Process

1. **Regular Season Data Collection**: Only non-playoff matchups are considered
2. **Record Calculation**: For each team, calculate wins, losses, ties, total points for, and total points against
3. **Head-to-Head Matrix**: Track wins by each team against every other team
4. **Grouping**: Group teams by number of wins
5. **Tiebreaking**: Within each win group, apply tiebreakers to determine order
6. **Final Ordering**: Combine all groups in descending order of wins

### Playoff Qualification

- **Top 6 teams** make the playoffs
- **Seeds 1-2**: Receive first-round byes
- **Seeds 3-6**: Play in the first round (quarterfinals)

## Post-Season Standings (Final Rankings)

Once the league is complete, the final standings are determined by playoff performance rather than regular season record. This means a team with a worse regular season record can finish ahead of teams with better regular season records if they perform better in the playoffs.

### Final Ranking Structure

#### Positions 1-2: Championship Game Results
- **1st Place**: Winner of the championship game (finals)
- **2nd Place**: Loser of the championship game (finals)

#### Positions 3-4: Third Place Game Results
- **3rd Place**: Winner of the third-place game
- **4th Place**: Loser of the third-place game

*Note: The third-place game is played between the two teams that lost in the semifinals.*

#### Positions 5-6: Quarterfinal Losers
- **5th & 6th Place**: Teams that lost in the quarterfinals (first round of playoffs)
- These two teams are ranked against each other using the same tiebreaker system as regular season standings
- The quarterfinal loser with the better regular season record (using the tiebreaker system) gets 5th place

#### Positions 7-12: Non-Playoff Teams
- **7th-12th Place**: Teams that did not make the playoffs
- Ranked using regular season standings (same system as described above)
- These positions do not change once playoffs begin

### Playoff Bracket Structure

The playoffs follow this structure:
- **Quarterfinals**: Seeds 3 vs 6, Seeds 4 vs 5 (Seeds 1 & 2 have byes)
- **Semifinals**: Seed 1 vs lowest remaining seed, Seed 2 vs highest remaining seed
- **Championship Game**: Winners of semifinals play for 1st/2nd place
- **Third Place Game**: Losers of semifinals play for 3rd/4th place

## Example Scenarios

### Regular Season Tiebreaker Example

Three teams tied at 8-4:
- **Team A**: 8-4, 1,500 PF, 1,400 PA, beat Team B and Team C (2 H2H wins)
- **Team B**: 8-4, 1,520 PF, 1,350 PA, beat Team C, lost to Team A (1 H2H win)  
- **Team C**: 8-4, 1,480 PF, 1,450 PA, lost to both Team A and B (0 H2H wins)

**Result**: Team A (1st), Team B (2nd), Team C (3rd) based on head-to-head record.

### Post-Season Ranking Change Example

Regular season top 6:
1. Team A (10-2)
2. Team B (9-3) 
3. Team C (8-4)
4. Team D (8-4)
5. Team E (7-5)
6. Team F (6-6)

Playoff results:
- **Quarterfinals**: Team C beats Team F, Team E beats Team D
- **Semifinals**: Team C beats Team A, Team B beats Team E  
- **Championship**: Team C beats Team B
- **Third Place**: Team A beats Team E

**Final standings**:
1. Team C (champion, was 3rd seed)
2. Team B (runner-up, was 2nd seed)
3. Team A (3rd place winner, was 1st seed)
4. Team E (3rd place loser, was 5th seed)
5. Team D (quarterfinal loser with better regular season record)
6. Team F (quarterfinal loser with worse regular season record)
7-12. Non-playoff teams in regular season order

This example shows how playoff performance can significantly change final rankings compared to regular season standings.

## Technical Implementation Notes

- **Data Sources**: Standings are calculated from matchup data stored in the database
- **Real-time Updates**: Regular season standings update after each completed week
- **Playoff Validation**: The system validates that playoff bracket data is complete and consistent before calculating final standings
- **Tie Handling**: The system properly handles tied games (though rare in fantasy football)

## Frequently Asked Questions

**Q: Why did Team X finish behind Team Y even though they had a better record?**
A: In final standings, playoff performance determines positions 1-6. Regular season record only matters for playoff seeding and final ranking of teams within the same playoff outcome (like quarterfinal losers).

**Q: How are head-to-head records calculated in multi-team ties?**
A: Each team's wins against other teams in the tie are counted. If Teams A, B, and C are tied, we count A's wins vs B and C, B's wins vs A and C, etc.

**Q: What happens if two teams never played each other?**
A: Head-to-head wins would be 0 for both teams against each other, so tiebreaking would move to the next criteria (Points For).

**Q: Can standings change after the regular season ends?**
A: Once playoffs begin, positions 7-12 are locked based on regular season performance. Only positions 1-6 change based on playoff results.