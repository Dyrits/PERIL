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
	"strconv"
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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Failed to get the username: %v", err)
	}

	state := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.QueueTypeTransient,
		handlePause(state),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to pause: %v", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.QueueTypeTransient,
		handleMove(state, channel),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to moves: %v", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.QueueTypeDurable,
		handleWar(state, channel),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to wars: %v", err)
	}

	for {
		input := gamelogic.GetInput()

		if len(input) == 0 {
			continue
		}

		switch {
		case input[0] == "spawn":
			err := state.CommandSpawn(input)
			if err != nil {
				fmt.Printf("Failed to spawn: %v\n", err)
			}
		case input[0] == "move":
			move, err := state.CommandMove(input)

			if err != nil {
				fmt.Printf("Failed to move: %v\n", err)
			} else {
				fmt.Println("Successfully moved. Status:", move)
				err = pubsub.PublishJSON(
					channel,
					routing.ExchangePerilTopic,
					routing.ArmyMovesPrefix+"."+username,
					move,
				)
				if err != nil {
					fmt.Printf("Failed to publish the move. %v:\n", err)
				}
			}
		case input[0] == "status":
			state.CommandStatus()
		case input[0] == "help":
			gamelogic.PrintClientHelp()
		case input[0] == "spam":
			// Convert the number of messages to spam.
			quantity, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Println("Spamming not allowed yet!")
			} else {
				for _ = range quantity {
					err := pubsub.PublishGOB(
						channel,
						routing.ExchangePerilTopic,
						routing.GameLogSlug+"."+username,
						routing.GameLog{
							Username: state.GetUsername(),
							Message:  gamelogic.GetMaliciousLog(),
						},
					)
					if err != nil {
						fmt.Println("Failed to publish the message. Error:", err)
					}
				}
			}

		default:
			fmt.Println("Unknown command:", input[0])
		}
	}

	// Wait for a signal to exit the application.
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	<-exit
}
