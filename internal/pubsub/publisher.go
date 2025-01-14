package pubsub

import (
	"bytes"
	"encoding/gob"
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

func PublishGOB[T any](channel *amqp.Channel, exchange, key string, value T) error {
	// Convert the value to GOB:
	encoded, err := encode(value)
	if err != nil {
		return err
	}

	// Publish the JSON to the exchange.
	return channel.Publish(exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        encoded,
	})
}

func encode[T any](value T) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
