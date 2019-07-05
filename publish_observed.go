package deliver

import "sync"

// NewObservedPublisher returns a new publisher that can be used when testing.
func NewObservedPublisher() *ObservedPublisher {
	p := new(ObservedPublisher)
	p.messages = make([]Message, 0)

	return p
}

// ObservedPublisher can be used to test the way an application publishes messages.
type ObservedPublisher struct {
	mu sync.RWMutex
	messages []Message
}

// Publish stores the given message internally so it can be retrieved later on.
func (p *ObservedPublisher) Publish(m Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append(p.messages, m)

	return nil
}

// Close performs no action.
func (p *ObservedPublisher) Close() error {
	return nil
}

// Clear clears out any stored messages.
func (p *ObservedPublisher) Clear() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = make([]Message, 0)

	return nil
}

// Messages returns all stored messages.
func (p *ObservedPublisher) Messages() []Message {
	p.mu.RLock()
	messages := p.messages
	p.mu.RUnlock()

	return messages
}
