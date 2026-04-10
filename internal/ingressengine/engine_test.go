package ingressengine

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/horlerdipo/webhook-relay/internal/datastore"
	"github.com/horlerdipo/webhook-relay/internal/enums"
	"github.com/horlerdipo/webhook-relay/internal/testutils"
	"testing"
)

func TestDefaultIngressEngine_RegisterIncomingWebhookEvent(t *testing.T) {
	redisStore, redisSrv, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()
	engine := NewDefaultIngressEngine(redisStore)

	ctx := context.Background()
	route := testutils.NewRouteModel()
	webhook := testutils.NewWebhookEventModel(route.Identifier)
	var webhooks []string
	webhooks = append(webhooks, webhook.Identifier)
	_, err := engine.Datastore.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("error adding route: %v", err)
	}

	err = engine.Datastore.AddWebhookEvents(ctx, webhooks)
	if err != nil {
		t.Fatalf("error adding webhook events: %v", err)
	}

	err = engine.RegisterIncomingWebhookEvent(ctx, "unknown-id", webhook)
	if err == nil {
		t.Fatal("RegisterIncomingWebhookEvent() should have errored")
	}

	errString := "route does not exist"
	if err.Error() != errString {
		t.Fatalf("expected error to be `%s`, got %s", errString, err.Error())
	}

	err = engine.RegisterIncomingWebhookEvent(ctx, route.Identifier, webhook)
	if err != nil {
		t.Fatalf("expected error to be nil, got %s", err.Error())
	}

	val, err := redisSrv.RPop(datastore.IncomingWebhookEventsKey)
	if err != nil {
		t.Fatalf("expected error to be nil, got %s", err.Error())
	}

	if val != webhook.Identifier {
		t.Fatalf("expected incoming webhook events list to have %s, got %s", webhook.Identifier, val)
	}

	identifier := redisSrv.HGet(fmt.Sprintf("%s%s", datastore.WebhookEventKey, webhook.Identifier), "identifier")
	if identifier != webhook.Identifier {
		t.Fatalf("expected identifier to be '%s' got '%s'", webhook.Identifier, identifier)
	}
}

func TestDefaultIngressEngine_validateWebhookEventCredentials(t *testing.T) {
	redisStore, redisServ, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()
	engine := NewDefaultIngressEngine(redisStore)

	ctx := context.Background()
	route := testutils.NewRouteModel()
	webhook := testutils.NewWebhookEventModel(route.Identifier)
	var webhooks []string
	webhooks = append(webhooks, webhook.Identifier)
	_, err := engine.Datastore.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("error adding route: %v", err)
	}

	err = engine.Datastore.AddWebhookEvents(ctx, webhooks)
	if err != nil {
		t.Fatalf("error adding webhook events: %v", err)
	}

	webhook.Status = enums.Attempted
	err = engine.validateWebhookEventCredentials(ctx, route.Identifier, webhook)
	if err == nil {
		t.Fatal("validateWebhookEventCredentials() should have errored")
	}

	errString := "webhook status must be pending"
	if err.Error() != errString {
		t.Fatalf("expected error to be `%s`, got %s", errString, err.Error())
	}

	webhook.Status = enums.Pending
	err = engine.validateWebhookEventCredentials(ctx, "unknown-id", webhook)
	if err == nil {
		t.Fatal("validateWebhookEventCredentials() should have errored")
	}

	redisServ.HSet(fmt.Sprintf("%s%s", datastore.RouteKey, route.Identifier), "active", "false")
	err = engine.validateWebhookEventCredentials(ctx, route.Identifier, webhook)
	if err == nil {
		t.Fatal("validateWebhookEventCredentials() should have errored")
	}
	errString = "route is currently not active"
	if err.Error() != errString {
		t.Fatalf("expected error to be `%s`, got %s", errString, err.Error())
	}

	redisServ.HSet(fmt.Sprintf("%s%s", datastore.RouteKey, route.Identifier), "active", "true")
	redisServ.HSet(fmt.Sprintf("%s%s", datastore.RouteKey, route.Identifier), "verification_type", string(enums.None))
	err = engine.validateWebhookEventCredentials(ctx, route.Identifier, webhook)
	if err != nil {
		t.Fatalf("expected error to be nil, got %s", err.Error())
	}

	redisServ.HSet(fmt.Sprintf("%s%s", datastore.RouteKey, route.Identifier), "active", "true")
	redisServ.HSet(fmt.Sprintf("%s%s", datastore.RouteKey, route.Identifier), "verification_type", string(enums.None))
	err = engine.validateWebhookEventCredentials(ctx, route.Identifier, webhook)
	if err != nil {
		t.Fatalf("expected error to be nil, got %s", err.Error())
	}
}

func TestDefaultIngressEngine_validateWebhookEventCredentialsStaticToken(t *testing.T) {
	redisStore, redisServ, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()
	engine := NewDefaultIngressEngine(redisStore)
	staticToken := uuid.NewString()

	ctx := context.Background()
	route := testutils.NewRouteModel()
	route.Active = true
	route.VerificationToken = staticToken
	route.VerificationKeyLocation = enums.Header
	route.VerificationKeyName = "Authorization"
	route.VerificationType = enums.StaticToken
	webhook := testutils.NewWebhookEventModel(route.Identifier)

	_, err := engine.Datastore.AddRoute(ctx, route)

	t.Run("should correctly validate static token on headers", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": staticToken,
		}

		headerJson, _ := json.Marshal(headers)
		webhook.Headers = headerJson
		redisServ.HSet(
			fmt.Sprintf("%s%s", datastore.WebhookEventKey, webhook.Identifier),
			"headers", string(headerJson),
		)
		err = engine.validateWebhookEventCredentials(ctx, route.Identifier, webhook)
		if err != nil {
			t.Fatalf("expected error to be nil, got %s", err.Error())
		}

	})

	t.Run("should return error when static token on header is incorrect", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": staticToken + "-nonsense",
		}

		headerJson, _ := json.Marshal(headers)
		webhook.Headers = headerJson
		redisServ.HSet(
			fmt.Sprintf("%s%s", datastore.WebhookEventKey, webhook.Identifier),
			"headers", string(headerJson),
		)
		err = engine.validateWebhookEventCredentials(ctx, route.Identifier, webhook)
		errString := "unable to verify webhook static token"
		if err == nil {
			t.Fatalf("expected error to be %s, got nil", errString)
		}

		if err.Error() != errString {
			t.Fatalf("expected error to be %s, got %s", errString, err.Error())
		}
	})

	t.Run("should correctly validate static token on payload", func(t *testing.T) {
		body := map[string]string{
			"Authorization": staticToken,
		}

		bodyJson, _ := json.Marshal(body)
		webhook.Payload = bodyJson
		redisServ.HSet(
			fmt.Sprintf("%s%s", datastore.WebhookEventKey, webhook.Identifier),
			"payload", string(bodyJson),
		)
		redisServ.HSet(
			fmt.Sprintf("%s%s", datastore.RouteKey, route.Identifier),
			"verification_key_location", string(enums.Body),
		)
		err = engine.validateWebhookEventCredentials(ctx, route.Identifier, webhook)
		if err != nil {
			t.Fatalf("expected error to be nil, got %s", err.Error())
		}
	})

	t.Run("should return error when static token on payload is incorrect", func(t *testing.T) {
		body := map[string]string{
			"Authorization": staticToken + "-nonsense",
		}

		bodyJson, _ := json.Marshal(body)
		webhook.Payload = bodyJson
		redisServ.HSet(
			fmt.Sprintf("%s%s", datastore.WebhookEventKey, webhook.Identifier),
			"payload", string(bodyJson),
		)
		redisServ.HSet(
			fmt.Sprintf("%s%s", datastore.RouteKey, route.Identifier),
			"verification_key_location", string(enums.Body),
		)
		err = engine.validateWebhookEventCredentials(ctx, route.Identifier, webhook)
		errString := "unable to verify webhook static token"
		if err == nil {
			t.Fatalf("expected error to be %s, got nil", errString)
		}

		if err.Error() != errString {
			t.Fatalf("expected error to be %s, got %s", errString, err.Error())
		}
	})
}

func TestDefaultIngressEngine_validateWebhookEventCredentialsRequestSigning(t *testing.T) {
	redisStore, redisServ, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()
	engine := NewDefaultIngressEngine(redisStore)
	secret := uuid.NewString()

	ctx := context.Background()
	route := testutils.NewRouteModel()
	route.Active = true
	route.VerificationToken = secret
	route.VerificationKeyName = "Signature"
	route.VerificationType = enums.RequestSigning

	_, err := engine.Datastore.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("error adding route: %v", err)
	}

	tests := []struct {
		name                 string
		keyLocation          enums.VerificationKeyLocation
		payload              map[string]interface{}
		headers              map[string]string
		modifySignature      func(sig string) string // if provided, mutate signature to produce failure cases
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:            "header_success",
			keyLocation:     enums.Header,
			payload:         map[string]interface{}{"foo": "bar"},
			headers:         nil,
			modifySignature: nil,
			expectError:     false,
		},
		{
			name:                 "header_failure_bad_sig",
			keyLocation:          enums.Header,
			payload:              map[string]interface{}{"foo": "bar"},
			headers:              nil,
			modifySignature:      func(sig string) string { return sig + "-bad" },
			expectError:          true,
			expectedErrorMessage: "unable to verify webhook signing token",
		},
		{
			name:            "body_success",
			keyLocation:     enums.Body,
			payload:         map[string]interface{}{"foo": "bar"},
			headers:         nil,
			modifySignature: nil,
			expectError:     false,
		},
		{
			name:                 "body_failure_bad_sig",
			keyLocation:          enums.Body,
			payload:              map[string]interface{}{"foo": "bar"},
			headers:              nil,
			modifySignature:      func(sig string) string { return "wrong" },
			expectError:          true,
			expectedErrorMessage: "unable to verify webhook signing token",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			r := testutils.NewRouteModel()
			r.Identifier = route.Identifier
			r.Active = true
			r.VerificationToken = secret
			r.VerificationKeyLocation = tc.keyLocation
			r.VerificationKeyName = route.VerificationKeyName
			r.VerificationType = enums.RequestSigning

			redisServ.HSet(fmt.Sprintf("%s%s", datastore.RouteKey, r.Identifier), "verification_key_location", string(tc.keyLocation))

			webhook := testutils.NewWebhookEventModel(r.Identifier)

			//compute message bytes and signature
			//For header: message is the payload
			//For body: message is the payload without the Signature field
			var messageBytes []byte
			var err error
			// for body, message will be payload serialized (without sig)
			messageBytes, err = json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("marshal payload: %v", err)
			}

			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(messageBytes)
			sig := hex.EncodeToString(mac.Sum(nil))
			if tc.modifySignature != nil {
				sig = tc.modifySignature(sig)
			}

			if tc.keyLocation == enums.Header {
				headers := map[string]string{r.VerificationKeyName: sig}
				headerJson, _ := json.Marshal(headers)
				webhook.Headers = headerJson
				webhook.Payload = messageBytes

				redisServ.HSet(fmt.Sprintf("%s%s", datastore.WebhookEventKey, webhook.Identifier), "headers", string(headerJson))
				redisServ.HSet(fmt.Sprintf("%s%s", datastore.WebhookEventKey, webhook.Identifier), "payload", string(messageBytes))
			} else {

				bodyWithSig := make(map[string]interface{})
				for k, v := range tc.payload {
					bodyWithSig[k] = v
				}
				bodyWithSig[r.VerificationKeyName] = sig
				bodyJson, _ := json.Marshal(bodyWithSig)
				webhook.Payload = bodyJson
				redisServ.HSet(fmt.Sprintf("%s%s", datastore.WebhookEventKey, webhook.Identifier), "payload", string(bodyJson))
			}

			err = engine.validateWebhookEventCredentials(ctx, r.Identifier, webhook)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if tc.expectedErrorMessage != "" && err.Error() != tc.expectedErrorMessage {
					t.Fatalf("expected error to be %s, got %s", tc.expectedErrorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			}
		})
	}
}

func TestDefaultIngressEngine_queueDeliveryAttempts(t *testing.T) {
	redisStore, _, cleanup := testutils.NewTestRedisStore(t)
	defer cleanup()
	engine := NewDefaultIngressEngine(redisStore)

	ctx := context.Background()
	route := testutils.NewRouteModel()
	route.Active = true
	route.VerificationToken = "secret"
	route.VerificationKeyName = "Signature"
	route.VerificationType = enums.RequestSigning

	_, err := engine.Datastore.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("error adding route: %v", err)
	}

	webhook := testutils.NewWebhookEventModel(route.Identifier)
	_, err = engine.Datastore.StashIncomingWebhookEvent(ctx, webhook)
	if err != nil {
		t.Fatalf("error adding webhook events: %v", err)
	}

	destination := testutils.NewDestinationModel(route.Identifier)
	_, err = engine.Datastore.AddDestination(ctx, route.Identifier, destination)
	if err != nil {
		t.Fatalf("error adding destination: %v", err)
	}

	t.Run("should queue delivery attempts for all active destinations into the Redis stream", func(t *testing.T) {
		err = engine.queueDeliveryAttempts(ctx, webhook)
		if err != nil {
			t.Fatalf("expected error to be nil, got %s", err.Error())
		}

		//_, _ = redisSrv.XLen(datastore.DeliveryAttemptsQueueKey).Result()
	})

}
