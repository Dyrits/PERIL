package main

import (
	"fmt"
	pubsub "github.com/bootdotdev/learn-pub-sub-starter/internal"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	_, queue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, "game_logs", "game_logs.*", pubsub.QueueTypeDurable)
	if err != nil {
		log.Fatalf("Failed to declare and bind the queue: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gamelogic.PrintServerHelp()

	for {
		input := gamelogic.GetInput()

		if len(input) == 0 {
			continue
		}

		switch {
		case input[0] == "pause":
			fmt.Println("Pausing the game.")
			err := pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Fatalf("Failed to publish the message. Error:", err)
			}
		case input[0] == "resume":
			fmt.Println("Resuming the game.")
			err := pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
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
