package testutils

import (
	"github.com/google/uuid"
	"github.com/horlerdipo/webhook-relay/internal/enums"
	"github.com/horlerdipo/webhook-relay/internal/models"
	"time"
)

func NewRouteModel() models.Route {
	routeUuid := uuid.NewString()
	return models.Route{
		HttpMethod:              enums.Post,
		Name:                    routeUuid,
		Identifier:              routeUuid,
		VerificationType:        enums.None,
		VerificationKeyLocation: enums.NoLocation,
		VerificationKeyName:     "none",
		VerificationToken:       routeUuid,
		Active:                  true,
		Destinations:            nil,
	}
}

func NewDestinationModel(routeId string) models.Destination {
	destinationUuid := uuid.NewString()
	return models.Destination{
		HttpMethod:        enums.Post,
		Identifier:        destinationUuid,
		Active:            true,
		Url:               "https://google.com",
		VerificationToken: destinationUuid,
		RouteIdentifier:   routeId,
	}
}

func NewWebhookEventModel(routeId string) models.WebhookEvent {
	id := uuid.NewString()

	return models.WebhookEvent{
		RouteIdentifier: routeId,
		Identifier:      id,
		ReceivedAt:      time.Time{},
		Payload:         nil,
		Headers:         nil,
		Status:          enums.Pending,
	}
}
