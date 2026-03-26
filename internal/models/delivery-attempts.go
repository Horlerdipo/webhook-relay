package models

import (
	"github.com/horlerdipo/webhook-relay/internal/enums"
	"time"
)

type DeliveryAttempt struct {
	Identifier             string                      `json:"identifier,omitempty" redis:"identifier"`
	WebhookEventIdentifier string                      `json:"webhook_event_identifier,omitempty" redis:"webhook_event_identifier"`
	DestinationIdentifier  string                      `json:"destination_identifier,omitempty" redis:"destination_identifier"`
	LastAttemptedAt        time.Time                   `json:"last_attempted_at,omitempty" redis:"last_attempted_at"`
	Tries                  int                         `json:"tries,omitempty" redis:"tries"`
	Status                 enums.DeliveryAttemptStatus `json:"status,omitempty" redis:"status"`
}
