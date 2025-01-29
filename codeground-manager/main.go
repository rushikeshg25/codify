package main

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueMessage struct {
	Codeground Codeground `json:"codeground"`
	Msg        string      `json:"msg"` 
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"codeground-queue", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var queueMsg QueueMessage
			err := json.Unmarshal(d.Body, &queueMsg)
			if err != nil {
				log.Printf("Error decoding message: %s", err)
				continue
			}
			codeground:=queueMsg.Codeground
			switch queueMsg.Msg {
			case "create":
				log.Printf("Processing 'create' action for codeground: %+v\n", queueMsg.Codeground)
				codeground.CreateCodeground()
			case "delete":
				log.Printf("Processing 'delete' action for codeground: %+v\n", queueMsg.Codeground)
				codeground.DeleteCodeground()
			case "start":
				log.Printf("Processing 'start' action for codeground: %+v\n", queueMsg.Codeground)
				codeground.StartCodeground()
			case "stop":
				log.Printf("Processing 'stop' action for codeground: %+v\n", queueMsg.Codeground)
				codeground.StopCodeground()
			default:
				log.Printf("Unknown action action: %s\n", queueMsg.Msg)
			}		
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}