package pubsub

import (
	"encoding/json"
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
