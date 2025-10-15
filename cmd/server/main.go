package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connStr := "amqp://guest:guest@localhost:5672/"
	amqpConn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("Error connecting to Rabbit MQ: %v", err)
	}
	defer amqpConn.Close()
	fmt.Println("RabbitMQ connection successful...")

	mqChan, err := amqpConn.Channel()
	if err != nil {
		log.Fatalf("Error creating mqChan: %v", err)
	}

	exchange := routing.ExchangePerilTopic
	queueName := routing.GameLogSlug
	key := routing.GameLogSlug + ".cheemx"
	queueType := pubsub.Durable

	_, _, err = pubsub.DeclareAndBind(amqpConn, exchange, queueName, key, queueType)
	if err != nil {
		log.Fatalf("Error declaring and binding queue: %v", err)
	}

	gamelogic.PrintServerHelp()

	for {
		cmds := gamelogic.GetInput()
		if len(cmds) == 0 {
			continue
		}

		switch cmds[0] {
		case "pause":
			log.Println("Sending a Pause message...")
			err = pubsub.PublishJSON(mqChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				log.Printf("Error publishing PlayingState to Rabbit MQ: %v", err)
			}
		case "resume":
			log.Println("Sending a resume message...")
			err = pubsub.PublishJSON(mqChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Printf("Error publishing PlayingState to Rabbit MQ: %v", err)
			}
		case "quit":
			log.Println("Exiting...")
			return
		default:
			log.Println("Couldn't understand the command")
		}
	}
}
