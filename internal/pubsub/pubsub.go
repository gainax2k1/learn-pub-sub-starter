package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

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

type AckType int

const (
	Ack AckType = iota
	NackDiscard
	NackRequeue
)

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
	//includes "x-dead-letter-exchange" key
	q, err := ch.QueueDeclare(queueName, durable, autoDelete, exclusive, noWait, amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	})
	if err != nil {
		return nil, amqp.Queue{}, err

	}
	err = ch.QueueBind(queueName, key, exchange, noWait, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, q, nil

}

// listens for messages and recieves them
func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	err = ch.Qos(10, 0, false)
	if err != nil {
		return err
	}
	messages, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	//go routine to avoid blocking forever on the for...range loop
	go func() {
		for msg := range messages {
			var body T
			err := json.Unmarshal(msg.Body, &body)
			if err != nil {
				fmt.Printf("Error unmarshalling JOSN body: %v", err)
				continue
			}

			ackType := handler(body)

			//handle ack for queue
			//delivery.Ack(false)
			switch ackType {
			case Ack:
				msg.Ack(false)

			case NackDiscard:
				msg.Nack(false, false)

			case NackRequeue:
				msg.Nack(false, true)
			default:
				fmt.Println("handler default case reached in client subscribejson")
			}
		}
	}()
	return nil
}

/*
main calls SubscribeJSON
    -> passes handlerPause(gs) as the handler
        -> handlerPause(gs) returns a func(routing.PlayingState)
            -> SubscribeJSON stores that function as "handler"
                -> calls it each time a message arrives
*/

// listens for messages and recieves them
func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	// call channel.qos here?
	err = ch.Qos(10, 0, false)
	if err != nil {
		return err
	}
	messages, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	//go routine to avoid blocking forever on the for...range loop
	go func() {
		for msg := range messages {
			buf := bytes.NewBuffer(msg.Body)
			decoder := gob.NewDecoder(buf)

			var body T
			err := decoder.Decode(&body)
			if err != nil {
				fmt.Printf("Error decoding  gob body: %v", err)
				continue
			}

			ackType := handler(body)

			//handle ack for queue
			//delivery.Ack(false)
			switch ackType {
			case Ack:
				msg.Ack(false)

			case NackDiscard:
				msg.Nack(false, false)

			case NackRequeue:
				msg.Nack(false, true)
			default:
				fmt.Println("handler default case reached in client subscribeGob")
			}
		}
	}()
	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(val)
	if err != nil {
		return err
	}

	// publishing config
	ctx := context.Background()
	mandatory := false
	immediate := false

	err = ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	})

	return err
}
