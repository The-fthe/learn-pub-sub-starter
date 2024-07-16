package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
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
	return subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		func(data []byte) (T, error) {
			var target T
			err := json.Unmarshal(data, &target)
			return target, err
		},
	)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	unmarshaller := func(data []byte) (T, error) {
		var target T
		buffer := bytes.NewBuffer(data)
		decoder := gob.NewDecoder(buffer)
		err := decoder.Decode(&target)
		if err != nil {
			return target, err

		}
		return target, nil
	}
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		unmarshaller,
	)
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
	unmarshaller func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	err = ch.Qos(10, 0, false)
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

	go func() {
		defer ch.Close()
		for msg := range chDelivery {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Println("unmarshaller error Occured", err)
				return
			}
			acktype := handler(target)
			switch acktype {
			case Ack:
				err = msg.Ack(false)
				if err != nil {
					fmt.Println("error Occured", err)
					return
				}

			case NackRequeue:
				err = msg.Nack(false, true)
				if err != nil {
					fmt.Println("nack requeue  error Occured", err)
					return
				}

			case NackDiscard:
				err = msg.Nack(false, false)
				if err != nil {
					fmt.Println("nack discard error Occured", err)
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

	queue, err := ch.QueueDeclare(
		queueName,                             // name
		simpleQueueType == SimpleQueueDurable, // durable
		simpleQueueType != SimpleQueueDurable, // delete when unused
		simpleQueueType != SimpleQueueDurable, // exclusive
		false,                                 // no-wait
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)

	err = ch.QueueBind(
		queueName, //queue name
		key,       //routing key
		exchange,  //exchange
		false,     //no wiat
		nil,       //args
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return ch, queue, nil
}
