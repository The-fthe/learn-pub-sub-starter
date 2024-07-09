package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func main() {
	amqp, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
		return
	}
	defer amqp.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	// ch, err := amqp.Channel()
	// if err != nil {
	// 	log.Fatalf("channel err: %v", err)
	// 	return
	// }

	ch, queue, err := pubsub.DeclareAndBind(
		amqp,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQuereDurable)
	if err != nil {
		log.Fatalf("could not subscribe to pause  err: %v", err)
		return
	}
	fmt.Printf("Queue %v declared and bound\n", queue.Name)

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) <= 0 {
			continue
		}
		switch words[0] {

		case "pause":
			// err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Fatalf("pause: %v", err)
				return
			}
			log.Println("Pause is trigger")
		case "resume":
			//err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Fatalf("resume: %v", err)
				return
			}
			log.Println("Resume is trigger")
		case "quit":
			fmt.Println("goodbye", words[0])
			return
		default:
			fmt.Println("Unknow message:", words[0])
		}
	}
}
