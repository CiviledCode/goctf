package ctf

import (
	"errors"
)

var ErrUserNotFound error = errors.New("User not found.")

var ErrPipeThresholdExceed error = errors.New("The pipe didn't receive the content within the required threshold.")

var ErrPipeClosed error = errors.New("The pipe is closed.")

// User holds all the credentials and information about a client connection.
type User struct {
	// Aliase is the display name of the user.
	Aliase string

	// ID is a publicly referencable unique identifier.
	ID string

	// Token is the private token of the user.
	Token string

	// Pipe is the pipe used to send messages through the game webhook.
	Pipe chan []byte

	// Team is the team the user is currently within.
	Team *Team
}

// NewUser creates a new user with the Aliase provided.
func newUser(aliase string, room *Room) *User {
	u := &User{
		Aliase: aliase,
	}

	// Generate and check if the token is unique.
	for {
		token := randomKey(16, false)

		if room.UserByPrivate(token) == nil {
			u.Token = token
			break
		}
	}

	// Generate and check if the ID is unique.
	for {
		id := randomKey(8, false)

		if room.Users[id] == nil {
			u.ID = id
			break
		}
	}

	return u
}

// JoinTeam attempts to join a team.
// If the team is full, ErrTeamTooBig is returned.
func (u *User) JoinTeam(team *Team) error {
	if team.Room.Config.MaxTeamSize <= len(team.UserScores) {
		return ErrTeamTooBig
	}

	// Once in a team, you shouldn't be able to leave.
	if u.Team != nil {
		return nil
	}

	team.UserScores[u.ID] = 0

	u.Team = team

	return nil
}
