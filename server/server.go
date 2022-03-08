package server

import (
	"fmt"
	"html/template"
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
			fmt.Printf("Server Connection Error: %v\n", err)
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
		fmt.Printf("Server Close Error: %v\n", err)
	}

	Started = false
}

func ScoresHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World!")
	for _, cookie := range r.Cookies() {
		fmt.Println(cookie)
	}
}

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	cook := &http.Cookie{Name: "sample", Value: "sample", HttpOnly: false}
	http.SetCookie(w, cook)
}

type joinData struct {
	HideInvalidLogin bool
	ErrorMessage string
}

func JoinHandler(w http.ResponseWriter, r *http.Request) {
	var data joinData
	if r.Method == "POST" {
		r.ParseForm()
		fmt.Println("Attempted join")
		data = joinData {
			HideInvalidLogin: false,
		}

		// Check if the user already has the proper room code and token.
		cookie, err := r.Cookie("room_code")
		if err != nil {
			cookie, err = r.Cookie("token")
			if err != nil {
				http.Redirect(w, r, "/play", 303)
			}
		}

		aliase, ok := r.Form["aliase"]
		sanitized := sanitize(aliase[0])

		if !ok || sanitized == "" {
			data.ErrorMessage = "Invalid aliase."
		}

		roomcode, ok := r.Form["room_code"]

		if !ok {
			data.ErrorMessage = "Empty room code."
		}

		if data.ErrorMessage == "" {
			room := ctf.Rooms[roomcode[0]]
			if room == nil {
				data.ErrorMessage = "Invalid room code."
				joinTemplate.Execute(w, data)
				return
			}

			user := room.CreateUser(sanitized)

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
			fmt.Println("Logged in!")
		} else {
			fmt.Printf("Failed attempt with room %v and user %v and error %v\n%v", roomcode, sanitized, data.ErrorMessage, ctf.Rooms)
		}
		
	} else if r.Method == "GET" {
		data = joinData {
			HideInvalidLogin: true,
			ErrorMessage: "nil",
		}	
	} else {
		w.WriteHeader(405)
	}

	joinTemplate.Execute(w, data)
}

func TeamHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Do whatever") 
}

func PlayHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Address)	
}

func JoinTeamHandler(w http.ResponseWriter, r *http.Request) {

}

func CreateTeamHandler(w http.ResponseWriter, r *http.Request)  {

}


func sanitize(input string) string {
	if input == "" {
		return ""
	}
	return input
}
