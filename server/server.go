package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
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
	mux.HandleFunc("/buyhint", BuyHintHandler)
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

	// TODO: Remove after testing.
	upgrader.CheckOrigin = func(_ *http.Request) bool { return true }
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

type submitData struct {
	QuestionID string `json:"question_id"`

	Answer string `json:"answer"`
}

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		room, user, err := VerifyCredentials(r)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(404)
			return
		}

		var data submitData
		json.NewDecoder(r.Body).Decode(&data)
		solved, err := room.AnswerQuestion(user, data.QuestionID, data.Answer)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(404)
			return
		}

		if solved {
			w.WriteHeader(200)
			return
		} else {
			w.WriteHeader(400)
			return
		}
	}
	w.WriteHeader(404)
}

type buyHintData struct {
	QuestionID string `json:"question_id"`

	HintID string `json:"hint_id"`

	Error string `json:"error_message"`
}

func BuyHintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		room, user, err := VerifyCredentials(r)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(404)
			return
		}

		var data buyHintData
		json.NewDecoder(r.Body).Decode(&data)
		val, err := strconv.ParseInt(data.HintID, 10, 32)
		if err != nil {
			w.WriteHeader(404)
		}
		_, err = room.BuyHint(user, data.QuestionID, int(val))
		if err != nil {
			data.Error = err.Error()
		}
		content, err := json.Marshal(data)
		if err != nil {
			log.Printf("BuyHint Error Marshaling Error: %v\n", err)
		}

		w.Write(content)

	}
}

type joinData struct {
	// HideInvalidLogin depicts if the login invalid dialogue should be hidden or not.
	HideInvalidLogin bool

	// ErrorMessage is the message that we are showing within the error dialogue.
	ErrorMessage string
}

// VerifyCredentials retrieves values from the request and validates them.
// This first retrieves the token and room code from the cookies. If there's a valid token or room code within the headers, it overrides this.
// If no room is found, ErrRoomNotFound is returned. Otherwise ErrUserNotFound is returned.
func VerifyCredentials(r *http.Request) (*ctf.Room, *ctf.User, error) {
	var room_code, token string

	// Get from cookies

	token_cookie, err := r.Cookie("token")
	if err == nil {
		if token_cookie.Value != "" {
			token = token_cookie.Value
		}
	}

	room_cookie, err := r.Cookie("room_code")
	if err == nil {
		if room_cookie.Value != "" {
			room_code = room_cookie.Value
		}
	}

	// Get from headers

	header_room_code := r.Header.Get("room_code")
	header_token := r.Header.Get("token")
	if header_room_code != "" {
		room_code = header_room_code
	}

	if header_token != "" {
		token = header_token
	}

	// Get from URL params

	tokens, ok := r.URL.Query()["token"]
	if ok && len(tokens[0]) > 0 {
		token = tokens[0]
	}

	tokens, ok = r.URL.Query()["room_code"]
	if ok && len(tokens[0]) > 0 {
		room_code = tokens[0]
	}

	// Check credentials

	room := ctf.Rooms[room_code]
	if room == nil {
		return nil, nil, ctf.ErrRoomNotFound
	}

	user := room.UserByPrivate(token)
	if user == nil {
		return nil, nil, ctf.ErrUserNotFound
	}

	return room, user, nil
}

func JoinHandler(w http.ResponseWriter, r *http.Request) {
	var data joinData

	_, _, err := VerifyCredentials(r)
	if err == nil {
		http.Redirect(w, r, "/home", 303)
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

			// Send the token and room code as headers.
			w.Header().Add("token", user.Token)

			w.Header().Add("room_code", roomcode[0])

			// Set the token and room code cookies for later login.
			cookie := &http.Cookie{
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

		room, user, err := VerifyCredentials(r)
		if err != nil {
			w.WriteHeader(401)
			return
		}

		if !room.Started() {
			w.WriteHeader(403)
			return
		}

		// Upgrade the connection.
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket Upgrade Error: %v\n", err)
			return
		}

		defer c.Close()

		user.Pipe = make(chan []byte)

		// Listen for a message on the websocket. If one is received, close the connection.
		go func() {
			for {
				c.ReadMessage()
				user.Pipe = nil
				return
			}
		}()

		// Send the user all the questions the team can view.
		go func() {
			// TODO: This seems hacky and prone to race conditions. Fix it somehow
			time.Sleep(100)
			allIds := make([]string, len(room.Questions))
			i := 0

			for questionid, _ := range room.Questions {
				allIds[i] = questionid
				i++
			}

			user.Team.UpdateUser(user.ID, allIds...)
		}()

		// Listen for a new message from the pipe and relay it through the websocket connection.
		// If it's closed, close the connection, make the pipe nil, and return.
		for {
			msg, open := <-user.Pipe

			if !open {
				c.Close()
				user.Pipe = nil
				return
			}

			c.WriteMessage(1, msg)
		}
	}
}

func TeamHandler(w http.ResponseWriter, r *http.Request) {
	_, user, err := VerifyCredentials(r)
	if err != nil {
		http.Redirect(w, r, "/join", 303)
		return
	}

	if user.Team != nil {
		http.Redirect(w, r, "/home", 303)
		return
	}

	teamTemplate.Execute(w, nil)
}

type playData struct {
	GameURL string

	FlagPlaceholder string
}

func PlayHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user already has the proper room code and token.
	var room *ctf.Room
	var user *ctf.User

	room, user, err := VerifyCredentials(r)
	if err != nil {
		http.Redirect(w, r, "/join", 303)
		return
	}

	if user.Team == nil {
		http.Redirect(w, r, "/team", 303)
		return
	}

	data := playData{
		GameURL:         fmt.Sprintf("ws://%v/game?token=%v&room=%v", server.Addr, user.Token, room.Code),
		FlagPlaceholder: room.Config.FlagPlaceholder,
	}
	playTemplate.Execute(w, data)
}

func JoinTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		room, user, err := VerifyCredentials(r)
		if err != nil {
			http.Redirect(w, r, "/join", 303)
			return
		}

		if user.Team != nil {
			http.Redirect(w, r, "/home", 303)
			return
		}

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
	}
}

func CreateTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		room, user, err := VerifyCredentials(r)
		if err != nil {
			http.Redirect(w, r, "/join", 303)
			return
		}

		if user.Team != nil {
			http.Redirect(w, r, "/home", 303)
			return
		}

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
				log.Printf("Error creating team %v", err)
				http.Redirect(w, r, "/team", 303)
				return
			}

			err = user.JoinTeam(team)
			if err != nil {
				http.Redirect(w, r, "/team", 303)
				return
			}

			log.Printf("Team %v with code %v Created by %v", team.Name, team.JoinCode, user.ID)
			http.Redirect(w, r, "/play", 303)
		}

	}
}

func sanitize(input string) string {
	// TODO: Properly sanitize input.
	// TODO: Add a profanity filter config option
	if input == "" {
		return ""
	}
	return input
}
