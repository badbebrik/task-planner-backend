package auth

import (
	"sync"
	"time"
)

type RateLimiter struct {
	attempts map[string][]time.Time
	mu       sync.RWMutex
	window   time.Duration
	limit    int
}

func NewRateLimiter(window time.Duration, limit int) *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string][]time.Time),
		window:   window,
		limit:    limit,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	if attempts, exists := rl.attempts[key]; exists {
		valid := attempts[:0]
		for _, t := range attempts {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		rl.attempts[key] = valid

		if len(valid) >= rl.limit {
			return false
		}
	}

	rl.attempts[key] = append(rl.attempts[key], now)
	return true
}

func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for key, attempts := range rl.attempts {
		valid := attempts[:0]
		for _, t := range attempts {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.attempts, key)
		} else {
			rl.attempts[key] = valid
		}
	}
}
