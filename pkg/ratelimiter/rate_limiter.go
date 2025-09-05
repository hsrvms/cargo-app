package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// SafeCubeAPIRateLimiter handles rate limiting for SafeCube API requests
// SafeCube API allows 10 requests per 10 seconds per API key
type SafeCubeAPIRateLimiter struct {
	limiter *rate.Limiter
	mu      sync.RWMutex
}

// NewSafeCubeAPIRateLimiter creates a new rate limiter for SafeCube API
// SafeCube API rate limit: 10 requests/10 seconds = 1 request per second on average
func NewSafeCubeAPIRateLimiter() *SafeCubeAPIRateLimiter {
	// Allow burst of 10 requests, then 1 request per second
	// This matches SafeCube's "10 requests per 10 seconds" limit
	limiter := rate.NewLimiter(rate.Every(1*time.Second), 10)

	return &SafeCubeAPIRateLimiter{
		limiter: limiter,
	}
}

// Wait waits until the rate limiter allows the request to proceed
// It returns an error if the context is cancelled
func (rl *SafeCubeAPIRateLimiter) Wait(ctx context.Context) error {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	err := rl.limiter.Wait(ctx)
	if err != nil {
		return fmt.Errorf("rate limiter wait failed: %w", err)
	}
	return nil
}

// Allow checks if a request is allowed without blocking
// Returns true if the request can proceed immediately
func (rl *SafeCubeAPIRateLimiter) Allow() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limiter.Allow()
}

// Reserve reserves a token for future use and returns a Reservation
// The reservation can be cancelled if needed
func (rl *SafeCubeAPIRateLimiter) Reserve() *rate.Reservation {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limiter.Reserve()
}

// SetBurst allows changing the burst size at runtime
func (rl *SafeCubeAPIRateLimiter) SetBurst(burst int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiter.SetBurst(burst)
}

// SetLimit allows changing the rate limit at runtime
func (rl *SafeCubeAPIRateLimiter) SetLimit(newRate rate.Limit) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiter.SetLimit(newRate)
}

// TokensAvailable returns the number of tokens available at this moment
func (rl *SafeCubeAPIRateLimiter) TokensAvailable() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return int(rl.limiter.Tokens())
}
