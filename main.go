package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	"github.com/civiledcode/goctf/ctf"
	"github.com/civiledcode/goctf/ctf/config"
	"github.com/civiledcode/goctf/server"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// Load the config
	configData, err := os.ReadFile("./configs/test_1.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v\n", err)
	}
	var conf config.Config

	err = json.Unmarshal(configData, &conf)

	if err != nil {
		log.Fatalf("Error parsing config file: %v\n", err)
	}

	//
	// Create the room
	room := ctf.NewRoom(conf)

	user := room.CreateUser("Test")
	team, err := room.CreateTeam("Test")
	if err != nil {
		log.Fatalf("Error creating new team: %v\n", err)
	}
	user.JoinTeam(team)

	err = team.Complete(user.ID, "question_one")
	if err != nil {
		log.Fatalf("Error answering question: %v\n", err)
	}

	fmt.Printf("\n\nRoom Code: %v\nUser Token: %v\nTeam Code: %v\n\n", room.Code, user.Token, team.JoinCode)

	room.Start()
	server.Start("localhost", 8000)

	for {

	}
}
