package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQuereDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	chDelivery, err := ch.Consume(
		queue.Name, //queue
		"",         //consumer
		false,      //auto-ack
		false,      //exclusive
		false,      //no-local
		false,      //no-wait
		nil,        // args
	)
	if err != nil {
		return err
	}

	unmarshaller := func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}
	go func() {
		defer ch.Close()
		for msg := range chDelivery {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Println("error Occured", err)
				return
			}
			handler(target)
			err = msg.Ack(false)
			if err != nil {
				fmt.Println("error Occured", err)
				return
			}
		}

	}()
	return nil

}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	isDurable := simpleQueueType == SimpleQuereDurable
	isTransit := simpleQueueType == SimpleQueueTransient

	queue, err := ch.QueueDeclare(queueName, isDurable, isTransit, isTransit, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, queue, nil
}
