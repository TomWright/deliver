package deliver

import (
	"context"
	"github.com/pkg/errors"
	"sync"
)

// ConsumeFn is a function to handle a consumed message.
type ConsumeFn func(messageType string, messageBytes []byte) error

// SubscribeArgs contains a set of arguments used when Subscribing to Messages.
type SubscribeOptions struct {
	// ConsumeFn is the function to handle the consumed messages.
	ConsumeFn ConsumeFn
	// A message will only be consumed once per group.
	Group string
	// Types is the set of messages types to subscribe the ConsumeFn to.
	Types []string
	// IgnoreErrors defines whether or not errors returned from ConsumeFn will be written to Errors.
	// If this is false, a value must be provided for Errors.
	IgnoreErrors bool
	// Errors will receive any errors returned from ConsumeFn, if IgnoreErrors is false.
	Errors chan<- error
}

// Validate makes sure we have a set of valid options and applies defaults.
func (x *SubscribeOptions) Validate() error {
	if x.Group == "" {
		x.Group = "default"
	}

	var err error
	switch {
	case x.ConsumeFn == nil:
		err = errors.New("missing consumer function")
	case len(x.Types) == 0:
		err = errors.New("no message types to subscribe to")
	case x.IgnoreErrors && x.Errors != nil:
		err = errors.New("ignore errors is on but error channel was provided")
	case !x.IgnoreErrors && x.Errors == nil:
		err = errors.New("ignore errors is off but no error channel was provided")
	}

	return err
}

// Subscriber defines an interface that can be used to consume messages.
type Subscriber interface {
	// Subscribe starts a consumer with the given context.
	//
	// If an error is returned then the consumer has not been started, otherwise you should listen
	// on the errChan and handle any consumer errors.
	//
	// The consumer will be stopped when the given context is cancelled.
	Subscribe(ctx context.Context, options SubscribeOptions) error
}

// SubscribeNonBlocking allows you to start many consumers with the same SubscribeOptions, in a non-blocking way.
// If nothing is reading from errChan this function will be blocked and you will
// not be notified if any consumers weren't started.
func SubscribeNonBlocking(ctx context.Context, options SubscribeOptions, subscriber Subscriber, wg *sync.WaitGroup, consumerCount int) chan error {
	errChan := make(chan error)

	// increment the wait group
	wg.Add(consumerCount)

	for i := 0; i < consumerCount; i++ {
		// start the consumer in a new go routine
		go func() {
			// when this function returns, the consumer has stopped running
			defer wg.Done()

			// block here until the consumer stops running
			err := subscriber.Subscribe(ctx, options)
			if err != nil {
				// write any start-up errors to errChan.
				errChan <- err
			}
		}()
	}

	// ensure errChan is closed once all consumers using it have stopped
	go func() {
		wg.Wait()
		close(errChan)
	}()

	return errChan
}
