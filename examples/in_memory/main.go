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

	// start up a go routine to listen for and log errors returned from the consumer
	consumerErrors := make(chan error)
	go logConsumerErrors(consumerErrors)

	// create a WaitGroup so we know when all consumers have stopped
	wg := &sync.WaitGroup{}

	// start 2 consumers that will stop after 10 seconds
	consumerCtx, _ := context.WithTimeout(ctx, time.Second*10)
	wg.Add(4) // add to the WaitGroup for the consumers
	go startConsumer(consumerCtx, wg, subscriber, consumerErrors)
	go startConsumer(consumerCtx, wg, subscriber, consumerErrors)
	go startConsumer(consumerCtx, wg, subscriber, consumerErrors)
	go startConsumer(consumerCtx, wg, subscriber, consumerErrors)

	// publish some messagesw
	go publishMessages(publisher)

	log.Println("waiting for consumers to stop")
	wg.Wait()

	// close consumerErrors channel now all consumers using it have stopped
	close(consumerErrors)

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

func publishMessages(publisher deliver.Publisher) {
	// publish messages with different Username's
	for i := 1; i < 100; i++ {
		message := &messages.UserCreated{
			Username: fmt.Sprintf("Tom%d", i),
		}
		if i%5 == 0 {
			message.Username = ""
		}

		// publish the message
		if err := publisher.Publish(message); err != nil {
			log.Printf("could not publish message: %s\n", err.Error())
		}
		// message published successfully

		time.Sleep(time.Millisecond * 5)
	}
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

// subscribe the handler function to receive user created messages.
func startConsumer(ctx context.Context, wg *sync.WaitGroup, subscriber deliver.Subscriber, consumerErrors chan error) {
	// when this function returns, the consumer has stopped running
	defer wg.Done()

	// block here until the consumer stops running
	err := subscriber.Subscribe(ctx, deliver.SubscribeOptions{
		ConsumeFn: userCreatedHandler,
		Group:     "message-logger",
		Types:     []string{messages.TypeUserCreated},
		Errors:    consumerErrors,
	})
	if err != nil {
		log.Printf("could not start consumer: %s\n", err.Error())
		return
	}
}
