package config

// Hint represents a hint for a question that can be unlocked for a team.
// When unlocking a hint for a team, 'Cost' points are deducted from the team balance.
type Hint struct {
	// Content is the content of the hint.
	Content string `json:"content"`

	// Cost is the amount that the hint costs to purchase.
	Cost int64 `json:"cost"`
}
