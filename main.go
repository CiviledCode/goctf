package main

import (
	"github.com/civiledcode/goctf/server"
)

func main() {
	/*
	ctf.Random = rand.New(rand.NewSource(time.Now().Unix()))
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

	content, err := json.Marshal(conf)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(content), "\n\x1b[34mConfig Loaded!\x1b[0m\n")	

	room := ctf.NewRoom(conf)

	fmt.Println("New room created\n")

	team := room.CreateTeam("Clones")
	
	fmt.Printf("New team '%v' created with UUID '%v'\n\n", team.Name, team.ID)

	me := room.CreateUser("Civiled")
	err = me.JoinTeam(team)

	fmt.Printf("Created your account with UUID '%v'\n\n", me.UUID)
	if err != nil {
		fmt.Println(err)
	}

	for i := 0; i < 25; i++ {
		user := room.CreateUser(fmt.Sprintf("Clone %v", i))

		fmt.Printf("Aliase: %v Token: %v UserID: %v\n", user.Aliase, user.Token, user.UUID)
		err = user.JoinTeam(team)
		
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	fmt.Println()
	for questionID, question := range room.Questions {
		fmt.Printf("Attempting question with ID '%v'\n", questionID)
		isRight, err := room.AnswerQuestion(me.UUID, questionID, "21")
		if err != nil {
			fmt.Println(err)
		}

		if isRight {
			fmt.Printf("Question '%v' got right with answer '21'!\n", question.Question)
		}
	}

	fmt.Printf("\n\n\x1b[32mFinished! \x1b[0mTeam '%v' with data %v", team.Name, team.Data())
	*/
	server.Start("", 8000)

	for {

	}	
}
