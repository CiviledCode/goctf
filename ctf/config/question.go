package config

import (
	"strings"
)

// Question represents a CTF question that can be marshaled from JSON.
type Question struct {
	// Name represents the display name of the question. This has no programmatic significance.
	Name string `json:"name"`

	// Category represents the category of the question and how it will be grouped. This is case sensitive.
	Category string `json:"category"`

	// ID represents the hardcoded version of a questions ID. This should only be set if there may be some programmatic reference to it.
	// This ID isn't private, so it shouldn't be named anything to give it away.
	ID string `json:"id"`

	// Question represents the question that is being asked.
	Question string `json:"question"`

	// Answer represents the valid question answer.
	Answer string `json:"answer"`

	// Hint represents a hint that can be purchased for HintCost amount of points
	// from the team balance.
	Hint string `json:"hint"`

	// Points represents the amount of points gained from completing a question correctly.
	Points int64 `json:"point_gain"`

	// HintCost represents the amount of points deducted from your teams points for taking this hint.
	HintCost int64 `json:"hint_cost"`

	// WrongCost represents the amount of points deducted from your teams points for getting the answer wrong.
	WrongCost int64 `json:"wrong_loss"`

	// CaseSensitive depicts if the answer cares about capitalization.
	CaseSensitive bool `json:"case_sensitive"`

	// RequiredSolved depicts the questions that are needed to be solved before this question can be viewed or answered.
	RequiredSolved []string `json:"required_solved"`
}

// IsRight checks an answer using the proper capitalization configuration and returns the answer.
func (q Question) IsRight(answer string) bool {
	if !q.CaseSensitive {
		return strings.EqualFold(answer, q.Answer)
	}

	return strings.Compare(answer, q.Answer) == 0
}
