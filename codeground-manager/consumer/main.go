package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueMessage struct {
	Codeground *Codeground `json:"codeground"`
	Msg        string      `json:"msg"` 
}

type Codeground struct {
	ID             string    `json:"id"`
	UserID         int       `json:"user_id"`
	Name           string    `json:"name"`
	CodegroundType string    `json:"codeground_type"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func StartConsumer(ctx context.Context, rabbitMQURL string) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare the queue
	queueName := "codeground_queue"
	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println("Consumer started. Waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopped due to context cancellation")
			return
		case msg := <-msgs:
			msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := processMessage(msgCtx, msg.Body); err != nil {
				log.Printf("Failed to process message: %v", err)
			}
		}
	}
}

func processMessage(ctx context.Context, body []byte) error {
	var queueMsg QueueMessage
	if err := json.Unmarshal(body, &queueMsg); err != nil {
		return err
	}

	// Handle the message based on the action
	switch queueMsg.Msg {
	case "create":
		log.Printf("Processing 'create' message for codeground: %+v\n", queueMsg.Codeground)
		// Add your logic here
	case "update":
		log.Printf("Processing 'update' message for codeground: %+v\n", queueMsg.Codeground)
		// Add your logic here
	case "delete":
		log.Printf("Processing 'delete' message for codeground: %+v\n", queueMsg.Codeground)
		// Add your logic here
	case "start":
		log.Printf("Processing 'start' message for codeground: %+v\n", queueMsg.Codeground)
		// Add your logic here
	case "stop":
		log.Printf("Processing 'stop' message for codeground: %+v\n", queueMsg.Codeground)
		// Add your logic here
	default:
		log.Printf("Unknown message action: %s\n", queueMsg.Msg)
	}

	return nil
}