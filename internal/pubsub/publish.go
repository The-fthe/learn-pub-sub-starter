package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	publish := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, publish)
	if err != nil {
		return err
	}

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(val)
	if err != nil {
		return err
	}
	publish := amqp.Publishing{
		ContentType: "application/gob",
		Body:        buffer.Bytes(),
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, publish)
	if err != nil {
		return err
	}
	return nil

}
