package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
		return
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("get username failed", err)
		return
	}

	gameState := gamelogic.NewGameState(username)

	//comsume pause routing key
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		HandlePause(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	//comsume army_move
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		HandleMove(gameState, publishCh),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	//comsume war
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		HandleWar(gameState, publishCh),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) <= 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				log.Println("command Spawn is failed")
				continue
			}

		case "move":
			am, err := gameState.CommandMove(words)
			if err != nil {
				log.Println("command move is failed")
				continue
			}
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				am,
			)
			if err != nil {
				log.Fatalf("pause: %v", err)
				return
			}
			log.Println("Pause is trigger")

		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(words) != 2 {
				log.Println("spam count is not set")
				return
			}
			spamCount, err := strconv.Atoi(words[1])
			if err != nil {
				log.Fatalf("Atoi err : %v", err)
				return
			}
			for i := 0; i < spamCount; i++ {
				msg := gamelogic.GetMaliciousLog()
				err = publishGameLog(publishCh, username, msg)
				if err != nil {
					log.Fatalf("Log publish failed: %v\n", err)
				}
				log.Printf("Published %v malicious log\n", i+1)
			}

		case "quit":
			gamelogic.PrintQuit()
			os.Exit(1)
			return
		default:
			log.Println("Unknown input", words[0])
		}

	}
}
func publishGameLog(publishCh *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		publishCh,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     msg,
		},
	)

}
