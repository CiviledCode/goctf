package ctf

import (
	"encoding/json"
	"errors"

	"github.com/civiledcode/goctf/ctf/config"
)


var ErrTeamTooBig error = errors.New("Team size exceeds configured max size.") 

var ErrAlreadyInTeam error = errors.New("You are already in a team.")

var ErrTeamNotFound error = errors.New("Team not found.")

var ErrCannotAfford error = errors.New("Your team has insufficient points to make this purchase.")

var ErrHintAlreadyOwned error = errors.New("Your team already owns this hint.")

var ErrQuestionNotFound error = errors.New("Question not found with that id.")

var ErrQuestionAlreadyAnswered error = errors.New("Your team has already answered this question.")

// Team represents a group of users that are scored and displayed amongst the leaderboard.
// Correct answer points are awarded to teams, but individual contributions are mapped too.
type Team struct {
	// Name is the display name of the team.
	Name string

	// JoinCode is a private unique key used for joining the team.
	JoinCode string
	
	// ID is a public key used for identifying a team.
	ID string

	// Room is a reference to the game the team belongs to.
	Room *Room

	// score represents the teams accumulative score.
	score int64

	// CompletedQuestions maps the questions UUID to the UUID of the user who completed it.
	CompletedQuestions map[string]string

	// OwnedHints holds the question IDs for the hints owned.
	OwnedHints map[string]string

	// userScores holds all the current scores for the 
	UserScores map[string]int64

	// deductions is the amount of points deducted from their score.
	deductions int64

	// modified depicts if the data has been modified since the last time data was calculated.
	modified bool

	// teamData is a cache for a JSON representation of the team that's sent to the client.
	teamData string
}

// Score represents the teams final score, including deductions.
func (t *Team) Score() int64 {
	return t.score - t.deductions
}

// Points returns the amount of points earned. This doesn't take into account deductions.
func (t *Team) Points() int64 {
	return t.score
}

// Award adds more points to the teams total.
func (t *Team) Award(amount int64) {
	t.modified = true
	t.score += amount
}

// Deductions returns the amount of points that are being deducted from the score.
func (t *Team) Deductions() int64 {
	return t.deductions
}

// Deduct adds more to the deductions.
func (t *Team) Deduct(amount int64) {
	t.modified = true
	t.deductions += amount
}



// IsComplete depicts if the team has a question completed or not.
// If completed, the users ID  who completed it is also returned.
func (t *Team) IsComplete(questionID string) (bool, string) {
	complete := t.CompletedQuestions[questionID]
	return complete != "", complete
}

// Complete marks a question complete for the team and maps the user who answered it correctly.
// This will update the teams score with the questions point value.
// If the question cannot be found, ErrQuestionNotFound is returned.
// If the question has already been answered, ErrQuestionAlreadyAnswered is returned.
func (t *Team) Complete(userid, questionid string) error {
	question := t.Room.Questions[questionid]

	if question == (config.Question{}) {
		return ErrQuestionNotFound
	}

	if t.CompletedQuestions[questionid] != "" {
		return ErrQuestionAlreadyAnswered
	}

	t.CompletedQuestions[questionid] = userid

	t.UserScores[userid] += question.Points

	t.score += question.Points

	t.modified = true

	return nil
}

// BuyHint attempts to buy a hint using the teams points.
// If the team doesn't have enough points, ErrCannotAfford is returned.
// If the team already owns the hint, ErrHintAlreadyOwned is returned.
func (t *Team) BuyHint(userid, questionid string) error {
	owned, _ := t.OwnsHint(questionid)

	if owned {
		return ErrHintAlreadyOwned	
	}

	question := t.Room.Questions[questionid]

	if question != (config.Question{}) {
		if t.score >= question.HintCost {
			t.OwnedHints[questionid] = userid
			t.deductions += question.HintCost
			t.modified = true
		} else {
			return ErrCannotAfford
		}
	}

	return nil
}


// OwnsHint depicts if a hint has been purchased, and if so by who.
func (t *Team) OwnsHint(questionid string) (bool, string) {
	owner := t.OwnedHints[questionid]
	return owner != "", owner
}

// Data parses private team data for the team users.
// This data is cached until a modification happens.
//
// JSON Structure:
// name string | The name of the team.
// points int | The amount of points earned (without deductions)
// users map[string]int | Mapping of all the teams users to their points earned (without deductions)
// deductions int | The amount of points deducted from the teams final amount.
// completed map[string]string | Mapping of all completed question IDs to the userID who completed them.
// hints map[string]string | Mapping of ids of all owned question hints to the hint content.
func (t *Team) Data() string {
	if !t.modified {
		return t.teamData
	}

	teamData := map[string]interface{}{
		"name": t.Name,
		"points": t.score,
		"deductions": t.deductions,
		"completed": t.CompletedQuestions,
		"users": t.UserScores,
	}

	hints := make(map[string]string)
	for hintQuestionID, _ := range t.OwnedHints {
		question := t.Room.Questions[hintQuestionID]

		if question != (config.Question{}) {
			hints[hintQuestionID] = question.Hint
		}
	}
	teamData["hints"] = hints

	content, _ := json.Marshal(teamData)
	t.modified = false
	t.teamData = string(content)

	return t.teamData
}
