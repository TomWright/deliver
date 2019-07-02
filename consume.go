package deliver

import (
	"context"
)

// ConsumeFn is a function to handle a consumed message.
type ConsumeFn func(messageType string, messageBytes []byte) error

// Subscriber defines an interface that can be used to consume messages.
type Subscriber interface {
	// Subscribe starts a consumer to handle the given message types, under the given
	// consumer group. This is a blocking action.
	//
	// Messages will only be handled once per consumer group.
	//
	// If an error is returned then the consumer has not been started, otherwise you should listen
	// on the errChan and handle any consumer errors.
	//
	// The consumer will be stopped when the given context is cancelled.
	Subscribe(ctx context.Context, fn ConsumeFn, consumerGroup string, errChan chan error, messageTypes ...string) error
}
