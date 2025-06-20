package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// LeagueSettings represents the configuration for a specific league
type LeagueSettings struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Custom      CustomStandings `json:"custom_standings"`
}

// CustomStandings represents custom standings configuration
type CustomStandings struct {
	Enabled       bool     `json:"enabled"`
	Instructions  string   `json:"instructions"`
	TiebreakOrder []string `json:"tiebreak_order"`
	Notes         string   `json:"notes"`
}

// LeagueConfig represents the entire league configuration file
type LeagueConfig struct {
	Instructions    string                    `json:"_instructions,omitempty"`
	Leagues         map[string]LeagueSettings `json:"leagues"`
	DefaultSettings LeagueSettings            `json:"default_settings"`
	Template        map[string]LeagueSettings `json:"_template,omitempty"`
}

// LoadLeagueSettings loads league configuration from the settings file
func LoadLeagueSettings() (*LeagueConfig, error) {
	// Try to find the config file - first check relative to working directory
	configPaths := []string{
		"configs/league_settings.json",
		"../configs/league_settings.json",
		"../../configs/league_settings.json",
	}
	
	var configData []byte
	var foundPath string
	
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			var readErr error
			configData, readErr = os.ReadFile(path)
			if readErr == nil {
				foundPath = path
				break
			}
		}
	}
	
	if foundPath == "" {
		// If no config file found, return default configuration
		return &LeagueConfig{
			Leagues: make(map[string]LeagueSettings),
			DefaultSettings: LeagueSettings{
				Name:        "Default League",
				Description: "League with standard Sleeper tiebreakers",
				Custom: CustomStandings{
					Enabled:       false,
					Instructions:  "Use Sleeper default tiebreakers (wins, then points for)",
					TiebreakOrder: []string{"wins", "points_for"},
					Notes:         "Standard Sleeper tiebreaker rules apply",
				},
			},
		}, nil
	}
	
	var config LeagueConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse league settings from %s: %w", foundPath, err)
	}
	
	return &config, nil
}

// GetLeagueSettings returns settings for a specific league ID
func (c *LeagueConfig) GetLeagueSettings(leagueID string) LeagueSettings {
	if settings, exists := c.Leagues[leagueID]; exists {
		return settings
	}
	
	// Return default settings if league not found
	return c.DefaultSettings
}

// HasCustomStandings checks if a league has custom standings enabled
func (c *LeagueConfig) HasCustomStandings(leagueID string) bool {
	settings := c.GetLeagueSettings(leagueID)
	return settings.Custom.Enabled
}

// GetCustomInstructions returns the custom instructions for a league
func (c *LeagueConfig) GetCustomInstructions(leagueID string) string {
	settings := c.GetLeagueSettings(leagueID)
	if settings.Custom.Enabled && settings.Custom.Instructions != "" {
		return settings.Custom.Instructions
	}
	
	return ""
}