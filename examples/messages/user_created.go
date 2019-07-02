package messages

import "encoding/json"

// TypeUserCreated messages are published when a user is created
const TypeUserCreated = "user.created"

// UserCreated is the message related to TypeUserCreated
type UserCreated struct {
	Username string `json:"username"`
}

func (m *UserCreated) Type() string {
	return TypeUserCreated
}

func (m *UserCreated) Payload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *UserCreated) WithPayload(bytes []byte) error {
	return json.Unmarshal(bytes, m)
}
