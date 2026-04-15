package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"eventId"`
	UserID    uuid.UUID `json:"userId"`
	Title     string    `json:"title"`
	EventTime time.Time `json:"eventTime"`
}

func (n *Notification) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

func (n *Notification) FromJSON(data []byte) error {
	return json.Unmarshal(data, n)
}
