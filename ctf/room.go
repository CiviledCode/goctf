package ctf

import (
	"log"
	"strings"
	"time"

	"github.com/civiledcode/goctf/ctf/config"
)

// Rooms maps room codes to their respective room structs. This allows more than one game to be going on at once.
var Rooms map[string]*Room

// Room represents a CTF game or instance.
// This instance is not thread safe
type Room struct {
	// Code represents the unique ID used to join the room.
	Code string

	// Teams maps all the current team ids to their team items.
	Teams map[string]*Team

	// Config represents the configuration used to control the isntance.
	Config config.Config

	// Questions maps the question ids to the question object.
	// Questions should only be stored here to allow for global modification.
	Questions map[string]config.Question

	// Users maps userid to the user object.
	Users map[string]*User

	// AnswerCallback is a callback executed before we attempt to check if the answer is correct.
	// If this returns true, no flag checks will be done and it will automatically mark it as corrrect.
	// If this returns false, the flag will be checked to see if it's correct.
	// The first argument is the user who answered, then the question that's being answered, followed by the answer submitted.
	AnswerCallback Callback

	// CompleteCallback is called after the question has been marked as correct and the points have been given to the team.
	// The return result means nothing.
	// The first argument is the user who answered, then the question that's being answered, followed by the answer submitted.
	CompleteCallback Callback

	// WrongCallback is called after the answer of a question has been deemed as wrong and the points have been deducted (if any).
	// The return result means nothing.
	// The first argument is the user who answered, then the question that's being answered, followed by the answer submitted.
	WrongCallback Callback

	// tokens maps a users private token to their userid.
	tokens map[string]string

	// started depicts if the room has started already. If this is true, no longer accept new members.
	started bool

	// elapsedTime is the amount of time in seconds that has gone on in total outside of the current stretch that is going on currently.
	// startTime is the unix time that this stretch was started.
	elapsedTime, startTime int64
}

type Callback func(*User, config.Question, string) bool

func init() {
	Rooms = make(map[string]*Room)
	log.Println("Instantiated Rooms")
}

// NewRoom creates a new room with a random code using the config passed through.
func NewRoom(con config.Config) *Room {
	r := &Room{
		Code:      randomKey(6, true),
		Teams:     make(map[string]*Team),
		Config:    con,
		Questions: make(map[string]config.Question),
		started:   false,
		tokens:    make(map[string]string),
		Users:     make(map[string]*User),
	}

	for _, question := range con.Questions {
		// If an ID is hardcoded, we should allow it.
		if question.ID != "" {
			r.Questions[question.ID] = question
			continue
		}

		for {
			key := randomKey(4, false)

			if r.Questions[key].Question == "" {
				r.Questions[key] = question
				break
			}
		}
	}

	for {
		code := randomKey(6, true)

		if Rooms[code] == nil {
			r.Code = code
			break
		}
	}

	Rooms[r.Code] = r

	return r
}

// Start will allow answering, buying hints, and viewing questions. The time will also start being elapsed from this point.
func (r *Room) Start() {
	r.startTime = time.Now().Unix()
	r.started = true
}

// Stop will stop the time and disable all answering, buying hints, and viewing questions until started again.
func (r *Room) Stop() {
	r.elapsedTime += time.Now().Unix() - r.startTime
	r.started = false
}

// Started depicts if the room is started or not.
func (r *Room) Started() bool {
	return r.started
}

// CurrentTime calculates the amount of time that has passed that the game room has been started.
func (r *Room) CurrentTime() int64 {
	return (time.Now().Unix() - r.startTime) + r.elapsedTime
}

// RemoveUser removes a user using their userid.
// This will delete their team entry, user entry, and token.
// If removePoints is true, a deduction of the point total gained by the player will be added to the team.
// If the user cannot be found, ErrUserNotFound is returned.
func (r *Room) RemoveUser(userid string, removePoints bool) error {
	user := r.Users[userid]

	if user == nil {
		return ErrUserNotFound
	}

	if user.Team == nil {
		delete(r.tokens, user.Token)
		delete(r.Users, userid)
		return nil
	}

	if removePoints {
		user.Team.Deduct(user.Team.UserScores[userid])
		delete(user.Team.UserScores, userid)
		delete(r.tokens, user.Token)
		delete(r.Users, userid)
	}

	return nil
}

// UserByPrivate retrieves a user using their private token.
// if no user is found, nil is returned.
func (r *Room) UserByPrivate(privateToken string) *User {
	userid := r.tokens[privateToken]

	if userid != "" {
		return r.Users[userid]
	}

	return nil
}

// CreateUser creates a new user with the aliase provided.
// This initiates values in room.Users and room.tokens
func (r *Room) CreateUser(aliase string) *User {
	user := newUser(aliase, r)
	r.Users[user.UUID] = user
	r.tokens[user.Token] = user.UUID
	return user
}

// CreateTeam creates a new empty team with the name provided and adds it to the current room.
func (r *Room) CreateTeam(name string) (*Team, error) {
	if r.Config.ForceUniqueTeams {
		for _, team := range r.Teams {
			if strings.EqualFold(team.Name, name) {
				return nil, ErrTeamNameUsed
			}
		}
	}

	t := &Team{
		Name:               name,
		UserScores:         make(map[string]int64),
		Room:               r,
		CompletedQuestions: make(map[string]config.AnsweredQuestion),
		modified:           true,
	}

outside:
	for {
		// Generate a random join code and verify that it's unique.
		key := randomKey(6, true)

		for _, team := range r.Teams {
			if team.JoinCode == key {
				continue outside
			}
		}

		t.JoinCode = key
		break
	}

	for {
		// Generate a random team ID and verify that it's unique.
		key := randomKey(10, false)

		if r.Teams[key] == nil {
			t.ID = key
			break
		}
	}

	r.Teams[t.ID] = t

	return t, nil
}

// DeleteTeam removes a team and all its members from the room using its ID.
func (r *Room) DeleteTeam(id string) {
	t := r.Teams[id]

	if t == nil {
		return
	}

	for userid := range t.UserScores {
		user := r.Users[userid]
		delete(r.tokens, user.Token)
		delete(r.Users, userid)

	}

	delete(r.Teams, id)
}

// TeamByCode retrieves a team based on its joincode.
func (r *Room) TeamByCode(teamCode string) *Team {
	for _, team := range r.Teams { 
		if team.JoinCode == teamCode {
			return team
		}
	}

	return nil
}

// AnswerQuestion attempts to answer a question on behalf of the user provided.
// The proper point allocations, deductions, answer, and security checks are done on this.
// This returns a bool representing if you have the question answered now and an error.
// This executes the Answer, Complete, and Wrong callbacks respectively.
// If the user isn't found, ErrUserNotFound is returned.
// If the question isn't found, ErrQuestionNotFound is returned.
// If the question was already answered by the users team, true is returned alongside ErrQuestionAlreadyAnswered.
func (r *Room) AnswerQuestion(userid, questionid, answer string) (bool, error) {
	user := r.Users[userid]

	if user == nil {
		return false, ErrUserNotFound
	}

	if user.Team == nil {
		return false, ErrTeamNotFound
	}

	question := r.Questions[questionid]

	if question.Question == "" {
		return false, ErrQuestionNotFound
	}

	if done, _ := user.Team.IsComplete(questionid); done {
		return true, ErrQuestionAlreadyAnswered
	}

	if len(question.RequiredSolved) > 0 {
		for _, required := range question.RequiredSolved {
			if complete, _ := user.Team.IsComplete(required); !complete {
				return false, ErrQuestionRequiredUnsolved
			}
		}
	}

	if r.AnswerCallback != nil {
		if r.AnswerCallback(user, question, answer) {
			user.Team.Complete(userid, questionid)
			if r.CompleteCallback != nil {
				r.CompleteCallback(user, question, answer)
			}
			return true, nil
		}
	}

	if question.IsRight(answer) {
		user.Team.Complete(userid, questionid)
		if r.CompleteCallback != nil {
			r.CompleteCallback(user, question, answer)
		}
		return true, nil
	} else if question.WrongCost >= 1 {
		user.Team.Deduct(question.WrongCost)
	}

	if r.WrongCallback != nil {
		r.WrongCallback(user, question, answer)
	}

	return false, nil
}
