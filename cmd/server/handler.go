package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func HandleLog() func(routing.GameLog) pubsub.Acktype {
	return func(log routing.GameLog) pubsub.Acktype {
		defer fmt.Println(">")

		err := gamelogic.WriteLog(log)
		if err != nil {
			fmt.Printf("error writing log:%v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
