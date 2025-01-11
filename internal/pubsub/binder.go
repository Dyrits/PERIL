package pubsub

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

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
