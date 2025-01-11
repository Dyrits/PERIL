package pubsub

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

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
