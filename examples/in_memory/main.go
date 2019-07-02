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

	publisher := deliver.NewInMemoryPublisher(0, false)
	subscriber := deliver.NewInMemorySubscriber(publisher)

	// make sure we close the publisher when we're finished
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("could not close publisher: %s\n" + err.Error())
		}
	}()

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
	startConsumer := func() {
		// when this function returns, the consumer has stopped running
		defer wg.Done()

		// only run the consumer for 10 seconds
		ctx, _ = context.WithTimeout(ctx, time.Second*10)

		// block here until the consumer stops running
		err := subscriber.Subscribe(ctx, userCreatedHandler, "message-logger", consumerErrors, messages.TypeUserCreated)
		if err != nil {
			log.Printf("could not start consumer: %s\n", err.Error())
			return
		}
	}
	wg.Add(1) // add to the WaitGroup for the consumer
	go startConsumer()
	wg.Add(1) // add to the WaitGroup for the consumer
	go startConsumer()

	// publish some messages
	go func() {
		for i := 0; i < 20; i++ {
			m := &messages.UserCreated{
				Username: fmt.Sprintf("Tom%d", i),
			}
			if i%5 == 0 {
				m.Username = ""
			}
			if err := publisher.Publish(m); err != nil {
				log.Printf("publish message failed: %s\n", err.Error())
			}
			time.Sleep(time.Millisecond * 500)
		}
		// close the publisher when our work is done
		if err := publisher.Close(); err != nil {
			log.Printf("could not close publisher: %s\n", err)
		}
		log.Println("publisher stopped")
	}()

	log.Println("waiting for consumers to stop")
	wg.Wait()

	log.Println("all consumers have stopped")
}
