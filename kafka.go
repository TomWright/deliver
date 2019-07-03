package deliver

import (
	"context"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"sync"
	"time"
)

// NewKafkaSubscriber returns a Subscriber that will consume from kafka.
func NewKafkaSubscriber(brokers []string) Subscriber {
	config := sarama.NewConfig()
	config.Version = sarama.V1_1_0_0
	config.Consumer.Return.Errors = true
	return &kafkaSubscriber{
		consumerGroups: make(map[string]sarama.ConsumerGroup),
		config:         config,
		brokers:        brokers,
	}
}

type kafkaSubscriber struct {
	consumerGroups map[string]sarama.ConsumerGroup
	config         *sarama.Config
	brokers        []string
}

// Subscribe starts a consumer to handle the given message types, under the given
// consumer group. This is a blocking action.
//
// Messages will only be handled once per consumer group.
//
// If an error is returned then the consumer has not been started, otherwise you should listen
// on the errChan and handle any consumer errors.
func (x *kafkaSubscriber) Subscribe(ctx context.Context, options SubscribeOptions) error {
	if err := options.Validate(); err != nil {
		return err
	}

	// Create the in-memory consumer group if it doesn't exists
	if _, exists := x.consumerGroups[options.Group]; !exists {
		cg, err := sarama.NewConsumerGroup(x.brokers, options.Group, x.config)
		if err != nil {
			return err
		}
		x.consumerGroups[options.Group] = cg
	}

	stopConsumingMu := &sync.Mutex{}
	stopConsuming := false

	// start up a go routine to set stopConsuming to true when the context is done
	go func() {
		<-ctx.Done()
		stopConsumingMu.Lock()
		stopConsuming = true
		stopConsumingMu.Unlock()
	}()

	// start up a go routine to receive errors from the consumers and write them to the error chan
	go func() {
		select {
		case <-ctx.Done():
			// if context is done, stop the routine
			return

		case err, ok := <-x.consumerGroups[options.Group].Errors():
			if !ok {
				// if error chan is closed, stop the routine
				return
			}
			if options.Errors != nil {
				options.Errors <- err
			}
		}
	}()

consumeLoop:
	for {
		stopConsumingMu.Lock()
		breakLoop := stopConsuming
		stopConsumingMu.Unlock()
		if breakLoop {
			break consumeLoop
		}

		h := &kafkaConsumerGroupHandler{
			consumeFn: options.ConsumeFn,
			errChan:   options.Errors,
		}

		if err := x.consumerGroups[options.Group].Consume(ctx, options.Types, h); err != nil {
			return fmt.Errorf("could not start consuming: %s", err)
		}
	}

	// Make sure the consumer group is closed
	if consumer, ok := x.consumerGroups[options.Group]; ok {
		if err := consumer.Close(); err != nil {
			return fmt.Errorf("could not close consumer group: %s: %s", options.Group, err.Error())
		}
	} else {
		return fmt.Errorf("consumer group did not exist when attempting to close: %s", options.Group)
	}

	return nil
}

type kafkaConsumerGroupHandler struct {
	consumeFn ConsumeFn
	errChan   chan<- error
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (x *kafkaConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (x *kafkaConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (x *kafkaConsumerGroupHandler) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {
	for {
		select {
		case consumerMessage, ok := <-c.Messages():
			if !ok {
				return nil
			}
			s.MarkMessage(consumerMessage, "")

			if err := x.consumeFn(consumerMessage.Topic, consumerMessage.Value); err != nil {
				x.errChan <- err
			}
		}
	}
}

func NewKafkaPublisher(brokerList []string) (Publisher, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V1_1_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Producer.Timeout = time.Second
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, err
	}
	p := &kafkaPublisher{
		producer: producer,
	}
	return p, nil
}

type kafkaPublisher struct {
	producer sarama.SyncProducer
}

func (x *kafkaPublisher) Publish(m Message) error {
	if x.producer == nil {
		return errors.New("cannot publish on closed publisher")
	}

	payload, err := m.Payload()
	if err != nil {
		return fmt.Errorf("could not get message payload: %s", err)
	}
	_, _, err = x.producer.SendMessage(&sarama.ProducerMessage{
		Topic: m.Type(),
		Value: sarama.StringEncoder(payload),
	})

	if err != nil {
		return err
	}

	return nil
}

func (x *kafkaPublisher) Close() error {
	if x.producer == nil {
		return errors.New("missing producer cannot be closed")
	}
	return x.producer.Close()
}
