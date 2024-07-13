package main

import (
	"fmt"
	"log"
	"os"

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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("get username failed", err)
		return
	}

	//subscribe to pause routing key
	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatal("Binding failed")
	}
	fmt.Printf("Queue %v declared and bound! \n", queue.Name)

	//subscribe to army_move routing key
	moveCh, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatal("Binding failed")
	}
	defer moveCh.Close()
	fmt.Printf("Queue %v declared and bound! \n", queue.Name)

	gameState := gamelogic.NewGameState(username)

	//comsume pause routing key
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
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
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		HandleMove(moveCh, gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	//comsume war
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"war",
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQuereDurable,
		HandleWar(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to war: %v", err)
	}

	//wait for ctrl +c
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan

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
				moveCh,
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
			println("Spamming not allowed yet")
		case "quit":
			gamelogic.PrintQuit()
			os.Exit(1)
			return
		default:
			log.Println("Unknown input", words[0])
		}

	}
}
