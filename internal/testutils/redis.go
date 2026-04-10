package testutils

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/horlerdipo/webhook-relay/internal/datastore"
	"github.com/redis/go-redis/v9"
	"testing"
)

func NewTestRedisStore(t *testing.T) (*datastore.RedisStore, *miniredis.Miniredis, func()) {
	t.Helper()
	srv := miniredis.RunT(t)

	opt := &redis.Options{Addr: srv.Addr()}
	client := redis.NewClient(opt)
	rs := datastore.NewRedisStore(client)

	cleanup := func() {
		client.Close()
		srv.Close()
	}
	return rs, srv, cleanup
}
