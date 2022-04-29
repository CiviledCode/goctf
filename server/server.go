package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/civiledcode/goctf/ctf"
	"github.com/gorilla/websocket"
)

// TODO: Properly log errors with correct severity for better vulnerability tracing.

var server *http.Server
var Started bool

var joinTemplate *template.Template
var teamTemplate *template.Template
var playTemplate *template.Template

var upgrader = websocket.Upgrader{} // use default options

func Start(ip string, port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/scores", ScoresHandler)
	mux.HandleFunc("/submit", SubmitHandler)
	mux.HandleFunc("/join", JoinHandler)
	mux.HandleFunc("/team", TeamHandler)
	mux.HandleFunc("/play", PlayHandler)
	mux.HandleFunc("/game", GameHandler)
	mux.HandleFunc("/jointeam", JoinTeamHandler)
	mux.HandleFunc("/createteam", CreateTeamHandler)

	server = &http.Server{Addr: fmt.Sprintf("%v:%v", ip, port), Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Server Connection Error: %v\n", err)
		}
	}()

	Started = true
	log.Printf("Server started on %v:%v\n", ip, port)
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	joinTemplate = template.Must(template.ParseFiles(wd + "/style/join.html"))
	teamTemplate = template.Must(template.ParseFiles(wd + "/style/team.html"))
	playTemplate = template.Must(template.ParseFiles(wd + "/style/play.html"))
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

func GameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if token, ok := r.URL.Query()["token"]; ok {
			if room_code, ok := r.URL.Query()["room"]; ok {
				room := ctf.Rooms[room_code[0]]
				user := room.UserByPrivate(token[0])
				if room == nil || user == nil {
					w.WriteHeader(401)
					return
				}

				if !room.Started() {
					w.WriteHeader(403)
					return
				}

				c, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					log.Printf("WebSocket Upgrade Error: %v\n", err)
					return
				}

				defer c.Close()

				user.Pipe = make(chan []byte)
				readChan := make(chan bool)

				go func() {
					for {
						c.ReadMessage()
						readChan <- true
					}
				}()
				go func() {
					time.Sleep(100)
					allIds := make([]string, len(room.Questions))
					i := 0

					for questionid, _ := range room.Questions {
						allIds[i] = questionid
						i++
					}
					fmt.Println(allIds)
					user.Team.UpdateUser(user.ID, allIds...)
				}()
				for {
					select {
					case msg := <-user.Pipe:
						err = c.WriteMessage(1, msg)
						if err != nil {
							log.Printf("WebSocket Write Error: %v\n", err)
							return
						}
					case <-readChan:
						user.Pipe = nil
						return
					}
				}
			}
		}
	}
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

type playData struct {
	GameURL string
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

				data := playData{GameURL: fmt.Sprintf("ws://%v/game?token=%v&room=%v", server.Addr, user.Token, room.Code)}
				playTemplate.Execute(w, data)
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
}

func JoinTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		cookie, err := r.Cookie("room_code")
		if err == nil {
			if room := ctf.Rooms[cookie.Value]; cookie != nil && room != nil {
				cookie, err = r.Cookie("token")
				if err == nil {
					if user := room.UserByPrivate(cookie.Value); cookie != nil && user != nil {
						if user.Team == nil {
							// The user isn't already inside of a team so they can properly join one.
							err = r.ParseForm()
							if err != nil {
								log.Printf("JoinTeam Handler Error: %v\n", err)
							}

							teamcode := r.Form["team_code"]
							if len(teamcode) == 0 {
								http.Redirect(w, r, "/team", 303)
								return
							}

							team := room.TeamByCode(teamcode[0])

							if team == nil {
								http.Redirect(w, r, "/team", 303)
								return
							}

							err = user.JoinTeam(team)
							if err != nil {
								http.Redirect(w, r, "/team", 303)
								return
							}

							http.Redirect(w, r, "/play", 303)
							return
						}
					}
				}
			}
		}
	}
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
								log.Printf("CreateTeam Handler Error: %v\n", err)
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

								log.Printf("Team %v Created by %v", *team, user.ID)
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
