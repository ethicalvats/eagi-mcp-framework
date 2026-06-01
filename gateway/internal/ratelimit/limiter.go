package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity  float64
	tokens    float64
	fillRate  float64 // tokens per second
	lastToken time.Time
	mu        sync.Mutex
}

type Limiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}

func NewLimiter() *Limiter {
	return &Limiter{
		buckets: make(map[string]*TokenBucket),
	}
}

func (l *Limiter) Allow(userID string, role string) bool {
	l.mu.RLock()
	bucket, exists := l.buckets[userID]
	l.mu.RUnlock()

	if !exists {
		// Default limits based on role (could be configurable)
		var capacity, fillRate float64
		switch role {
		case "admin":
			capacity, fillRate = 100, 10
		case "developer":
			capacity, fillRate = 50, 5
		default:
			capacity, fillRate = 10, 1 // Strict for viewers
		}

		bucket = &TokenBucket{
			capacity:  capacity,
			tokens:    capacity,
			fillRate:  fillRate,
			lastToken: time.Now(),
		}

		l.mu.Lock()
		l.buckets[userID] = bucket
		l.mu.Unlock()
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastToken).Seconds()
	bucket.tokens += elapsed * bucket.fillRate
	if bucket.tokens > bucket.capacity {
		bucket.tokens = bucket.capacity
	}
	bucket.lastToken = now

	if bucket.tokens >= 1.0 {
		bucket.tokens--
		return true
	}

	return false
}

// CheckLimitMiddleware could be added to HTTP handlers
func (l *Limiter) CheckLimit(userID string, role string) error {
	if !l.Allow(userID, role) {
		return fmt.Errorf("rate limit exceeded for user %s", userID)
	}
	return nil
}
