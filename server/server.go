package server

import (
	"fmt"
	"net/http"
)

var server *http.Server
var Started bool

func Start(ip string, port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/scores", ScoresHandler)
	mux.HandleFunc("/submit", SubmitHandler)

	server = &http.Server{Addr: fmt.Sprintf("%v:%v", ip, port), Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("Server Connection Error: %v\n", err)
		}
	}()

	Started = true
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


