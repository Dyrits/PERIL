package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handleLog() func(log routing.GameLog) pubsub.AckType {
	return func(log routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")

		err := gamelogic.WriteLog(log)
		if err != nil {
			fmt.Println("Failed to write the log. Error:", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
