package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func rabbitMQURL() string {
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		return url
	}
	return "amqp://guest:guest@localhost:5672/"
}

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
	conn, err := amqp.Dial(rabbitMQURL())
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	// Durable, so a broker restart does not silently discard every pending
	// create and delete. Must match the manager's declaration exactly or the
	// channel is closed with PRECONDITION_FAILED.
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	return conn, ch, q
}

/*
PublishToQueue hands a codeground action to the manager.

This used to panic the whole API server when the broker was unreachable, which
turned a transient RabbitMQ blip into an outage for every user — and, on the
delete path, left Kubernetes resources orphaned with no database row left to
find them by. Failures are returned so the caller can answer 5xx and leave
state consistent.
*/
func (q *Queue) PublishToQueue(codeground Codeground, msg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message := QueueMessage{
		codeground,
		msg,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("serialize codeground message: %w", err)
	}

	if err := q.Ch.PublishWithContext(ctx,
		"",       // exchange
		q.Q.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		}); err != nil {
		return fmt.Errorf("publish codeground message: %w", err)
	}

	log.Printf(" [x] Sent %s\n", string(body))
	return nil
}
