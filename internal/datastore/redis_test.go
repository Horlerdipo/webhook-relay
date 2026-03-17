package datastore

import (
	"context"
	"errors"
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
	if _, err := rs.Set(ctx, "foo", "bar", 0); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	val, err := rs.Get(ctx, "foo")
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
	if _, err := rs.Set(ctx, "temp", "val", 100*time.Millisecond); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	//key must exist now
	if _, err := rs.Get(ctx, "temp"); err != nil {
		t.Fatalf("Get() before expiry error = %v", err)
	}

	// advance miniredis clock past the set TTL
	srv.FastForward(200 * time.Millisecond)

	_, err := rs.Get(ctx, "temp")
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
	_, err := rs.Get(ctx, "missing-key")
	if err == nil {
		t.Fatalf("expected error for missing key, got nil")
	}
	if !errors.Is(err, redis.Nil) {
		t.Fatalf("expected redis.Nil for missing key, got %v", err)
	}
}
