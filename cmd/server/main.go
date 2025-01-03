package main

import (
	"fmt"
	pubsub "github.com/bootdotdev/learn-pub-sub-starter/internal"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
)

func main() {
	rabbit := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbit)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ. Error:", err)
		return
	}
	channel, err := connection.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel. Error:", err)
		return
	}
	pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	defer connection.Close()
	fmt.Println("The connection to RabbitMQ was successful.")

	// Wait for a signal to exit the application.
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	<-exit
}
