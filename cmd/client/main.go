package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
