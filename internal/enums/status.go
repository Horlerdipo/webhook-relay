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

type DeliveryAttemptStatus string

const (
	Queued     DeliveryAttemptStatus = "queued"
	Processing DeliveryAttemptStatus = "processing"
	Success    DeliveryAttemptStatus = "success"
	Failed     DeliveryAttemptStatus = "failed"
)

func (h DeliveryAttemptStatus) MarshalBinary() ([]byte, error) {
	return []byte(h), nil
}

func ParseDeliveryAttemptStatus(s string) (DeliveryAttemptStatus, error) {
	switch s {
	case "queued":
		return Queued, nil
	case "processing":
		return Processing, nil
	case "success":
		return Success, nil
	case "failed":
		return Failed, nil
	default:
		return "", fmt.Errorf("invalid delivery attempt status: %s", s)
	}
}
