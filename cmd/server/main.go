package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"os/signal"
)

func main() {
	rabbit := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(rabbit)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	fmt.Println("The connection to RabbitMQ was successful.")
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	err = pubsub.SubscribeGOB(connection, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.QueueTypeDurable, handleLog())
	if err != nil {
		log.Fatalf("Failed to subscribe to game logs: %v", err)
	}

	gamelogic.PrintServerHelp()

	for {
		input := gamelogic.GetInput()

		if len(input) == 0 {
			continue
		}

		switch {
		case input[0] == "pause":
			fmt.Println("Pausing the game.")
			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Fatalf("Failed to publish the message. Error:", err)
			}
		case input[0] == "resume":
			fmt.Println("Resuming the game.")
			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Fatalf("Failed to publish the message. Error:", err)
			}
		case input[0] == "quit":
			fmt.Println("Quitting the game.")
			break
		default:
			fmt.Println("Unknown command:", input[0])
		}
	}

	// Wait for a signal to exit the application.
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	<-exit
}
