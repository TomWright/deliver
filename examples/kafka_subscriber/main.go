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

	// define a handler function to log user created messages
	userCreatedHandler := func(messageType string, messageBytes []byte) error {
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

	// start up a go routine to listen for and log errors returned from the consumer
	consumerErrors := make(chan error)
	go func() {
		for {
			err, ok := <-consumerErrors
			if !ok {
				return
			}
			log.Printf("error received from consumer: %s\n", err.Error())
		}
	}()

	// create a WaitGroup so we know when all consumers have stopped
	wg := &sync.WaitGroup{}

	// subscribe the handler function to receive user created messages.
	// do this in a go routine so as we don't block the main routine
	wg.Add(1) // add to the WaitGroup for the consumer
	go func() {
		// when this function returns, the consumer has stopped running
		defer wg.Done()

		// only run the consumer for 10 seconds
		ctx, _ = context.WithTimeout(ctx, time.Second*10)

		// block here until the consumer stops running
		err := subscriber.Subscribe(ctx, userCreatedHandler, "message-logger", consumerErrors, messages.TypeUserCreated)
		if err != nil {
			log.Printf("could not start consumers: %s\n", err.Error())
			return
		}
	}()

	log.Println("waiting for consumers to stop")
	wg.Wait()

	log.Println("all consumers have stopped")
}
