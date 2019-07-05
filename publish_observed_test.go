package deliver

import (
	"encoding/json"
	"testing"
)

type testMessage struct {
	ID string `json:"id"`
}

func (m *testMessage) Type() string {
	return "test"
}

func (m *testMessage) Payload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *testMessage) WithPayload(bytes []byte) error {
	return json.Unmarshal(bytes, m)
}

func TestObservedPublisher_Publish(t *testing.T) {
	m1 := &testMessage{ID:"1"}
	m2 := &testMessage{ID:"2"}
	m3 := &testMessage{ID:"3"}

	p := NewObservedPublisher()
	if err := p.Publish(m1); err != nil {
		t.Errorf("unexpected error when publishing message: %s", err)
		return
	}
	if err := p.Publish(m2); err != nil {
		t.Errorf("unexpected error when publishing message: %s", err)
		return
	}
	if err := p.Publish(m3); err != nil {
		t.Errorf("unexpected error when publishing message: %s", err)
		return
	}

	messages := p.Messages()
	if exp, got := 3, len(messages); exp != got {
		t.Errorf("expected %d stored messages, got %d", exp, got)
		return
	}

	if exp, got := "1", messages[0].(*testMessage).ID; exp != got {
		t.Errorf("expected message id %s, got %s", exp, got)
		return
	}
	if exp, got := "2", messages[1].(*testMessage).ID; exp != got {
		t.Errorf("expected message id %s, got %s", exp, got)
		return
	}
	if exp, got := "3", messages[2].(*testMessage).ID; exp != got {
		t.Errorf("expected message id %s, got %s", exp, got)
		return
	}
}

func TestObservedPublisher_Clear(t *testing.T) {
	m1 := &testMessage{ID:"1"}
	m2 := &testMessage{ID:"2"}
	m3 := &testMessage{ID:"3"}

	p := NewObservedPublisher()
	if err := p.Publish(m1); err != nil {
		t.Errorf("unexpected error when publishing message: %s", err)
		return
	}
	if err := p.Publish(m2); err != nil {
		t.Errorf("unexpected error when publishing message: %s", err)
		return
	}
	if err := p.Publish(m3); err != nil {
		t.Errorf("unexpected error when publishing message: %s", err)
		return
	}

	messages := p.Messages()
	if exp, got := 3, len(messages); exp != got {
		t.Errorf("expected %d stored messages, got %d", exp, got)
		return
	}

	if err := p.Clear(); err != nil {
		t.Errorf("unexpected error when clearing publisher messages: %s", err)
		return
	}

	messages = p.Messages()
	if exp, got := 0, len(messages); exp != got {
		t.Errorf("expected %d stored messages, got %d", exp, got)
		return
	}
}
