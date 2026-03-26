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
	ProcessStashedWebhookEvents(ctx context.Context, routeId string, webhooks []models.WebhookEvent) error
	validateWebhookEventCredentials(ctx context.Context, routeId string, webhook models.WebhookEvent) error
	queueDeliveryAttempts(ctx context.Context, routeId string, webhook models.WebhookEvent) error
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

	_, err = ie.Datastore.StashIncomingWebhookEvent(ctx, routeId, webhook)
	if err != nil {
		return err
	}
	return nil
}

func (ie *DefaultIngressEngine) ProcessStashedWebhookEvents(ctx context.Context, routeId string, webhooks []models.WebhookEvent) error {
	//fetch webhook identifier chunk from webhook:incoming list
	//process each chunk
	//On Each chunk:
	//check if it has hash, if no has remove from list
	//validate webhook event credentials
	//if false, push to webhook:all and update status to validation_failed
	//if true, fetch destinations, XADD to stream, update status to attempted(this is queueDeliveryAttempts)
	return nil
}

func (ie *DefaultIngressEngine) validateWebhookEventCredentials(ctx context.Context, routeId string, webhook models.WebhookEvent) error {
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
