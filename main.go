package main

import (
	"encoding/json"
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

	// Create the room
	room := ctf.NewRoom(conf)

	room.Start()
	log.Printf("New room created with code %v\n", room.Code)
	server.Start("localhost", 8000)

	for {

	}
}
