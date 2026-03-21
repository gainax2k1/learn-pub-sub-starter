package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	valJ, err := json.Marshal(val)
	if err != nil {
		return err
	}

	// publishing config
	ctx := context.Background()
	mandatory := false
	immediate := false

	err = ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, amqp.Publishing{
		ContentType: "application/json",
		Body:        valJ,
	})

	return err
}

type SimpleQueueType int

const (
	Transient SimpleQueueType = iota
	Durable
)

var queueType = map[SimpleQueueType]string{
	Transient: "transient",
	Durable:   "durable",
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // SimpleQueueType is an "enum" type to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	//func (ch *Channel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args Table) (Queue, error)
	var durable, autoDelete, exclusive, noWait bool
	if queueType == Durable {
		durable = true
	}
	if queueType == Transient {
		autoDelete = true
		exclusive = true
	}
	q, err := ch.QueueDeclare(queueName, durable, autoDelete, exclusive, noWait, nil)
	if err != nil {
		return nil, amqp.Queue{}, err

	}
	err = ch.QueueBind(queueName, key, exchange, noWait, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, q, nil

}
