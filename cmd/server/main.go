package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	amqp, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
		return
	}
	defer amqp.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	ch, err := amqp.Channel()
	if err != nil {
		log.Fatalf("channel err: %v", err)
		return
	}
	exchange := routing.ExchangePerilDirect
	key := routing.PauseKey
	state := routing.PlayingState{
		IsPaused: true,
	}
	err = pubsub.PublishJSON(ch, exchange, key, state)
	if err != nil {
		log.Fatalf("PublishErr: %v", err)
		return
	}
	fmt.Println("Exchange channel created")

	//wait for ctrl +c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

}
