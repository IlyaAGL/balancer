package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity    int
	tokens      int
	refillRate  int
	mu          sync.Mutex
	stopChan    chan struct{}
	lastAccess  time.Time
}

func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	tb := &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		stopChan:   make(chan struct{}),
		lastAccess: time.Now(),
	}
	go tb.startRefill()
	return tb
}

func (tb *TokenBucket) startRefill() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tb.mu.Lock()
			tb.tokens += tb.refillRate
			if tb.tokens > tb.capacity {
				tb.tokens = tb.capacity
			}
			tb.mu.Unlock()
		case <-tb.stopChan:
			return
		}
	}
}

func (tb *TokenBucket) TryTake() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.lastAccess = time.Now()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) Stop() {
	close(tb.stopChan)
}