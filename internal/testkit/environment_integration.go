//go:build integration

package testkit

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"clinic-vet-api/database/sqlc"
	"clinic-vet-api/internal/core/notifications"
	api "clinic-vet-api/internal/infrastructure/http"
	"clinic-vet-api/internal/middleware"
	"clinic-vet-api/internal/shared/log"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

const jwtSecret = "integration-test-secret-that-is-long-enough"

type Environment struct {
	Router *gin.Engine
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Email  *Sender
	SMS    *Sender
}

type Sender struct {
	mu    sync.Mutex
	Sent  []*notifications.Notification
	label string
}

// AccessToken creates a role-scoped token for routes whose behavior does not
// require a persisted staff profile.
func AccessToken(t *testing.T, userID uint, role string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": fmt.Sprintf("%d", userID),
		"email":   "staff@example.com",
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}
	return signed
}

func (s *Sender) Send(_ context.Context, notification *notifications.Notification) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := *notification
	s.Sent = append(s.Sent, &copy)
	return s.label, nil
}

var (
	containersOnce sync.Once
	postgresURL    string
	redisAddress   string
	containersErr  error
)

// NewEnvironment creates an isolated database schema and a real, fully wired API router.
func NewEnvironment(t *testing.T) *Environment {
	t.Helper()
	startContainers(t)
	if containersErr != nil {
		t.Fatalf("start integration containers: %v", containersErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	schema := "e2e_" + uuid.NewString()[0:8]
	admin, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		t.Fatalf("connect PostgreSQL admin pool: %v", err)
	}
	if _, err := admin.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)); err != nil {
		admin.Close()
		t.Fatalf("create test schema: %v", err)
	}

	databaseURL := postgresURL + "&search_path=" + schema
	if err := migrateSchema(databaseURL); err != nil {
		admin.Close()
		t.Fatalf("migrate test schema: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(postgresURL)
	if err != nil {
		admin.Close()
		t.Fatalf("parse PostgreSQL pool config: %v", err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		admin.Close()
		t.Fatalf("connect test schema: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddress})
	if err := redisClient.FlushDB(ctx).Err(); err != nil {
		db.Close()
		admin.Close()
		t.Fatalf("clear Redis: %v", err)
	}

	email := &Sender{label: "email"}
	sms := &Sender{label: "sms"}
	router := gin.New()
	router.Use(gin.Recovery())
	if err := api.BootstrapWithConfig(&api.APIConfig{
		Router:                 router,
		Queries:                sqlc.New(db),
		Validator:              newValidator(),
		AuthMiddleware:         middleware.NewAuthMiddleware(jwtSecret),
		Redis:                  redisClient,
		JWTSecret:              jwtSecret,
		TwoFactorEncryptionKey: []byte("0123456789abcdef0123456789abcdef"),
		EmailSender:            email,
		SMSSender:              sms,
	}); err != nil {
		_ = redisClient.Close()
		db.Close()
		admin.Close()
		t.Fatalf("bootstrap router: %v", err)
	}

	t.Cleanup(func() {
		_ = redisClient.Close()
		db.Close()
		_, _ = admin.Exec(context.Background(), fmt.Sprintf(`DROP SCHEMA "%s" CASCADE`, schema))
		admin.Close()
	})
	return &Environment{Router: router, DB: db, Redis: redisClient, Email: email, SMS: sms}
}

func newValidator() *validator.Validate { return validator.New() }

func startContainers(t *testing.T) {
	t.Helper()
	containersOnce.Do(func() {
		gin.SetMode(gin.ReleaseMode)
		log.App = zap.NewNop()
		log.Audit = zap.NewNop()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		postgres, err := testcontainers.Run(ctx, "postgres:16-alpine",
			testcontainers.WithEnv(map[string]string{"POSTGRES_USER": "clinic", "POSTGRES_PASSWORD": "clinic", "POSTGRES_DB": "clinic"}),
			testcontainers.WithExposedPorts("5432/tcp"),
			testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
		)
		if err != nil {
			containersErr = err
			return
		}
		host, err := postgres.Host(ctx)
		if err != nil {
			containersErr = err
			return
		}
		port, err := postgres.MappedPort(ctx, "5432/tcp")
		if err != nil {
			containersErr = err
			return
		}
		postgresURL = fmt.Sprintf("postgres://clinic:clinic@%s:%s/clinic?sslmode=disable", host, port.Port())
		if err := waitForPostgres(ctx, postgresURL); err != nil {
			containersErr = err
			return
		}

		redisContainer, err := testcontainers.Run(ctx, "redis:7-alpine",
			testcontainers.WithExposedPorts("6379/tcp"),
			testcontainers.WithWaitStrategy(wait.ForListeningPort("6379/tcp")),
		)
		if err != nil {
			containersErr = err
			return
		}
		host, err = redisContainer.Host(ctx)
		if err != nil {
			containersErr = err
			return
		}
		port, err = redisContainer.MappedPort(ctx, "6379/tcp")
		if err != nil {
			containersErr = err
			return
		}
		redisAddress = host + ":" + port.Port()
	})
}

func migrateSchema(databaseURL string) error {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("locate migration directory")
	}
	migrations := filepath.Join(filepath.Dir(file), "..", "..", "database", "migrations")
	m, err := migrate.New("file://"+migrations, databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func waitForPostgres(ctx context.Context, databaseURL string) error {
	for {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err == nil {
			err = pool.Ping(ctx)
			pool.Close()
		}
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return fmt.Errorf("wait for PostgreSQL: %w", err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
