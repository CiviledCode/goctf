package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/civiledcode/goctf/ctf"
)

var server *http.Server
var Started bool

var joinTemplate *template.Template
var teamTemplate *template.Template

func Start(ip string, port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/scores", ScoresHandler)
	mux.HandleFunc("/submit", SubmitHandler)
	mux.HandleFunc("/join", JoinHandler)
	mux.HandleFunc("/team", TeamHandler)
	mux.HandleFunc("/play", PlayHandler)
	mux.HandleFunc("/jointeam", JoinTeamHandler)
	mux.HandleFunc("/createteam", CreateTeamHandler)

	server = &http.Server{Addr: fmt.Sprintf("%v:%v", ip, port), Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Server Connection Error: %v\n", err)
		}
	}()

	Started = true
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	joinTemplate = template.Must(template.ParseFiles(wd + "/style/join.html"))
	teamTemplate = template.Must(template.ParseFiles(wd + "/style/team.html"))
}

func Stop() {
	err := server.Close()

	if err != nil {
		log.Printf("Server Close Error: %v\n", err)
	}

	Started = false
}

func ScoresHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Serve scores as a json map mapping each team uuid to a struct
	// This struct should contain the teams name, team members, and another map mapping the start time offset (in seconds) to the amount of points scored at that exact time.

	// This should only be called by the fetch API
}

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Accept all POST requests and attempt to solve a question as a user.
	// This should contain the ID of the question and the answer being submitted.
}

type joinData struct {
	// HideInvalidLogin depicts if the login invalid dialogue should be hidden or not.
	HideInvalidLogin bool

	// ErrorMessage is the message that we are showing within the error dialogue.
	ErrorMessage string
}

// Join allows users to create users within a room given a room id and aliase.
//
// Methods:
// 	- GET: Serves the Join page template.
//
//	- POST: Attempts to join the room using the room code  with the name provided. This will create the room_id and token cookies on the.
//	  Params: room_code, aliase
func JoinHandler(w http.ResponseWriter, r *http.Request) {
	var data joinData

	// Check if the user already has the proper room code and token.
	cookie, err := r.Cookie("room_code")
	if err == nil {
		if room := ctf.Rooms[cookie.Value]; cookie != nil && room != nil {
			cookie, err = r.Cookie("token")
			if err == nil {
				if user := room.UserByPrivate(cookie.Value); cookie != nil && user != nil {
					// Redirect to the main play handler as the user is already authenticated properly in a room.
					http.Redirect(w, r, "/play", 303)
				}
			}
		}
	}

	// The client is attempting to join some sort of room.
	if r.Method == "POST" {
		r.ParseForm()

		data = joinData{}

		// Retrieve the aliase from the form, sanitize it, and check to see if it's valid.
		aliase, ok := r.Form["aliase"]
		if len(aliase) == 0 {
			data.ErrorMessage = "Invalid aliase"
			w.WriteHeader(401)
		}
		sanitized := sanitize(aliase[0])

		if !ok || sanitized == "" {
			data.ErrorMessage = "Invalid aliase"
			w.WriteHeader(401)
		}

		// Retrieve the room code from the form and validate it.
		roomcode, ok := r.Form["room_code"]

		if !ok {
			data.ErrorMessage = "Invalid room code"
			w.WriteHeader(401)
		}

		// If there was no error, create the user.
		if data.ErrorMessage == "" {
			room := ctf.Rooms[roomcode[0]]
			if room == nil {
				data.ErrorMessage = "Invalid room code"
				w.WriteHeader(401)
				joinTemplate.Execute(w, data)
				return
			}

			// Create the user using the sanitized aliase.
			user := room.CreateUser(sanitized)

			// Store the token and room_code as cookies.
			cookie = &http.Cookie{
				Name:  "token",
				Value: user.Token,
			}
			http.SetCookie(w, cookie)

			cookie = &http.Cookie{
				Name:  "room_code",
				Value: roomcode[0],
			}
			http.SetCookie(w, cookie)

			// Redirect to the play path because we are successfully authenticated.
			http.Redirect(w, r, "/play", 303)
		} else {
			log.Printf("Failed attempt to join room '%v' with aliase '%v'. Error: '%v'\n", roomcode[0], sanitized, data.ErrorMessage)
		}

		// The client is attempting to retrieve the webpage to join.
	} else if r.Method == "GET" {
		// TODO: Create the error message dialogue inside the html.
		data = joinData{
			HideInvalidLogin: true,
			ErrorMessage:     "",
		}
		// An unknown method has occured.
	} else {
		w.WriteHeader(405)
	}

	joinTemplate.Execute(w, data)
}

// Team allows players to select between creating and joining a team in one place.
//
// Methods:
//	- GET: Serves the team template and shows the forms to join or create a team.
func TeamHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("room_code")
	if err == nil {
		if room := ctf.Rooms[cookie.Value]; cookie != nil && room != nil {
			cookie, err = r.Cookie("token")
			if err == nil {
				if user := room.UserByPrivate(cookie.Value); cookie != nil && user != nil {
					if user.Team != nil {
						http.Redirect(w, r, "/play", 303)
						return
					}
					// Serve the page because the user is authenticated.
					teamTemplate.Execute(w, nil)
					return
				}
			}
		}
	}

	// The user isn't properly authenticated, so we need to redirect them to the join page.
	http.Redirect(w, r, "/join", 303)

}

func PlayHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user already has the proper room code and token.
	var room *ctf.Room
	var user *ctf.User

	cookie, err := r.Cookie("room_code")
	if err == nil {
		if room = ctf.Rooms[cookie.Value]; cookie != nil && room != nil {
			cookie, err = r.Cookie("token")
			if err == nil {
				if user = room.UserByPrivate(cookie.Value); cookie == nil || user == nil {
					http.Redirect(w, r, "/join", 303)
					return
				}

				if user.Team == nil {
					http.Redirect(w, r, "/team", 303)
					return
				}
			} else {
				http.Redirect(w, r, "/join", 303)
				return
			}
		} else {
			http.Redirect(w, r, "/join", 303)
			return
		}
	} else {
		http.Redirect(w, r, "/join", 303)
		return
	}

	// TODO: Serve webpage capable of fetching questions from /questions

}

func JoinTeamHandler(w http.ResponseWriter, r *http.Request) {

}

// CreateTeam creates a new team using the name received and adds the user to this team.
// If the team name isn't unique and config.ForceUniqueTeams is true, this will redirect back to teams.
//
// Methods:
// 	- POST: Receives the team name attempts to create a team with it, ultimately adding the user to it.
//	  Params: team_name
func CreateTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		cookie, err := r.Cookie("room_code")
		if err == nil {
			if room := ctf.Rooms[cookie.Value]; cookie != nil && room != nil {
				cookie, err = r.Cookie("token")
				if err == nil {
					if user := room.UserByPrivate(cookie.Value); cookie != nil && user != nil {
						if user.Team == nil {
							err = r.ParseForm()
							if err != nil {
								panic(err)
							}
							teamName := r.Form["team_name"]
							if len(teamName) == 0 {
								http.Redirect(w, r, "/team", 303)
								return
							}

							sanitized := sanitize(teamName[0])
							if sanitized != "" {
								team, err := room.CreateTeam(sanitized)
								if err != nil {
									http.Redirect(w, r, "/team", 303)
									return
								}

								err = user.JoinTeam(team)
								if err != nil {
									http.Redirect(w, r, "/team", 303)
									return
								}

								fmt.Printf("Team %v Created by %v", *team, user.Aliase)
								http.Redirect(w, r, "/play", 303)
								return
							}
						}
					}
				}
			}
		}
	}
	// The user isn't properly authenticated, so we need to redirect them to the join page.
	http.Redirect(w, r, "/join", 303)
}

func sanitize(input string) string {
	// TODO: Properly sanitize input.
	// TODO: Add a profanity filter config option
	if input == "" {
		return ""
	}
	return input
}
