package pubsub

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](channel *amqp.Channel, exchange, key string, value T) error {
	// Convert the value to JSON:
	json, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Publish the JSON to the exchange.
	return channel.Publish(exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        json,
	})
}

func DeclareAndBind(
	connection *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := connection.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel. Error:", err)
		return nil, amqp.Queue{}, err
	}
	queue, err := channel.QueueDeclare(queueName, queueType == QueueTypeDurable, queueType == QueueTypeTransient, queueType == QueueTypeTransient, false, nil)
	if err != nil {
		fmt.Println("Failed to declare a queue. Error:", err)
		return nil, amqp.Queue{}, err
	}
	err = channel.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		fmt.Println("Failed to bind the queue. Error:", err)
		return nil, amqp.Queue{}, err
	}
	return channel, queue, nil
}

type QueueType int

const (
	QueueTypeDurable QueueType = iota
	QueueTypeTransient
)
