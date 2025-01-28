package main

import (
	"codeground-manager/consumer"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	rabbitMQURL := "amqp://guest:guest@localhost:5672/"
	go consumer.StartConsumer(ctx, rabbitMQURL)

	log.Println("Consumer Started")
	<-ctx.Done()
	log.Println("Graceful shutdown")
	time.Sleep(1 * time.Second) 
}