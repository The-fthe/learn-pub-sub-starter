package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandlePause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func HandleMove(ch *amqp.Channel, gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(am gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(am)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard

		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack

		case gamelogic.MoveOutcomeMakeWar:
			war := gamelogic.RecognitionOfWar{
				Attacker: am.Player,
				Defender: gs.Player,
			}

			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.Player.Username,
				war,
			)
			if err != nil {
				fmt.Println("Failed to publish message", err)
				return pubsub.NackRequeue
			}
			return pubsub.NackRequeue
		}
		return pubsub.NackDiscard
	}
}

func HandleWar(gs *gamelogic.GameState) func(d amqp.Delivery) pubsub.Acktype {
	return func(d amqp.Delivery) pubsub.Acktype {
		defer fmt.Print("> ")

		// Call HandleWar with the body of the message
		war := gamelogic.RecognitionOfWar{}
		log.Println("body message: ", string(d.Body))
		err := json.Unmarshal(d.Body, &war)
		if err != nil {
			fmt.Println("Unmarshal war failed")
			return pubsub.NackDiscard
		}
		outcome, _, _ := gs.HandleWar(war)

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack

		default:
			fmt.Println("Unknown outcome:", outcome)
			return pubsub.NackDiscard
		}
	}
}
