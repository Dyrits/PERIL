package main

import (
	"fmt"
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
	defer connection.Close()
	fmt.Println("The connection to RabbitMQ was successful.")

	// Wait for a signal to exit the application.
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel
}
