package main

import (
	"fmt"
	pubsub "github.com/bootdotdev/learn-pub-sub-starter/internal"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerMove(state *gamelogic.GameState, channel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := state.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+state.GetUsername(), gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: state.GetPlayerSnap(),
			})
			if err != nil {
				fmt.Println("Failed to publish the message. Error:", err)
				return pubsub.NackRequeue
			}
			return pubsub.NackRequeue
		default:
			fmt.Println("Unexpected outcome:", outcome)
			return pubsub.NackDiscard
		}

	}
}

func handlerPause(state *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		state.HandlePause(ps)
		return pubsub.Ack
	}
}

func handleWar(state *gamelogic.GameState) func(war gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(war gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, _, _ := state.HandleWar(war)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon, gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		default:
			fmt.Println("Unexpected outcome:", outcome)
			return pubsub.NackDiscard
		}
	}
}
