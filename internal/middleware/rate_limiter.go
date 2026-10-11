package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiterConfig holds the configuration for the rate limiter
type RateLimiterConfig struct {
	// Maximum requests allowed in the window duration
	MaxRequests int
	// Window duration for sliding window algorithm
	WindowDuration time.Duration
	// Key generator function to identify clients (IP, user ID, etc.)
	KeyGenerator func(*gin.Context) string
	// Skip function - if returns true, rate limiting is skipped for this request
	Skip func(*gin.Context) bool
	// Custom error handler
	ErrorHandler func(*gin.Context, RateLimitInfo)
	// Storage backend (memory, redis, etc.)
	Storage RateLimitStorage
}

// RateLimitInfo contains information about the rate limit status
type RateLimitInfo struct {
	Remaining int
	ResetTime time.Time
	Limit     int
}

// RateLimitStorage interface for different storage backends
type RateLimitStorage interface {
	Get(key string) (*ClientInfo, bool)
	Set(key string, info *ClientInfo)
	Delete(key string)
	Cleanup() // Clean expired entries
}

// ClientInfo stores rate limiting information for a client
type ClientInfo struct {
	Requests  []time.Time
	LastReset time.Time
	mu        sync.RWMutex
}

// MemoryStorage implements in-memory storage for rate limiting
type MemoryStorage struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
}

// NewMemoryStorage creates a new memory storage instance
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		clients: make(map[string]*ClientInfo),
	}

	// Start cleanup goroutine
	go storage.cleanupRoutine()

	return storage
}

func (m *MemoryStorage) Get(key string) (*ClientInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, exists := m.clients[key]
	return client, exists
}

func (m *MemoryStorage) Set(key string, info *ClientInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[key] = info
}

func (m *MemoryStorage) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, key)
}

func (m *MemoryStorage) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, client := range m.clients {
		client.mu.Lock()
		if now.Sub(client.LastReset) > time.Hour {
			delete(m.clients, key)
		}
		client.mu.Unlock()
	}
}

func (m *MemoryStorage) cleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.Cleanup()
	}
}

// DefaultConfig returns a default rate limiter configuration
func DefaultConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxRequests:    15, // 15 requests per minute
		WindowDuration: time.Minute,
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
		Skip: func(c *gin.Context) bool {
			return false
		},
		ErrorHandler: func(c *gin.Context, info RateLimitInfo) {
			c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(time.Until(info.ResetTime).Seconds()), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d requests. Retry after: %s", info.Limit, time.Until(info.ResetTime).String()),
				"retry_after": time.Until(info.ResetTime).Seconds(),
			})
		},
		Storage: NewMemoryStorage(),
	}
}

// RateLimiter creates a new rate limiting middleware
func RateLimiter(config ...RateLimiterConfig) gin.HandlerFunc {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
		// Apply defaults for unset fields
		if cfg.Storage == nil {
			cfg.Storage = NewMemoryStorage()
		}
		if cfg.KeyGenerator == nil {
			cfg.KeyGenerator = DefaultConfig().KeyGenerator
		}
		if cfg.ErrorHandler == nil {
			cfg.ErrorHandler = DefaultConfig().ErrorHandler
		}
		if cfg.WindowDuration == 0 {
			cfg.WindowDuration = time.Minute
		}
		if cfg.MaxRequests == 0 {
			cfg.MaxRequests = 10
		}
	}

	return func(c *gin.Context) {
		// Skip rate limiting if specified
		if cfg.Skip != nil && cfg.Skip(c) {
			c.Next()
			return
		}

		key := cfg.KeyGenerator(c)
		now := time.Now()

		// Get or create client info
		clientInfo, exists := cfg.Storage.Get(key)
		if !exists {
			clientInfo = &ClientInfo{
				Requests:  make([]time.Time, 0),
				LastReset: now,
			}
			cfg.Storage.Set(key, clientInfo)
		}

		clientInfo.mu.Lock()
		defer clientInfo.mu.Unlock()

		// Clean old requests (sliding window)
		cutoff := now.Add(-cfg.WindowDuration)
		validRequests := make([]time.Time, 0)
		for _, reqTime := range clientInfo.Requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		clientInfo.Requests = validRequests

		// Check if limit exceeded
		requestCount := len(clientInfo.Requests)
		limit := cfg.MaxRequests

		if requestCount >= limit {
			// Rate limit exceeded
			resetTime := clientInfo.Requests[0].Add(cfg.WindowDuration)
			rateLimitInfo := RateLimitInfo{
				Remaining: 0,
				ResetTime: resetTime,
				Limit:     limit,
			}

			cfg.ErrorHandler(c, rateLimitInfo)
			c.Abort()
			return
		}

		// Add current request
		clientInfo.Requests = append(clientInfo.Requests, now)
		clientInfo.LastReset = now

		// Set rate limit headers
		remaining := limit - len(clientInfo.Requests)
		var resetTime time.Time
		if len(clientInfo.Requests) > 0 {
			resetTime = clientInfo.Requests[0].Add(cfg.WindowDuration)
		} else {
			resetTime = now.Add(cfg.WindowDuration)
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		c.Next()
	}
}
