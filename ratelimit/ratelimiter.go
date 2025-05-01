package ratelimit

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*TokenBucket
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*TokenBucket),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) GetBucket(clientID string) *TokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = NewTokenBucket(10, 1)
		rl.buckets[clientID] = bucket
	}
	return bucket
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for clientID, bucket := range rl.buckets {
		bucket.mu.Lock()
		idleTooLong := time.Since(bucket.lastAccess) > 24*time.Hour
		bucket.mu.Unlock()

		if idleTooLong {
			bucket.Stop()
			delete(rl.buckets, clientID)
		}
	}
}

func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for _, bucket := range rl.buckets {
		bucket.Stop()
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = NewTokenBucket(10, 1)
		rl.buckets[clientID] = bucket
	}
	rl.mu.Unlock()

	return bucket.TryTake()
}