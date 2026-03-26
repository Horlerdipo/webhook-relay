package models

import (
	"encoding/json"
	"github.com/horlerdipo/webhook-relay/internal/enums"
	"time"
)

type WebhookEvent struct {
	RouteIdentifier string                   `json:"route_identifier,omitempty" redis:"route_identifier"`
	Identifier      string                   `json:"identifier,omitempty" redis:"identifier"`
	ReceivedAt      time.Time                `json:"received_at,omitempty" redis:"received_at"`
	Payload         json.RawMessage          `json:"payload,omitempty" redis:"payload"`
	Headers         json.RawMessage          `json:"headers,omitempty" redis:"headers"`
	Status          enums.WebhookEventStatus `json:"status,omitempty" redis:"status"`
}
