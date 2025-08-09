package middleware

import (
	"context"
	"fmt"
	"go-corenglish/internal/commons/response"
	"go-corenglish/internal/config"
	"go-corenglish/pkg/token"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)

		statusCode := c.Writer.Status()

		entry := logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"status":     statusCode,
			"latency":    latency,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		})

		if statusCode >= 400 {
			entry.Error("HTTP request completed with error")
		} else {
			entry.Info("HTTP request completed")
		}
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.WithFields(logrus.Fields{
					"error":  err,
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
				}).Error("Panic recovered")

				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  false,
					"error":   "internal_server_error",
					"message": "An internal server error occurred",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtSecret string, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			resp := response.UnauthorizedErrorWithAdditionalInfo(nil, "Authorization header is required")
			c.AbortWithStatusJSON(resp.StatusCode, resp)
			return
		}

		bearerToken := strings.Split(authHeader, "Bearer ")

		if len(bearerToken) != 2 {
			resp := response.UnauthorizedErrorWithAdditionalInfo(nil, "len token must be 2")
			c.AbortWithStatusJSON(resp.StatusCode, resp)
			return
		}

		payload, err := token.ValidateToken(bearerToken[1])
		if err != nil {
			resp := response.UnauthorizedErrorWithAdditionalInfo(err.Error())
			c.AbortWithStatusJSON(resp.StatusCode, resp)
			return
		}

		userID, err := uuid.Parse(payload.AuthId)
		if err != nil {
			resp := response.UnauthorizedErrorWithAdditionalInfo(nil, "Invalid user ID in token")
			c.AbortWithStatusJSON(resp.StatusCode, resp)
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting using Redis
func RateLimitMiddleware(redisClient *redis.Client, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if redisClient == nil {
			c.Next()
			return
		}

		key := "rate_limit:" + c.ClientIP()
		ctx := context.Background()

		current, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		if current >= cfg.RateLimitRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  false,
				"error":   "rate_limit_exceeded",
				"message": fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %d seconds", cfg.RateLimitRequests, cfg.RateLimitWindow),
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Duration(cfg.RateLimitWindow)*time.Second)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		// Add rate limit headers
		remaining := cfg.RateLimitRequests - (current + 1)
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.RateLimitRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(cfg.RateLimitWindow)*time.Second).Unix(), 10))

		c.Next()
	}
}

// InMemoryRateLimitMiddleware implements in-memory rate limiting (fallback)
type InMemoryRateLimiter struct {
	limiters map[string]*rate.Limiter
}

func NewInMemoryRateLimiter() *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (rl *InMemoryRateLimiter) GetLimiter(key string, requests int, window time.Duration) *rate.Limiter {
	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(window/time.Duration(requests)), requests)
		rl.limiters[key] = limiter
	}
	return limiter
}

func InMemoryRateLimitMiddleware(rateLimiter *InMemoryRateLimiter, requests int, window int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		limiter := rateLimiter.GetLimiter(key, requests, time.Duration(window)*time.Second)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  false,
				"error":   "rate_limit_exceeded",
				"message": fmt.Sprintf("Rate limit exceeded. Maximum %d requests per %d seconds", requests, window),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
