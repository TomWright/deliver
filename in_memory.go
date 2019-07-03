package deliver

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"sync"
)

// inMemoryMessage defines a message published by an InMemoryPublisher.
type inMemoryMessage struct {
	// Type is the type of the message.
	Type string
	// Payload is the payload of the message.
	Payload []byte
}

// NewInMemoryPublisher returns a new in-memory message publisher.
func NewInMemoryPublisher(bufferSize int, blockWrites bool) *inMemoryPublisher {
	p := &inMemoryPublisher{
		messageChan: make(chan *inMemoryMessage, bufferSize),
		blockWrites: blockWrites,
	}
	return p
}

type inMemoryPublisher struct {
	// messageChan is used to sent messages.
	messageChan chan *inMemoryMessage
	// if blockWrites is true writes will be blocked if there are no consumers running.
	blockWrites bool
}

func (x *inMemoryPublisher) Publish(m Message) error {
	if x.messageChan == nil {
		return errors.New("cannot publish to closed publisher")
	}

	payload, err := m.Payload()
	if err != nil {
		return fmt.Errorf("could not get message payload: %s", err)
	}

	message := &inMemoryMessage{
		Type:    m.Type(),
		Payload: payload,
	}

	if x.blockWrites {
		x.messageChan <- message
	} else {
		go func() {
			x.messageChan <- message
		}()
	}

	return nil
}

// Close closes the in-memory publisher.
// When all messages already published have been handled, the consumers will shutdown.
func (x *inMemoryPublisher) Close() error {
	if x.messageChan == nil {
		return errors.New("cannot close a closed in-memory publisher")
	}
	close(x.messageChan)
	x.messageChan = nil
	return nil
}

// NewInMemorySubscriber returns a new in-memory message subscriber.
func NewInMemorySubscriber(publisher *inMemoryPublisher) Subscriber {
	p := &inMemorySubscriber{
		messageChan: publisher.messageChan,
		consumers:   make(map[string]map[string]*consumerConfig),
	}
	return p
}

type consumerConfig struct {
	ctx          context.Context
	id           string
	messageTypes []string
	consumeFn    ConsumeFn
	errChan      chan<- error
}

// handleMessageType returns true if this consumer config has subscribed to the given message type.
func (x consumerConfig) handleMessageType(messageType string) bool {
	for _, configMessageType := range x.messageTypes {
		if configMessageType == messageType {
			return true
		}
	}
	return false
}

type inMemorySubscriber struct {
	// messageChan is used to receive messages.
	messageChan               <-chan *inMemoryMessage
	consumersMu               sync.Mutex
	consumers                 map[string]map[string]*consumerConfig
	distributeEventsRunningMu sync.Mutex
	distributeEventsRunning   bool
}

// distributeEvents is executed in a go routine every time a new subscriber is added.
func (x *inMemorySubscriber) distributeEvents() {
	// only have one instance of this function running in a go routine
	x.distributeEventsRunningMu.Lock()
	if x.distributeEventsRunning {
		x.distributeEventsRunningMu.Unlock()
		return
	}
	x.distributeEventsRunning = true
	x.distributeEventsRunningMu.Unlock()

	defer func() {
		x.distributeEventsRunningMu.Lock()
		x.distributeEventsRunning = false
		x.distributeEventsRunningMu.Unlock()
	}()

	for {
		select {
		case consumerMessage, ok := <-x.messageChan:
			if !ok {
				return
			}

			x.consumersMu.Lock()
		consumerGroupLoop:
			for _, consumerGroup := range x.consumers {
			consumerLoop:
				for _, c := range consumerGroup {
					if !c.handleMessageType(consumerMessage.Type) {
						continue consumerLoop
					}
					if err := c.consumeFn(consumerMessage.Type, consumerMessage.Payload); err != nil {
						if c.errChan != nil {
							c.errChan <- err
						}
					}
					continue consumerGroupLoop
				}
			}
			x.consumersMu.Unlock()
		}
	}
}

func (x *inMemorySubscriber) Subscribe(ctx context.Context, options SubscribeOptions) error {
	if x.messageChan == nil {
		return errors.New("missing message chan, was the publisher closed?")
	}
	if err := options.Validate(); err != nil {
		return err
	}

	go x.distributeEvents()

	consumer := &consumerConfig{
		ctx:          ctx,
		id:           uuid.New().String(),
		messageTypes: options.Types,
		consumeFn:    options.ConsumeFn,
		errChan:      options.Errors,
	}

	x.consumersMu.Lock()
	if _, ok := x.consumers[options.Group]; !ok {
		x.consumers[options.Group] = make(map[string]*consumerConfig, 0)
	}
	x.consumers[options.Group][consumer.id] = consumer
	x.consumersMu.Unlock()

	// stop the consumer (delete the config) when the context is done
	<-ctx.Done()
	x.consumersMu.Lock()
	delete(x.consumers[options.Group], consumer.id)
	x.consumersMu.Unlock()

	return nil
}
