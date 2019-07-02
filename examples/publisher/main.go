package main

import (
	"fmt"
	"github.com/tomwright/deliver"
	"github.com/tomwright/deliver/examples/messages"
	"log"
	"time"
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
			panic("could not close publisher: " + err.Error())
		}
	}()

	// publish messages with a blank Username
	go func() {
		for {
			var message deliver.Message = &messages.UserCreated{
				Username: "",
			}

			// publish the message
			if err := publisher.Publish(message); err != nil {
				// handle error
				panic("could not publish message: " + err.Error())
			}
			// message published successfully

			time.Sleep(time.Millisecond * 3500)
		}
	}()

	// publish messages with different Username's
	for i := 1; i < 100; i++ {
		var message deliver.Message = &messages.UserCreated{
			Username: fmt.Sprintf("Tom%d", i),
		}

		// publish the message
		if err := publisher.Publish(message); err != nil {
			// handle error
			panic("could not publish message: " + err.Error())
		}
		// message published successfully

		time.Sleep(time.Second)
	}

	log.Println("finished publishing")
}
