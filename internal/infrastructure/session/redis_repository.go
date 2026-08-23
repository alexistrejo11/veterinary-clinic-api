package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"clinic-vet-api/internal/core/auth"

	"github.com/redis/go-redis/v9"
)

const (
	tokenKeyPrefix = "auth:session:token:"
	userKeyPrefix  = "auth:session:user:"
)

// RedisRepository stores refresh-token sessions without placing raw tokens in Redis keys.
type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) (*RedisRepository, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	return &RedisRepository{client: client}, nil
}

var _ auth.SessionRepository = (*RedisRepository)(nil)

func (r *RedisRepository) Create(ctx context.Context, session *auth.JwtSession) error {
	if session == nil || session.UserID == "" || session.RefreshToken == "" {
		return errors.New("session user ID and refresh token are required")
	}

	userID, err := strconv.ParseUint(session.UserID, 10, 0)
	if err != nil {
		return fmt.Errorf("invalid session user ID: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return errors.New("session expiration must be in the future")
	}

	payload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	tokenID := tokenID(session.RefreshToken)
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, tokenKey(tokenID), payload, ttl)
	pipe.SAdd(ctx, userKey(uint(userID)), tokenID)
	pipe.Expire(ctx, userKey(uint(userID)), ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (auth.JwtSession, error) {
	if refreshToken == "" {
		return auth.JwtSession{}, errors.New("refresh token is required")
	}

	payload, err := r.client.Get(ctx, tokenKey(tokenID(refreshToken))).Bytes()
	if errors.Is(err, redis.Nil) {
		return auth.JwtSession{}, errors.New("session not found")
	}
	if err != nil {
		return auth.JwtSession{}, err
	}

	var session auth.JwtSession
	if err := json.Unmarshal(payload, &session); err != nil {
		return auth.JwtSession{}, fmt.Errorf("decode session: %w", err)
	}
	return session, nil
}

func (r *RedisRepository) GetByUserAndRefreshToken(ctx context.Context, userID uint, refreshToken string) (auth.JwtSession, error) {
	session, err := r.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return auth.JwtSession{}, err
	}
	if session.UserID != strconv.FormatUint(uint64(userID), 10) {
		return auth.JwtSession{}, errors.New("session does not belong to user")
	}
	return session, nil
}

func (r *RedisRepository) GetByUserID(ctx context.Context, userID uint) ([]auth.JwtSession, error) {
	tokenIDs, err := r.client.SMembers(ctx, userKey(userID)).Result()
	if err != nil {
		return nil, err
	}

	sessions := make([]auth.JwtSession, 0, len(tokenIDs))
	staleIDs := make([]interface{}, 0)
	for _, id := range tokenIDs {
		payload, err := r.client.Get(ctx, tokenKey(id)).Bytes()
		if errors.Is(err, redis.Nil) {
			staleIDs = append(staleIDs, id)
			continue
		}
		if err != nil {
			return nil, err
		}

		var session auth.JwtSession
		if err := json.Unmarshal(payload, &session); err != nil {
			return nil, fmt.Errorf("decode session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if len(staleIDs) > 0 {
		if err := r.client.SRem(ctx, userKey(userID), staleIDs...).Err(); err != nil {
			return nil, err
		}
	}
	return sessions, nil
}

func (r *RedisRepository) DeleteUserSession(ctx context.Context, userID uint, refreshToken string) error {
	if refreshToken == "" {
		return errors.New("refresh token is required")
	}

	tokenID := tokenID(refreshToken)
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, tokenKey(tokenID))
	pipe.SRem(ctx, userKey(userID), tokenID)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisRepository) DeleteAllUserSessions(ctx context.Context, userID uint) error {
	tokenIDs, err := r.client.SMembers(ctx, userKey(userID)).Result()
	if err != nil {
		return err
	}

	pipe := r.client.TxPipeline()
	for _, id := range tokenIDs {
		pipe.Del(ctx, tokenKey(id))
	}
	pipe.Del(ctx, userKey(userID))
	_, err = pipe.Exec(ctx)
	return err
}

func tokenID(refreshToken string) string {
	hash := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(hash[:])
}

func tokenKey(id string) string {
	return tokenKeyPrefix + id
}

func userKey(userID uint) string {
	return userKeyPrefix + strconv.FormatUint(uint64(userID), 10)
}
