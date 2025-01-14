package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handleMove(state *gamelogic.GameState, channel *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := state.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			// The war queue must have been declared through RabbitMQ Management UI before.
			err := pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+state.GetUsername(), gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: state.GetPlayerSnap(),
			})
			if err != nil {
				fmt.Println("Failed to publish the message. Error:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("Unexpected outcome:", outcome)
			return pubsub.NackDiscard
		}

	}
}

func handlePause(state *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		state.HandlePause(ps)
		return pubsub.Ack
	}
}

func handleWar(state *gamelogic.GameState, channel *amqp.Channel) func(war gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(war gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := state.HandleWar(war)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			err := pubsub.PublishGOB(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+state.GetUsername(), routing.GameLog{
				Username: state.GetUsername(),
				Message:  fmt.Sprintf("%v won a war against %v", winner, loser),
			})
			if err != nil {
				fmt.Println("Failed to publish the message. Error:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			err := pubsub.PublishGOB(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+state.GetUsername(), routing.GameLog{
				Username: state.GetUsername(),
				Message:  fmt.Sprintf("%v won a war against %v", winner, loser),
			})
			if err != nil {
				fmt.Println("Failed to publish the message. Error:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			err := pubsub.PublishGOB(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+state.GetUsername(), routing.GameLog{
				Username: state.GetUsername(),
				Message:  fmt.Sprintf("A war between %v and %v resulted in a draw", winner, loser),
			})
			if err != nil {
				fmt.Println("Failed to publish the message. Error:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("Unexpected outcome:", outcome)
			return pubsub.NackDiscard
		}
	}
}
