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
			_, err := gameState.CommandMove(words)
			if err != nil {
				log.Println("command move is failed")
				continue
			}

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

// func handlePause(gs *gamelogic.GameState) func(routing.PlayingState) {
// 	return func(ps routing.PlayingState) {
// 		defer fmt.Print("> ")
// 		gs.HandlePause(ps)
// 	}
// }
