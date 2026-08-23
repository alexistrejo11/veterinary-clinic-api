package token

import (
	"context"
	"testing"

	"clinic-vet-api/internal/core/auth"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestTwoFactorChallengeIsOpaqueAndSingleUse(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	service, err := NewTokenService(Config{Redis: client, JWTSecret: "test-secret"})
	if err != nil {
		t.Fatal(err)
	}

	challenge, err := service.CreateToken(context.Background(), map[string]any{
		auth.ClaimUserID: "42",
	}, auth.TokenTypeTwoFactorChallenge)
	if err != nil {
		t.Fatal(err)
	}
	if challenge.Code == "" {
		t.Fatal("expected opaque challenge code")
	}
	valid, err := service.VerifyToken(context.Background(), challenge.Code, auth.TokenTypeTwoFactorChallenge)
	if err != nil || !valid {
		t.Fatalf("expected valid challenge, valid=%t err=%v", valid, err)
	}
	if err := service.RevokeToken(context.Background(), challenge.Code, auth.TokenTypeTwoFactorChallenge); err != nil {
		t.Fatal(err)
	}
	valid, err = service.VerifyToken(context.Background(), challenge.Code, auth.TokenTypeTwoFactorChallenge)
	if err != nil || valid {
		t.Fatalf("expected revoked challenge, valid=%t err=%v", valid, err)
	}
}
