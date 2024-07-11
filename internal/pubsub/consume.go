package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQuereDurable SimpleQueueType = iota
	SimpleQueueTransient
)

type Acktype int

const (
	Ack Acktype = iota
	NackDiscard
	NackRequeue
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
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
			acktype := handler(target)
			switch acktype {
			case Ack:
				err = msg.Ack(false)
				log.Println("Ack is called")
				if err != nil {
					fmt.Println("error Occured", err)
					return
				}

			case NackRequeue:
				err = msg.Nack(false, true)
				log.Println("NackRequeue is called")
				if err != nil {
					fmt.Println("error Occured", err)
					return
				}

			case NackDiscard:
				err = msg.Nack(false, false)
				log.Println("NackDiscard is called")
				if err != nil {
					fmt.Println("error Occured", err)
					return
				}

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

	queue, err := ch.QueueDeclare(
		queueName,
		isDurable,
		isTransit,
		isTransit,
		false,
		amqp.Table{"x-dead-letter-exchange": "peril_dlx"},
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, queue, nil
}
