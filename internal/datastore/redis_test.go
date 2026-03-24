package datastore

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/horlerdipo/webhook-relay/internal/enums"
	"github.com/horlerdipo/webhook-relay/internal/models"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisStore(t *testing.T) (*RedisStore, *miniredis.Miniredis, func()) {
	t.Helper()
	srv := miniredis.RunT(t)

	opt := &redis.Options{Addr: srv.Addr()}
	client := redis.NewClient(opt)
	rs := NewRedisStore(client)

	cleanup := func() {
		client.Close()
		srv.Close()
	}
	return rs, srv, cleanup
}

func TestRedisStore_Ping(t *testing.T) {
	rs, _, cleanup := newTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	if err := rs.Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}

func TestRedisStore_SetGet(t *testing.T) {
	rs, _, cleanup := newTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := rs.set(ctx, "foo", "bar", 0); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	val, err := rs.get(ctx, "foo")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if val != "bar" {
		t.Fatalf("expected value 'bar', got %q", val)
	}
}

func TestRedisStore_SetTTLExpiration(t *testing.T) {
	rs, srv, cleanup := newTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := rs.set(ctx, "temp", "val", 100*time.Millisecond); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	//key must exist now
	if _, err := rs.get(ctx, "temp"); err != nil {
		t.Fatalf("Get() before expiry error = %v", err)
	}

	// advance miniredis clock past the set TTL
	srv.FastForward(200 * time.Millisecond)

	_, err := rs.get(ctx, "temp")
	if err == nil {
		t.Fatalf("expected error after TTL expiry, got nil")
	}
	if !errors.Is(err, redis.Nil) {
		t.Fatalf("expected redis.Nil after expiry, got %v", err)
	}
}

func TestRedisStore_GetMissingKey(t *testing.T) {
	rs, _, cleanup := newTestRedisStore(t)
	defer cleanup()

	ctx := context.Background()
	_, err := rs.get(ctx, "missing-key")
	if err == nil {
		t.Fatalf("expected error for missing key, got nil")
	}
	if !errors.Is(err, redis.Nil) {
		t.Fatalf("expected redis.Nil for missing key, got %v", err)
	}
}

func TestRedisStore_AddRoute(t *testing.T) {
	rs, srv, cleanup := newTestRedisStore(t)
	defer cleanup()
	ctx := context.Background()
	routeUuid := uuid.NewString()
	route := models.Route{
		Identifier:              routeUuid,
		Name:                    "test route",
		HttpMethod:              enums.Delete,
		VerificationType:        enums.None,
		VerificationKeyLocation: enums.NoLocation,
		VerificationKeyName:     "none",
		VerificationToken:       uuid.NewString(),
		Active:                  true,
	}

	identifier, err := rs.AddRoute(ctx, route)
	if err != nil {
		t.Fatalf("AddRoute() error = %v", err)
	}

	if identifier != routeUuid {
		t.Fatalf("expected route identifier '%s', got '%s'", routeUuid, identifier)
	}

	//check if redis has the set and hash
	isMember, _ := srv.SIsMember(RoutesKey, routeUuid)
	if !isMember {
		t.Fatalf("expected route '%s' to be a member of the '%s' set, not member", routeUuid, RoutesKey)
	}

	verificationToken := srv.HGet(fmt.Sprintf("%s%s", RouteKey, route.Identifier), "verification_token")
	if verificationToken != route.VerificationToken {
		t.Fatalf("expected verification token to be '%s', got '%s'", route.VerificationToken, verificationToken)
	}
}
