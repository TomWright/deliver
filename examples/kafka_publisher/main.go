package main

import (
	"fmt"
	"github.com/tomwright/deliver"
	"github.com/tomwright/deliver/examples/messages"
	"log"
)

func main() {
	// initialise the publisher
	publisher, err := deliver.NewKafkaPublisher([]string{"cmg-local-kafka:9092"})
	if err != nil {
		panic(err)
	}

	// make sure we close the publisher when we're finished
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("could not close publisher: %s\n", err.Error())
		}
	}()

	publishMessages(publisher)

	log.Println("finished publishing")
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
	}
}
