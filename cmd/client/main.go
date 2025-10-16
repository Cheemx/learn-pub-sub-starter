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
	fmt.Println("Starting Peril client...")
	connStr := "amqp://guest:guest@localhost:5672/"
	amqpConn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("Error connecting to Rabbit MQ: %v", err)
	}
	defer amqpConn.Close()
	fmt.Println("RabbitMQ connection successful...")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error getting username: %v", err)
	}

	exchange := routing.ExchangePerilTopic
	queueName := routing.PauseKey + "." + username
	pause := routing.PauseKey
	queueType := pubsub.Transient

	_, _, err = pubsub.DeclareAndBind(amqpConn, exchange, queueName, pause, queueType)
	if err != nil {
		log.Fatalf("Error declaring and binding queue: %v", err)
	}

	_, _, err = pubsub.DeclareAndBind(amqpConn, exchange, "army_moves."+username, "army_moves.*", queueType)
	if err != nil {
		log.Fatalf("Error declaring and binding queue: %v", err)
	}

	gameState := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON(amqpConn, exchange, queueName, pause, queueType, handlerPause(gameState))
	if err != nil {
		log.Fatalf("Error declaring and binding queue: %v", err)
	}

	err = pubsub.SubscribeJSON(amqpConn, exchange, "army_moves."+username, "army_moves.*", queueType, handlerMove(gameState))
	if err != nil {
		log.Fatalf("Error declaring and binding move queue: %v", err)
	}

	pubCh, err := amqpConn.Channel()
	if err != nil {
		log.Printf("Error using spawn command: %v", err)
	}
	defer pubCh.Close()

	for {
		cmds := gamelogic.GetInput()

		if len(cmds) == 0 {
			continue
		}

		switch cmds[0] {
		case "spawn":
			err := gameState.CommandSpawn(cmds)
			if err != nil {
				log.Printf("Error using spawn command: %v", err)
			}
		case "move":
			mv, err := gameState.CommandMove(cmds)
			if err != nil {
				log.Printf("Error using spawn command: %v", err)
			}
			err = pubsub.PublishJSON(pubCh, exchange, "army_moves."+username, mv)
			if err != nil {
				log.Printf("Error using spawn command: %v", err)
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			log.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Println("The entered command is not supported!")
		}
	}
}
