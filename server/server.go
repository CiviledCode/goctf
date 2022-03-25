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
}

func Stop() {
	err := server.Close()
	
	if err != nil {
		log.Printf("Server Close Error: %v\n", err)
	}

	Started = false
}

func ScoresHandler(w http.ResponseWriter, r *http.Request) {
}

func SubmitHandler(w http.ResponseWriter, r *http.Request) {

}

type joinData struct {
	HideInvalidLogin bool
	ErrorMessage string
}


// Handles all requests on the /join path.
func JoinHandler(w http.ResponseWriter, r *http.Request) {
	var data joinData

	// Check if the user already has the proper room code and token.
	cookie, err := r.Cookie("room_code")
	if err == nil {
		if room := ctf.Rooms[cookie.Value]; cookie != nil && room != nil {
			cookie, err = r.Cookie("token")
			if err == nil {
				if user := room.UserByPrivate(cookie.Value); cookie != nil && user != nil{
					// Redirect to the main play handler as the user is already authenticated properly in a room.
					http.Redirect(w, r, "/play", 303)
				}
			}
		}
	}

    // The client is attempting to join some sort of room.
	if r.Method == "POST" {
		r.ParseForm()

		data = joinData {
			HideInvalidLogin: false,
		}

        // Retrieve the aliase from the form, sanitize it, and check to see if it's valid.
		aliase, ok := r.Form["aliase"]
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
			cookie = &http.Cookie {
				Name: "token",
				Value: user.Token,
			}
			http.SetCookie(w, cookie)

			cookie = &http.Cookie {
				Name: "room_code",
				Value: roomcode[0],
			}
			http.SetCookie(w, cookie)
			
            // Redirect to the play path because we are successfully authenticated.
			http.Redirect(w, r, "/play", 303)
		} else {
            // TODO; Create the dialogue within the html.
			log.Printf("Failed attempt to join room '%v' with aliase '%v'. Error: '%v'\n", roomcode[0], sanitized, data.ErrorMessage)
		}
		
    // The client is attempting to retrieve the webpage to join.
	} else if r.Method == "GET" {
		data = joinData {
			HideInvalidLogin: true,
			ErrorMessage: "",
		}	
    // An unknown method has occured.
	} else {
		w.WriteHeader(405)
	}

	joinTemplate.Execute(w, data)
}

func TeamHandler(w http.ResponseWriter, r *http.Request) {
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
				}
			} else {
                http.Redirect(w, r, "/join", 303)
            }
		} else {
            http.Redirect(w, r, "/join", 303)
        }
	} else {
        http.Redirect(w, r, "/join", 303)
    }

    if user.Team == nil {
        http.Redirect(w, r, "/team", 303)
    }

    // TODO: Serve webpage capable of fetching questions from /questions
}

func JoinTeamHandler(w http.ResponseWriter, r *http.Request) {
}

func CreateTeamHandler(w http.ResponseWriter, r *http.Request)  {
}


func sanitize(input string) string {
	// TODO: Properly sanitize input.
	// TODO: Add a profanity filter config option
	if input == "" {
		return ""
	}
	return input
}
