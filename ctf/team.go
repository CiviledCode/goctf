package ctf

import (
	"errors"

	"github.com/civiledcode/goctf/ctf/config"
)


var ErrTeamTooBig error = errors.New("Team size exceeds configured max size.")

var ErrAlreadyInTeam error = errors.New("You are already in a team.")

var ErrTeamNotFound error = errors.New("Team not found.")

var ErrTeamNameUsed error = errors.New("Team name already in use.")

var ErrGameNotStarted error = errors.New("The game hasn't started yet.")

var ErrCannotAfford error = errors.New("Your team has insufficient points to make this purchase.")

var ErrHintAlreadyOwned error = errors.New("Your team already owns this hint.")

var ErrQuestionNotFound error = errors.New("Question not found with that id.")

var ErrQuestionAlreadyAnswered error = errors.New("Your team has already answered this question.")

var ErrQuestionRequiredUnsolved error = errors.New("This question requires other questions to be answered before you can solve it.")

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
	CompletedQuestions map[string]config.AnsweredQuestion

	// OwnedHints maps question ids to an array of all hint ids owned for that question.
	OwnedHints map[string][]int

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
func (t *Team) IsComplete(questionID string) (bool, config.AnsweredQuestion) {
	data, complete := t.CompletedQuestions[questionID]
	return complete, data
}

// Complete marks a question complete for the team and maps the user who answered it correctly.
// This will update the teams score with the questions point value.
// If the question cannot be found, ErrQuestionNotFound is returned.
// If the question has already been answered, ErrQuestionAlreadyAnswered is returned.
func (t *Team) Complete(userid, questionid string) error {
	if question, ok := t.Room.Questions[questionid]; ok {
		if _, ok := t.CompletedQuestions[questionid]; ok {
			return ErrQuestionAlreadyAnswered
		}

		t.CompletedQuestions[questionid] = config.AnsweredQuestion{Solver: userid, SolveTime: t.Room.CurrentTime()}
		t.UserScores[userid] += question.Points
		t.score += question.Points
		t.modified = true

		return nil
	}
	
	return ErrQuestionNotFound
	
}

// BuyHint attempts to buy a hint using the teams points.
// If the team doesn't have enough points, ErrCannotAfford is returned.
// If the team already owns the hint, ErrHintAlreadyOwned is returned.
func (t *Team) BuyHint(questionid string, hintid int) error {
	owned, _ := t.OwnsHint(questionid, hintid)

	if owned {
		return ErrHintAlreadyOwned
	}

	if question, ok := t.Room.Questions[questionid]; ok {
		if hint, ok := question.Hints[hintid]; ok {
			if t.score >= hint.Cost {
				t.OwnedHints[questionid] = append(t.OwnedHints[questionid], hintid)
				t.deductions += hint.Cost
				t.modified = true
			} else {
				return ErrCannotAfford
			}
		}
	}

	return nil
}

// OwnsHint depicts if a team owns a hint given a questionid and an id for the hint.
// If the team owns this hint, this returns true and the hint is returned.
func (t *Team) OwnsHint(questionid string, hintid int) (bool, config.Hint) {
	ownedHints := t.OwnedHints[questionid]
	if ownedHints == nil {
		return false, config.Hint{}
	}

	for _, id := range ownedHints {
		if id == hintid {
			return true, t.Room.Questions[questionid].Hints[hintid]
		}
	}

	return false, config.Hint{}
}

// QuestionData receives a questionid and attempts to convert it into data 
func (t *Team) QuestionData(questionid string) (map[string]interface{}, error) {
	if question, ok := t.Room.Questions[questionid]; ok {
		if len(question.RequiredSolved) > 0 {
			for _, requiredid := range question.RequiredSolved {
				if _, ok = t.CompletedQuestions[requiredid]; !ok {
					return nil, ErrQuestionRequiredUnsolved
				}
			}
		}

		questionData := map[string]interface{} {
			"name": question.Name,
			"category": question.Category,
			"id": question.ID,
			"question": question.Question,
			"type": "question",
			"points": question.Points,
			"wrong_cost": question.WrongCost,
		}

		hints := make(map[int]interface{})

		if len(question.Hints) > 0 {
			for hintid, hint := range question.Hints {
				hintContents := map[string]interface{} {
					"cost": hint.Cost,
				}

				// If the item is owned, add the content field on there too.
				owns, _ := t.OwnsHint(questionid, hintid)
				hintContents["owned"] = owns
				if owns {
					hintContents["content"] = hint.Content
				}

				hints[hintid] = hintContents
			}
		}

		questionData["hints"] = hints

		_, ok = t.CompletedQuestions[questionid]

		questionData["solved"] = ok

		if ok {
			answered := t.CompletedQuestions[question.ID]
			questionData["solver"] = answered.Solver
			questionData["solve_time"] = answered.SolveTime
		}

		return questionData, nil
	} else {
		return nil, ErrQuestionNotFound 
	}
}
