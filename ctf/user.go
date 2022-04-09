package ctf

import (
	"errors"
)

var ErrUserNotFound error = errors.New("User not found.") 

// User holds all the credentials and information about a client connection.
type User struct {
	// Aliase is the display name of the user. 
	Aliase string
	
	// UUID is a publicly referencable unique identifier.
	UUID string

	// Token is the private token of the user.
	Token string

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

	// Generate and check if the UUID is unique.
	for { 
		uuid := randomKey(6, false)

		if room.Users[uuid] == nil {
			u.UUID = uuid
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

	team.UserScores[u.UUID] = 0

	u.Team = team

	return nil
}


