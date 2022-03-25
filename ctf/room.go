package ctf

import (
	"fmt"

	"github.com/civiledcode/goctf/ctf/config"
)


var Rooms map[string]*Room

// Room represents a CTF game or instance.
// This instance is not thread safe
type Room struct {
	// Code represents the unique ID used to join the room.
	Code string

	// Teams represents a list of teams currently within the room.
	Teams map[string]*Team

	// Config represents the configuration used to control the isntance.
	Config config.Config

	// Questions maps the question ids to the question object.
	// Questions should only be stored here to allow for global modification.
	Questions map[string]config.Question

	// Users maps userid to the user object.
	Users map[string]*User

	// tokens maps a users private token to their userid.
	tokens map[string]string

	// started depicts if the room has started already. If this is true, no longer accept new members.
	started bool
}

func init() {
	Rooms = make(map[string]*Room)
	fmt.Println("Initiated rooms")
}

// NewRoom creates a new room with a random code using the config passed through.
func NewRoom(con config.Config) *Room {
	r := &Room{
		Code: randomKey(6, false), 
		Teams: make(map[string]*Team),
		Config: con,
		Questions: make(map[string]config.Question),
		started: false,
		tokens: make(map[string]string),
		Users: make(map[string]*User),
	}

	for _, question := range con.Questions {
		for {
			key := randomKey(4, false)

			if r.Questions[key] == (config.Question{}) {
				r.Questions[key] = question
				break
			}
		}
	}

	for {
		code := randomKey(6, false)

		if Rooms[code] == nil {
			r.Code = code
			break
		}
	}

	Rooms[r.Code] = r

	return r
}

func (r *Room) Start() {
	// TODO: Create some sort of timer that increments the amount of time since competition start.	
}

func (r *Room) Pause() {
	// TODO: Pause the timer, and deny all incoming attempts at solving or viewing the questions.
}

func (r *Room) Started() {

}

func (r *Room) Stop() {
	
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
func (r *Room) CreateTeam(name string) *Team {
	t := &Team{
		Name: name, 
		UserScores: make(map[string]int64),
		Room: r,
		CompletedQuestions: make(map[string]string),
		modified: true,
	}	
	
	outside:
	for {
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
		key := randomKey(10, false)

		if r.Teams[key] == nil {
			t.ID = key
			break
		}
	}

	r.Teams[t.ID] = t

	return t
}

// DeleteTeam removes a team and all its members from the room using its ID.
func (r *Room) DeleteTeam(id string) {
	t := r.Teams[id]

	if t == nil {
		return
	}

	for userid, _ := range t.UserScores {
		user := r.Users[userid]
		delete(r.tokens, user.Token)
		delete(r.Users, userid)
		
	}

	delete(r.Teams, id)
}

// AnswerQuestion attempts to answer a question on behalf of the user provided. 
// The proper point allocations, deductions, answer, and security checks are done on this.
// This returns a bool representing if you have the question answered now and an error.
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

	if question == (config.Question{}) {
		return false, ErrQuestionNotFound 
	}

	if done, _ := user.Team.IsComplete(questionid); done {
		return true, ErrQuestionAlreadyAnswered
	}

	if question.IsRight(answer) {
		user.Team.Complete(userid, questionid)
		return true, nil
	} else if question.WrongCost >= 1 {
		user.Team.Deduct(question.WrongCost)
	}

	return false, nil
}

