package session

import (
	"context"
	"testing"
	"time"

	"clinic-vet-api/internal/core/auth"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRepositorySessionLifecycle(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	repository, err := NewRedisRepository(client)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	session := &auth.JwtSession{
		UserID:       "42",
		RefreshToken: "refresh-token",
		DeviceInfo:   "browser",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
	}
	if err := repository.Create(context.Background(), session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	stored, err := repository.GetByUserAndRefreshToken(context.Background(), 42, session.RefreshToken)
	if err != nil {
		t.Fatalf("get user session: %v", err)
	}
	if stored.RefreshToken != session.RefreshToken || stored.UserID != session.UserID {
		t.Fatalf("stored session does not match created session: %#v", stored)
	}

	sessions, err := repository.GetByUserID(context.Background(), 42)
	if err != nil {
		t.Fatalf("list user sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected one user session, got %d", len(sessions))
	}

	if err := repository.DeleteUserSession(context.Background(), 42, session.RefreshToken); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	if _, err := repository.GetByRefreshToken(context.Background(), session.RefreshToken); err == nil {
		t.Fatal("expected deleted session lookup to fail")
	}
}

func TestRedisRepositoryRejectsInvalidSession(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	repository, err := NewRedisRepository(client)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}

	if err := repository.Create(context.Background(), &auth.JwtSession{UserID: "not-a-number", RefreshToken: "token", ExpiresAt: time.Now().Add(time.Hour)}); err == nil {
		t.Fatal("expected invalid user ID to be rejected")
	}
}
