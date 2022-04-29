package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/civiledcode/goctf/ctf"
	"github.com/civiledcode/goctf/ctf/config"
	"github.com/civiledcode/goctf/server"
)

func main() {
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

	log.Printf("New room created with code %v\n", room.Code)
	server.Start("127.0.0.1", 8000)

	for {

	}
}
