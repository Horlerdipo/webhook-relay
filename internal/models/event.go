package models

import (
	"encoding/json"
	"time"
)

type Event struct {
	RouteIdentifier string          `json:"route_identifier,omitempty" redis:"route_identifier"`
	Identifier      string          `json:"identifier,omitempty" redis:"identifier"`
	CreatedAt       time.Time       `json:"created_at,omitempty" redis:"created_at"`
	Payload         json.RawMessage `json:"payload,omitempty" redis:"payload"`
	Headers         json.RawMessage `json:"headers,omitempty" redis:"headers"`
}
