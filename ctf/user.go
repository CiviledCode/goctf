package ctf

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
)

var ErrUserNotFound error = errors.New("User not found.") 
var Random *rand.Rand

var RandomString = strings.Split("a b c d e f g h i j k l m n o p q r s t u v w x y z A B C D E F G H I J K L M N O P Q R S T U V W X Y Z 1 2 3 4 5 6 7 8 9", " ")

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

// randomKey generates a new random string using the constraints passed in.
func randomKey(length int, numbersOnly bool) string {
	var key string
	if !numbersOnly {
		for i := 0; i < length; i++ {			
			key += RandomString[Random.Int() % len(RandomString)]	
		}
	} else {
		for i := 0; i < length; i++ {
			val :=  Random.Int() % 10
			key += fmt.Sprintf("%v", val)
		}
	}

	return key
}
