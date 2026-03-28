package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	connectionString := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connection successful.")

	//func (c *Connection) Channel() (*Channel, error)
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error creating ch: %v", err)
	}

	err = pubsub.SubscribeGob[routing.GameLog](
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
		func(log routing.GameLog) pubsub.AckType {
			defer gamelogic.PrintServerHelp()

			err := gamelogic.WriteLog(log)
			if err != nil {
				return pubsub.NackDiscard
			}
			return pubsub.Ack
		},
	)
	if err != nil {
		log.Fatalf("Error subsribiingGob: %v", err)
	}

	// wait for ctrl+c
	/*
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		<-signalChan
	*/

	gamelogic.PrintServerHelp()
	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}
		switch userInput[0] {
		case "pause":
			fmt.Println("Sending 'pause' message...")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Fatalf("Error sending 'pause' message: %v", err)
			}
		case "resume":
			fmt.Println("Sending 'resume' message...")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Fatalf("Error sending 'resume' message: %v", err)
			}
		case "quit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Command not understood.")
		}
	}

	//fmt.Println("Shutting down and closing connection...")

}
