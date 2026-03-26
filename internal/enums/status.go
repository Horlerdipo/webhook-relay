package enums

import "fmt"

type WebhookEventStatus string

const (
	Pending          WebhookEventStatus = "pending"
	Attempted        WebhookEventStatus = "attempted"
	ValidationFailed WebhookEventStatus = "validation_failed"
)

func (h WebhookEventStatus) MarshalBinary() ([]byte, error) {
	return []byte(h), nil
}

func ParseWebhookEventStatus(s string) (WebhookEventStatus, error) {
	switch s {
	case "pending":
		return Pending, nil
	case "attempted":
		return Attempted, nil
	case "validation_failed":
		return ValidationFailed, nil
	default:
		return "", fmt.Errorf("invalid webhook event status: %s", s)
	}
}
