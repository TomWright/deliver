package main

import (
	"context"
	"fmt"
	"github.com/tomwright/deliver"
	"github.com/tomwright/deliver/examples/messages"
	"log"
	"sync"
	"time"
)

func main() {
	ctx := context.Background()

	// initialise the subscriber
	subscriber := deliver.NewKafkaSubscriber([]string{"cmg-local-kafka:9092"})

	// start up a go routine to listen for and log errors returned from the consumer
	consumerErrors := make(chan error)
	go logConsumerErrors(consumerErrors)

	// create a WaitGroup so we know when all consumers have stopped
	wg := &sync.WaitGroup{}

	// define consumer options
	opts := deliver.SubscribeOptions{
		ConsumeFn: userCreatedHandler,
		Group:     "message-logger",
		Types:     []string{messages.TypeUserCreated},
		Errors:    consumerErrors,
	}
	// start x instances of the consumer, that will timeout and stop after 60 seconds.
	// the timeout probably wouldn't be applicable in a real-world application.
	consumerCtx, _ := context.WithTimeout(ctx, time.Second*60)
	consumerStartupErrors := deliver.SubscribeNonBlocking(consumerCtx, opts, subscriber, wg, 1)
	// receive and log consumer start-up errors
	go func() {
		for {
			err, ok := <-consumerStartupErrors
			if !ok {
				return
			}
			log.Printf("failed to start consumer: %s\n", err.Error())
		}
	}()

	log.Println("waiting for consumers to stop")
	wg.Wait()

	log.Println("all consumers have stopped")
}

// userCreatedHandler accepts user.created messages and logs them.
func userCreatedHandler(messageType string, messageBytes []byte) error {
	switch messageType {
	case messages.TypeUserCreated:
		message := &messages.UserCreated{}
		if err := message.WithPayload(messageBytes); err != nil {
			return err
		}

		// check for a blank username
		if message.Username == "" {
			return fmt.Errorf("missing username in message: %s", message.Type())
		}

		log.Printf("user created: %s\n", message.Username)
	default:
		return fmt.Errorf("unexpected message type %s", messageType)
	}

	return nil
}

func logConsumerErrors(consumerErrors chan error) {
	for {
		err, ok := <-consumerErrors
		if !ok {
			return
		}
		log.Printf("error received from consumer: %s\n", err.Error())
	}
}
