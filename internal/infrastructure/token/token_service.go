package token

import (
	"clinic-vet-api/internal/core/auth"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	redisPrefixVerification       = "auth:verification:"
	redisPrefixPasswordReset      = "auth:password_reset:"
	redisPrefixTwoFactorChallenge = "auth:two_factor_challenge:"
	redisPrefixRefreshValid       = "auth:refresh:valid:"
	redisPrefixRefreshBlacklist   = "auth:refresh:blacklist:"

	accessTokenTTL        = 15 * time.Minute
	refreshTokenTTL       = 7 * 24 * time.Hour
	verificationTTL       = 24 * time.Hour
	passwordResetTTL      = 1 * time.Hour
	twoFactorChallengeTTL = 5 * time.Minute
)

// TokenServiceImpl implements auth.TokenService using JWT for access/refresh and Redis for verification/password-reset tokens.
type TokenServiceImpl struct {
	redis      *redis.Client
	jwtKey     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// Config configures TokenServiceImpl.
type Config struct {
	Redis      *redis.Client
	JWTSecret  string
	AccessTTL  time.Duration // default 15m
	RefreshTTL time.Duration // default 7d
}

// NewTokenService creates a new TokenServiceImpl.
func NewTokenService(cfg Config) (*TokenServiceImpl, error) {
	if cfg.Redis == nil {
		return nil, errors.New("redis client is required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("jwt secret is required")
	}
	at := cfg.AccessTTL
	if at == 0 {
		at = accessTokenTTL
	}
	rt := cfg.RefreshTTL
	if rt == 0 {
		rt = refreshTokenTTL
	}
	return &TokenServiceImpl{
		redis:      cfg.Redis,
		jwtKey:     []byte(cfg.JWTSecret),
		accessTTL:  at,
		refreshTTL: rt,
	}, nil
}

var _ auth.TokenService = (*TokenServiceImpl)(nil)

func (s *TokenServiceImpl) CreateToken(ctx context.Context, userClaims map[string]any, tokenType auth.TokenType) (auth.Token, error) {
	switch tokenType {
	case auth.TokenTypeAccessToken, auth.TokenJwtAccessToken:
		return s.createAccessToken(ctx, userClaims)
	case auth.TokenTypeRefreshToken, auth.TokenJwtRefreshToken:
		return s.createRefreshToken(ctx, userClaims)
	case auth.TokenTypeVerification:
		return s.createRedisToken(ctx, userClaims, redisPrefixVerification, verificationTTL, tokenType)
	case auth.TokenTypePasswordReset:
		return s.createRedisToken(ctx, userClaims, redisPrefixPasswordReset, passwordResetTTL, tokenType)
	case auth.TokenTypeTwoFactorChallenge:
		return s.createRedisToken(ctx, userClaims, redisPrefixTwoFactorChallenge, twoFactorChallengeTTL, tokenType)
	default:
		return auth.Token{}, fmt.Errorf("unsupported token type: %s", tokenType)
	}
}

func (s *TokenServiceImpl) createAccessToken(ctx context.Context, claims map[string]any) (auth.Token, error) {
	_ = ctx
	now := time.Now()
	exp := now.Add(s.accessTTL)
	jti := uuid.New().String()

	c := jwt.MapClaims{
		auth.ClaimUserID:    claims[auth.ClaimUserID],
		auth.ClaimEmail:     claims[auth.ClaimEmail],
		auth.ClaimRole:      claims[auth.ClaimRole],
		auth.ClaimTokenType: string(auth.TokenTypeAccessToken),
		auth.ClaimJTI:       jti,
		auth.ClaimIat:       now.Unix(),
		auth.ClaimExp:       exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(s.jwtKey)
	if err != nil {
		return auth.Token{}, err
	}

	return auth.Token{
		Code:      signed,
		Type:      auth.TokenTypeAccessToken,
		ExpiresAt: exp,
		CreatedAt: now,
	}, nil
}

func (s *TokenServiceImpl) createRefreshToken(ctx context.Context, claims map[string]any) (auth.Token, error) {
	_ = ctx
	now := time.Now()
	exp := now.Add(s.refreshTTL)
	jti := uuid.New().String()

	userID, _ := claims[auth.ClaimUserID].(string)
	if userID == "" {
		return auth.Token{}, errors.New("user_id is required for refresh token")
	}

	c := jwt.MapClaims{
		auth.ClaimUserID:    userID,
		auth.ClaimTokenType: string(auth.TokenTypeRefreshToken),
		auth.ClaimJTI:       jti,
		auth.ClaimIat:       now.Unix(),
		auth.ClaimExp:       exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(s.jwtKey)
	if err != nil {
		return auth.Token{}, err
	}

	return auth.Token{
		Code:      signed,
		Type:      auth.TokenTypeRefreshToken,
		ExpiresAt: exp,
		CreatedAt: now,
	}, nil
}

func (s *TokenServiceImpl) createRedisToken(ctx context.Context, claims map[string]any, prefix string, ttl time.Duration, tokenType auth.TokenType) (auth.Token, error) {
	code := uuid.New().String()
	now := time.Now()
	exp := now.Add(ttl)

	payload, err := json.Marshal(claims)
	if err != nil {
		return auth.Token{}, err
	}

	key := prefix + code
	if err := s.redis.Set(ctx, key, payload, ttl).Err(); err != nil {
		return auth.Token{}, err
	}

	return auth.Token{
		Code:      code,
		Type:      tokenType,
		ExpiresAt: exp,
		CreatedAt: now,
	}, nil
}

func (s *TokenServiceImpl) GetTokenClaims(ctx context.Context, tokenStr string, tokenType auth.TokenType) (map[string]any, error) {
	switch tokenType {
	case auth.TokenTypeAccessToken, auth.TokenJwtAccessToken, auth.TokenTypeRefreshToken, auth.TokenJwtRefreshToken:
		return s.getJWTClaims(tokenStr)
	case auth.TokenTypeVerification:
		return s.getRedisClaims(ctx, redisPrefixVerification+tokenStr)
	case auth.TokenTypePasswordReset:
		return s.getRedisClaims(ctx, redisPrefixPasswordReset+tokenStr)
	case auth.TokenTypeTwoFactorChallenge:
		return s.getRedisClaims(ctx, redisPrefixTwoFactorChallenge+tokenStr)
	default:
		return nil, fmt.Errorf("unsupported token type: %s", tokenType)
	}
}

func (s *TokenServiceImpl) getJWTClaims(tokenStr string) (map[string]any, error) {
	tok, err := jwt.Parse(tokenStr, func(*jwt.Token) (interface{}, error) { return s.jwtKey, nil })
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	out := make(map[string]any)
	for k, v := range claims {
		out[k] = v
	}
	return out, nil
}

func (s *TokenServiceImpl) getRedisClaims(ctx context.Context, key string) (map[string]any, error) {
	payload, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("token not found or expired")
		}
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *TokenServiceImpl) VerifyToken(ctx context.Context, tokenStr string, tokenType auth.TokenType) (bool, error) {
	switch tokenType {
	case auth.TokenTypeAccessToken, auth.TokenJwtAccessToken:
		_, err := s.getJWTClaims(tokenStr)
		return err == nil, nil
	case auth.TokenTypeRefreshToken, auth.TokenJwtRefreshToken:
		claims, err := s.getJWTClaims(tokenStr)
		if err != nil {
			return false, nil
		}
		jti, _ := claims[auth.ClaimJTI].(string)
		if jti == "" {
			return false, nil
		}
		_, err = s.redis.Get(ctx, redisPrefixRefreshBlacklist+jti).Result()
		if err == nil {
			return false, nil
		}
		if !errors.Is(err, redis.Nil) {
			return false, err
		}
		return true, nil
	case auth.TokenTypeVerification:
		_, err := s.getRedisClaims(ctx, redisPrefixVerification+tokenStr)
		return err == nil, nil
	case auth.TokenTypePasswordReset:
		_, err := s.getRedisClaims(ctx, redisPrefixPasswordReset+tokenStr)
		return err == nil, nil
	case auth.TokenTypeTwoFactorChallenge:
		_, err := s.getRedisClaims(ctx, redisPrefixTwoFactorChallenge+tokenStr)
		return err == nil, nil
	default:
		return false, fmt.Errorf("unsupported token type: %s", tokenType)
	}
}

func (s *TokenServiceImpl) RefreshAccessToken(ctx context.Context, refreshToken string) (auth.Token, error) {
	claims, err := s.getJWTClaims(refreshToken)
	if err != nil {
		return auth.Token{}, err
	}
	userID, _ := claims[auth.ClaimUserID].(string)
	if userID == "" {
		return auth.Token{}, errors.New("invalid refresh token claims")
	}
	return s.CreateToken(ctx, map[string]any{auth.ClaimUserID: userID}, auth.TokenTypeAccessToken)
}

func (s *TokenServiceImpl) RevokeToken(ctx context.Context, tokenStr string, tokenType auth.TokenType) error {
	switch tokenType {
	case auth.TokenTypeRefreshToken, auth.TokenJwtRefreshToken:
		claims, err := s.getJWTClaims(tokenStr)
		if err != nil {
			return err
		}
		jti, _ := claims[auth.ClaimJTI].(string)
		if jti == "" {
			return errors.New("refresh token has no jti")
		}
		key := redisPrefixRefreshBlacklist + jti
		return s.redis.Set(ctx, key, "1", s.refreshTTL).Err()
	case auth.TokenTypeVerification:
		return s.redis.Del(ctx, redisPrefixVerification+tokenStr).Err()
	case auth.TokenTypePasswordReset:
		return s.redis.Del(ctx, redisPrefixPasswordReset+tokenStr).Err()
	case auth.TokenTypeTwoFactorChallenge:
		return s.redis.Del(ctx, redisPrefixTwoFactorChallenge+tokenStr).Err()
	case auth.TokenTypeAccessToken, auth.TokenJwtAccessToken:
		return nil
	default:
		return nil
	}
}

func (s *TokenServiceImpl) SaveToken(ctx context.Context, t auth.Token) error {
	switch t.Type {
	case auth.TokenTypeRefreshToken, auth.TokenJwtRefreshToken:
		claims, err := s.getJWTClaims(t.Code)
		if err != nil {
			return err
		}
		jti, _ := claims[auth.ClaimJTI].(string)
		if jti == "" {
			return errors.New("refresh token has no jti")
		}
		ttl := time.Until(t.ExpiresAt)
		if ttl <= 0 {
			return nil
		}
		return s.redis.Set(ctx, redisPrefixRefreshValid+jti, "1", ttl).Err()
	case auth.TokenTypeVerification, auth.TokenTypePasswordReset:
		return nil
	default:
		return nil
	}
}
