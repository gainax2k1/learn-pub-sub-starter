package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	fmt.Println("Starting Peril client...")

	connectionString := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connection successful.")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error getting username: %v", err)
	}
	pubsub.DeclareAndBind(conn, "peril_direct", ("pause." + username), routing.PauseKey, pubsub.Transient)

	gs := gamelogic.NewGameState(username)

	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}
		switch userInput[0] {
		case "spawn":
			err := gs.CommandSpawn(userInput)
			if err != nil {
				fmt.Printf("Unable to spawn: %v", err)
			}

		case "move":

			_, err := gs.CommandMove(userInput)
			if err != nil {
				fmt.Printf("unable to move: %v", err)
				break
			}
			//fmt.Printf("Move: %v", mv)
			fmt.Println("Move successful.")
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			//fmt.Println("Exiting...")
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Command not understood.")
		}
	}

	// wait for ctrl+c
	/*
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		<-signalChan
	*/
}
