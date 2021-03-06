package config

type Config struct {
	// ConfigID is the a referencable ID for this config.
	ConfigID string `json:"config_name"`

	// MaxTeams is the max amount of teams allowed in the room.
	MaxTeams int `json:"max_teams"`

	// MaxTeamSize is the max amount of users allowed in a single team.
	MaxTeamSize int `json:"max_team_size"`

	// ForceUniqueTeams depicts if team names need to be unique or not.
	ForceUniqueTeams bool `json:"force_unique_teams"`

	// GameLength is the amount of minutes that the game runs for.
	GameLength int `json:"game_length"`

	// Questions is all the questions being asked.
	Questions []Question `json:"questions"`

	// FlagPlaceholder is a piece of text that shows an example flag.
	FlagPlaceholder string `json:"flag_placeholder"`
}
