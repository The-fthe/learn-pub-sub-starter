package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("get username failed", err)
		return
	}

	exchange := routing.ExchangePerilDirect
	queueName := routing.PauseKey + "." + username
	routingKey := routing.PauseKey
	pubsub.DeclareAndBind(amqp, exchange, queueName, routingKey, pubsub.SimpleQueueTransient)

	//wait for ctrl +c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
