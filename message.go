package deliver

// Message is a message that can be published and consumed.
type Message interface {
	// Type returns the type of the message.
	Type() string

	// Payload returns the message payload.
	Payload() ([]byte, error)

	// WithPayload validates and sets the given payload on the message.
	WithPayload(payload []byte) error
}
