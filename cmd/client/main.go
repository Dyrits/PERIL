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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Failed to get the username: %v", err)
	}

	_, queue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.QueueTypeTransient)
	if err != nil {
		log.Fatalf("Failed to declare and bind the queue: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	state := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.QueueTypeTransient, handlerPause(state))
	if err != nil {
		log.Fatalf("Failed to subscribe to pause: %v", err)
	}
	fmt.Printf("Subscribed to pause for queue %v!\n", queue.Name)

	channel, queue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, routing.ArmyMovesPrefix+".*", pubsub.QueueTypeTransient)
	if err != nil {
		log.Fatalf("Failed to declare and bind the queue: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	err = pubsub.SubscribeJSON(connection, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, routing.ArmyMovesPrefix+".*", pubsub.QueueTypeTransient, handlerMove(state))
	if err != nil {
		log.Fatalf("Failed to subscribe to moves: %v", err)
	}
	fmt.Printf("Subscribed to moves for queue %v!\n", queue.Name)

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
				err = pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, move)
				if err != nil {
					fmt.Printf("Failed to publish the move. %v:\n", err)
				}
			}
		case input[0] == "status":
			state.CommandStatus()
		case input[0] == "help":
			gamelogic.PrintClientHelp()
		case input[0] == "spam":
			fmt.Println("Spamming not allowed yet!")
		default:
			fmt.Println("Unknown command:", input[0])
		}
	}

	// Wait for a signal to exit the application.
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	<-exit
}

func handlerPause(game *gamelogic.GameState) func(routing.PlayingState) {
	defer fmt.Print("> ")
	return func(state routing.PlayingState) {
		game.HandlePause(state)
	}
}

func handlerMove(game *gamelogic.GameState) func(move gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		game.HandleMove(move)
		fmt.Print("> ")
	}
}
