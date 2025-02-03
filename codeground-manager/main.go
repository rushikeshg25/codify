package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	amqp "github.com/rabbitmq/amqp091-go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type QueueMessage struct {
	Codeground Codeground `json:"codeground"`
	Msg        string     `json:"msg"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	namespace := "default"
	k8s := NewK8s(clientset, namespace)
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"codeground-queue", // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
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
			codeground := queueMsg.Codeground
			codeground.K8s = k8s

			switch queueMsg.Msg {
			case "create":
				log.Printf("Processing 'create' action for codeground: %+v\n", queueMsg.Codeground)
				codeground.CreateCodeground()
			case "delete":
				log.Printf("Processing 'delete' action for codeground: %+v\n", queueMsg.Codeground)
				codeground.DeleteCodeground()
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
