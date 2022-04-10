package main

import (
	"fmt"

	"github.com/civiledcode/goctf/ctf"
	"github.com/civiledcode/goctf/ctf/config"
	"github.com/civiledcode/goctf/server"
)

func main() {
	conf := config.Config{
		ConfigID: "example",
		MaxTeams: 10,
		MaxTeamSize: 4,
		GameLength: 120,
		Questions: make([]config.Question, 6),
	}

	for i := 0; i < 5; i++ {
		conf.Questions[i] = config.Question {
			Question: fmt.Sprintf("This is question #%v", i + 1),
			Answer: "answer",
			Hint: "hint",
			Points: 10,
			HintCost: 5,
			WrongCost: 1,
			CaseSensitive: false,
		}
	}

	conf.Questions[5] = config.Question {
		Question: "What's 9+10?",
		Answer: "21",
		Hint: "",
		Points: 15,
		HintCost: 5,
		WrongCost: 1,
		CaseSensitive: false,
	}

	// I can do whatever I want in here

	fmt.Println()
	//fmt.Println(string(content), "\n\x1b[34mConfig Loaded!\x1b[0m\n")

	room := ctf.NewRoom(conf)

	fmt.Printf("New room created with code %v\n", room.Code)
	server.Start("", 8000)

	for {

	}
}
