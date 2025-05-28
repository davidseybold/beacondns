package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID       `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"createdAt"`
}

func NewEvent(eventType string, payload any) *Event {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return &Event{
		ID:        uuid.New(),
		Type:      eventType,
		Payload:   payloadJSON,
		CreatedAt: time.Now(),
	}
}
