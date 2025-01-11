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

	args := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}
	queue, err := channel.QueueDeclare(queueName, queueType == QueueTypeDurable, queueType == QueueTypeTransient, queueType == QueueTypeTransient, false, args)
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

func SubscribeJSON[T any](
	connection *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType QueueType,
	handler func(T) AckType,
) error {
	channel, queue, err := DeclareAndBind(connection, exchange, queueName, key, queueType)
	if err != nil {
		fmt.Println("Failed to declare and bind the queue. Error:", err)
		return err
	}
	delivery, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println("Failed to consume the queue. Error:", err)
		return err
	}
	go func() {
		for message := range delivery {
			var value T
			err := json.Unmarshal(message.Body, &value)
			if err != nil {
				fmt.Println("Failed to unmarshal the message. Error:", err)
				continue
			}

			ackType := handler(value)
			switch ackType {
			case Ack:
				err = message.Ack(false)
				if err == nil {
					fmt.Println("Message aknowledged (Ack)")
				}
			case NackRequeue:
				err = message.Nack(false, true)
				if err == nil {
					fmt.Println("Message requeued (NackRequeue)")
				}
			case NackDiscard:
				err = message.Nack(false, false)
				if err == nil {
					fmt.Println("Message discarded (NackDiscard)")
				}
			}
		}
	}()
	return nil
}

type QueueType int

const (
	QueueTypeDurable QueueType = iota
	QueueTypeTransient
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)
