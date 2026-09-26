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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	namespace := envOr("K8S_NAMESPACE", "default")
	k8s := NewK8s(clientset, namespace)
	conn, err := amqp.Dial(envOr("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"codeground-queue", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// One unacked message at a time: provisioning is slow and serial, and
	// without this a single manager claims the whole backlog and then loses it
	// if it restarts.
	failOnError(ch.Qos(1, 0, false), "Failed to set channel QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack: acknowledge only once the work is actually done
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
				// Undecodable: requeueing would just spin forever.
				log.Printf("Error decoding message: %s", err)
				_ = d.Nack(false, false)
				continue
			}
			codeground := queueMsg.Codeground
			codeground.K8s = k8s

			var actionErr error
			switch queueMsg.Msg {
			case "create":
				log.Printf("Processing 'create' action for codeground: %+v\n", queueMsg.Codeground)
				actionErr = codeground.CreateCodeground()
			case "delete":
				log.Printf("Processing 'delete' action for codeground: %+v\n", queueMsg.Codeground)
				actionErr = codeground.DeleteCodeground()
			case "stop":
				log.Printf("Processing 'stop' action for codeground: %+v\n", queueMsg.Codeground)
				actionErr = codeground.StopCodeground()
			default:
				log.Printf("Unknown action action: %s\n", queueMsg.Msg)
			}

			if actionErr != nil {
				log.Printf("Action %q failed for codeground %s: %v", queueMsg.Msg, codeground.Id, actionErr)
				// Retry once, then drop it rather than looping on a message
				// that is never going to succeed. The reconciler is the
				// backstop for anything left half-built.
				requeue := !d.Redelivered
				_ = d.Nack(false, requeue)
				continue
			}

			_ = d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
