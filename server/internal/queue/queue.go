package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Codeground struct {
	ID             string `json:"id"`
	UserID         int    `json:"user_id"`
	Name           string `json:"name"`
	CodegroundType string `json:"codeground_type"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type QueueMessage struct {
	Codeground Codeground `json:"codeground"`
	Msg        string     `json:"msg"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type Queue struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
	Q    amqp.Queue
}

func InitQueue(queueName string) (*amqp.Connection, *amqp.Channel, amqp.Queue) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	return conn, ch, q
}

func (q *Queue) PublishToQueue(codeground Codeground, msg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message := QueueMessage{
		codeground,
		msg,
	}

	body, err := json.Marshal(message)
	failOnError(err, "Failed to serialize message")

	err = q.Ch.PublishWithContext(ctx,
		"",       // exchange
		q.Q.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", string(body))
}
