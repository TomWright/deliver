package deliver

// Subscriber defines an interface that can be used to publish messages.
type Publisher interface {
	// Publish publishes the given message.
	// If an error is returned, the message has not been published.
	Publish(m Message) error

	// Close closes any open connections.
	Close() error
}
