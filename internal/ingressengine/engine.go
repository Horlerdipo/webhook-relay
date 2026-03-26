package ingressengine

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/horlerdipo/webhook-relay/internal/datastore"
	"github.com/horlerdipo/webhook-relay/internal/enums"
	"github.com/horlerdipo/webhook-relay/internal/models"
)

type IngressEngine interface {
	RegisterIncomingWebhookEvent(ctx context.Context, routeId string, webhook models.WebhookEvent) error
	ProcessStashedWebhookEvents(ctx context.Context) error
	validateWebhookEventCredentials(ctx context.Context, routeId string, webhook models.WebhookEvent) error
	queueDeliveryAttempts(ctx context.Context, webhook models.WebhookEvent) error
}

type DefaultIngressEngine struct {
	Datastore datastore.Store
}

func (ie *DefaultIngressEngine) RegisterIncomingWebhookEvent(ctx context.Context, routeId string, webhook models.WebhookEvent) error {
	// check if route exist
	routeExists, err := ie.Datastore.CheckRouteExistence(ctx, routeId)
	if err != nil {
		return err
	}

	if !routeExists {
		return errors.New("route does not exist")
	}

	_, err = ie.Datastore.StashIncomingWebhookEvent(ctx, webhook)
	if err != nil {
		return err
	}
	return nil
}

func (ie *DefaultIngressEngine) ProcessStashedWebhookEvents(ctx context.Context) error {

	//fetch webhook identifier chunk from webhook:incoming list
	webhooks, err := ie.Datastore.FetchIncomingWebhookEvents(ctx, 100)
	if err != nil {
		return err
	}

	//process each chunk
	var webhookIdentifiers []string
	for _, webhook := range webhooks {
		//validate webhook event credentials
		err = ie.validateWebhookEventCredentials(ctx, webhook.Identifier, webhook)
		if err != nil {
			//if false, push to webhook:all and update status to validation_failed
			webhookIdentifiers = append(webhookIdentifiers, webhook.Identifier)
			err := ie.Datastore.UpdateWebhookEventItem(ctx, webhook.Identifier, "status", string(enums.ValidationFailed))
			if err != nil {
				//todo: log error
				continue
			}

			//todo: log error
			continue
		}

		//if true, queueDeliveryAttempts
		err = ie.queueDeliveryAttempts(ctx, webhook)
		if err != nil {
			//todo: log error
			continue
		}
	}

	err = ie.Datastore.AddWebhookEvents(ctx, webhookIdentifiers)
	if err != nil {
		//todo: log error
		return err
	}

	return nil
}

func (ie *DefaultIngressEngine) queueDeliveryAttempts(ctx context.Context, webhook models.WebhookEvent) error {
	//fetch destinations, XADD to stream, update status to attempted
	destinations, err := ie.Datastore.FetchDestinations(ctx, webhook.RouteIdentifier)
	if err != nil {
		//todo: log error
		return err
	}

	err = ie.Datastore.QueueDeliveryAttempts(ctx, webhook, destinations)
	if err != nil {
		//todo: log error
		return err
	}

	return nil
}

func (ie *DefaultIngressEngine) validateWebhookEventCredentials(ctx context.Context, routeId string, webhook models.WebhookEvent) error {
	//check if webhook event has not been attempted already
	if webhook.Status != enums.Pending {
		return errors.New("webhook status must be pending")
	}

	// check if route exist and is active
	routeExists, err := ie.Datastore.CheckRouteExistence(ctx, routeId)
	if err != nil {
		return err
	}

	if !routeExists {
		return errors.New("route does not exist")
	}

	route, err := ie.Datastore.FetchRoute(ctx, routeId, false)
	if err != nil {
		return err
	}

	if !route.Active {
		return errors.New("route is currently not active")
	}

	//validate incoming webhook
	if route.VerificationType == enums.None {
		return nil
	}

	if route.VerificationType == enums.StaticToken {
		return ie.verifyStaticToken(&route, &webhook)
	}

	if route.VerificationType == enums.RequestSigning {
		return ie.verifyRequestSigningToken(&route, &webhook)
	}

	return nil
}

func (ie *DefaultIngressEngine) verifyStaticToken(route *models.Route, webhook *models.WebhookEvent) error {
	var payload json.RawMessage
	var decodedPayload map[string]interface{}

	if route.VerificationKeyLocation == enums.Body {
		payload = webhook.Payload
	} else {
		payload = webhook.Headers
	}

	err := json.Unmarshal(payload, &decodedPayload)
	if err != nil {
		return err
	}
	token := decodedPayload[route.VerificationKeyName].(string)

	if token != route.VerificationToken {
		return errors.New("unable to verify webhook static token")
	}
	return nil
}

func (ie *DefaultIngressEngine) verifyRequestSigningToken(route *models.Route, webhook *models.WebhookEvent) error {
	var payload json.RawMessage
	var decodedPayload map[string]interface{}

	if route.VerificationKeyLocation == enums.Body {
		payload = webhook.Payload
	} else {
		payload = webhook.Headers
	}

	err := json.Unmarshal(payload, &decodedPayload)
	if err != nil {
		return err
	}

	sentHMAC := decodedPayload[route.VerificationKeyName].(string)

	mac := hmac.New(sha256.New, []byte(route.VerificationToken))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	fmt.Printf("Generated HMAC: %x\n", expectedMAC)

	// Verify the HMAC
	if sentHMAC != expectedMAC {
		return errors.New("unable to verify webhook signing token")
	}
	return nil
}
