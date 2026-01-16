package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type SafeCubeAPIRateLimiter struct {
	limiter *rate.Limiter
	mu      sync.RWMutex
}

func NewSafeCubeAPIRateLimiter() *SafeCubeAPIRateLimiter {
	limiter := rate.NewLimiter(rate.Every(1*time.Second), 1)

	return &SafeCubeAPIRateLimiter{
		limiter: limiter,
	}
}

func (rl *SafeCubeAPIRateLimiter) Wait(ctx context.Context) error {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	err := rl.limiter.Wait(ctx)
	if err != nil {
		return fmt.Errorf("rate limiter wait failed: %w", err)
	}
	return nil
}

func (rl *SafeCubeAPIRateLimiter) Allow() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limiter.Allow()
}

func (rl *SafeCubeAPIRateLimiter) Reserve() *rate.Reservation {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limiter.Reserve()
}

func (rl *SafeCubeAPIRateLimiter) SetBurst(burst int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiter.SetBurst(burst)
}

func (rl *SafeCubeAPIRateLimiter) SetLimit(newRate rate.Limit) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiter.SetLimit(newRate)
}

func (rl *SafeCubeAPIRateLimiter) TokensAvailable() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return int(rl.limiter.Tokens())
}
