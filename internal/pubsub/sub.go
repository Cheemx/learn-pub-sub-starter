package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

const (
	Ack Acktype = iota
	NackDiscard
	NackRequeue
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	amqpChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	durable, autoDelete, exclusive := false, false, false
	if queueType == Durable {
		durable = true
	}
	if queueType == Transient {
		autoDelete = true
		exclusive = true
	}
	q, err := amqpChan.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = amqpChan.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return amqpChan, q, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	amqpChan, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deiverChan, err := amqpChan.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	unmarshaller := func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}

	go func() {
		defer amqpChan.Close()
		for delivery := range deiverChan {
			target, err := unmarshaller(delivery.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message body: %v\n", err)
				continue
			}
			switch handler(target) {
			case Ack:
				delivery.Ack(false)
				fmt.Println("Ack")
			case NackDiscard:
				delivery.Nack(false, false)
				fmt.Println("NackDiscard")
			case NackRequeue:
				delivery.Nack(false, true)
				fmt.Println("NackRequeue")
			}
		}
	}()
	return nil
}
