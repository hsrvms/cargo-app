package ratelimiter

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestNewSafeCubeAPIRateLimiter(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	if rl == nil {
		t.Fatal("Expected rate limiter to be created, got nil")
	}

	// Check that it starts with full burst capacity
	if rl.TokensAvailable() != 10 {
		t.Errorf("Expected 10 tokens available initially, got %d", rl.TokensAvailable())
	}
}

func TestSafeCubeAPIRateLimiter_Allow(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	// Should allow up to 10 requests immediately (burst)
	for i := 0; i < 10; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should be allowed in burst, but was denied", i+1)
		}
	}

	// 11th request should be denied (no tokens left)
	if rl.Allow() {
		t.Error("11th request should be denied after burst exhausted")
	}
}

func TestSafeCubeAPIRateLimiter_Wait(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()
	ctx := context.Background()

	// First 10 requests should not wait (burst)
	start := time.Now()
	for i := 0; i < 10; i++ {
		err := rl.Wait(ctx)
		if err != nil {
			t.Errorf("Wait %d failed: %v", i+1, err)
		}
	}
	elapsed := time.Since(start)

	// Should complete very quickly (within 100ms) due to burst
	if elapsed > 100*time.Millisecond {
		t.Errorf("Burst requests took too long: %v", elapsed)
	}

	// Next request should wait approximately 1 second
	start = time.Now()
	err := rl.Wait(ctx)
	elapsed = time.Since(start)

	if err != nil {
		t.Errorf("Wait after burst failed: %v", err)
	}

	// Should wait close to 1 second (allow some tolerance)
	if elapsed < 900*time.Millisecond || elapsed > 1100*time.Millisecond {
		t.Errorf("Expected wait time ~1s, got %v", elapsed)
	}
}

func TestSafeCubeAPIRateLimiter_WaitWithCancelledContext(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	// Exhaust the burst
	for i := 0; i < 10; i++ {
		rl.Allow()
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rl.Wait(ctx)
	if err == nil {
		t.Error("Expected error when waiting with cancelled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestSafeCubeAPIRateLimiter_WaitWithTimeout(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	// Exhaust the burst
	for i := 0; i < 10; i++ {
		if !rl.Allow() {
			break
		}
	}

	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := rl.Wait(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error")
	}

	// Should timeout - allow some tolerance but should be close to 100ms
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
	}

	if elapsed < 80*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("Expected timeout around 100ms, got %v", elapsed)
	}
}

func TestSafeCubeAPIRateLimiter_Reserve(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	// Reserve a token
	reservation := rl.Reserve()
	if reservation == nil {
		t.Fatal("Expected reservation, got nil")
	}

	// Should not need to wait for burst
	if reservation.Delay() > 0 {
		t.Errorf("Expected no delay for first reservation, got %v", reservation.Delay())
	}

	// Reserve 10 more tokens (exhaust burst + 1)
	for i := 0; i < 10; i++ {
		rl.Reserve()
	}

	// Next reservation should have a delay
	reservation = rl.Reserve()
	delay := reservation.Delay()

	if delay <= 0 {
		t.Error("Expected positive delay after burst exhausted")
	}

	// Cancel the reservation
	reservation.Cancel()
}

func TestSafeCubeAPIRateLimiter_SetBurst(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	// Change burst size
	rl.SetBurst(5)

	// Should allow 5 immediate requests
	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should be allowed with burst=5", i+1)
		}
	}

	// 6th request should be denied
	if rl.Allow() {
		t.Error("6th request should be denied with burst=5")
	}
}

func TestSafeCubeAPIRateLimiter_SetLimit(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	// Set a higher rate (2 requests per second)
	rl.SetLimit(rate.Every(500 * time.Millisecond))

	// Exhaust burst
	for i := 0; i < 10; i++ {
		rl.Allow()
	}

	// Next request should wait ~500ms instead of 1s
	ctx := context.Background()
	start := time.Now()
	err := rl.Wait(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}

	// Should wait close to 500ms
	if elapsed < 400*time.Millisecond || elapsed > 600*time.Millisecond {
		t.Errorf("Expected wait time ~500ms, got %v", elapsed)
	}
}

func TestSafeCubeAPIRateLimiter_TokensAvailable(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	// Should start with 10 tokens
	if tokens := rl.TokensAvailable(); tokens != 10 {
		t.Errorf("Expected 10 tokens initially, got %d", tokens)
	}

	// Use 3 tokens
	for i := 0; i < 3; i++ {
		rl.Allow()
	}

	// Should have 7 tokens left
	if tokens := rl.TokensAvailable(); tokens != 7 {
		t.Errorf("Expected 7 tokens after using 3, got %d", tokens)
	}
}

func TestSafeCubeAPIRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()
	ctx := context.Background()

	// Test concurrent access doesn't cause race conditions
	const numGoroutines = 20
	const requestsPerGoroutine = 5

	errors := make(chan error, numGoroutines*requestsPerGoroutine)
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < requestsPerGoroutine; j++ {
				err := rl.Wait(ctx)
				if err != nil {
					errors <- err
					return
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func BenchmarkSafeCubeAPIRateLimiter_Allow(b *testing.B) {
	rl := NewSafeCubeAPIRateLimiter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

func BenchmarkSafeCubeAPIRateLimiter_Wait(b *testing.B) {
	rl := NewSafeCubeAPIRateLimiter()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Wait(ctx)
	}
}

// TestSafeCubeAPIRateLimiter_SafeCubeCompliance verifies that the rate limiter
// properly implements SafeCube's "10 requests per 10 seconds" rate limit
func TestSafeCubeAPIRateLimiter_SafeCubeCompliance(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()
	ctx := context.Background()

	// Record when we make each request
	requestTimes := make([]time.Time, 15) // Test 15 requests

	start := time.Now()
	for i := 0; i < 15; i++ {
		err := rl.Wait(ctx)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		requestTimes[i] = time.Now()
	}
	totalDuration := time.Since(start)

	// First 10 requests should be immediate (burst)
	burstDuration := requestTimes[9].Sub(start)
	if burstDuration > 100*time.Millisecond {
		t.Errorf("Burst of 10 requests took too long: %v", burstDuration)
	}

	// Requests 11-15 should be spaced ~1 second apart
	for i := 10; i < 15; i++ {
		expectedTime := start.Add(time.Duration(i-9) * time.Second)
		actualTime := requestTimes[i]

		// Allow 200ms tolerance
		tolerance := 200 * time.Millisecond
		if actualTime.Before(expectedTime.Add(-tolerance)) || actualTime.After(expectedTime.Add(tolerance)) {
			t.Errorf("Request %d timing off. Expected around %v, got %v (diff: %v)",
				i+1, expectedTime.Sub(start), actualTime.Sub(start), actualTime.Sub(expectedTime))
		}
	}

	// Total time should be around 5 seconds (burst immediate + 5 * 1 second waits)
	expectedTotal := 5 * time.Second
	tolerance := 300 * time.Millisecond

	if totalDuration < expectedTotal-tolerance || totalDuration > expectedTotal+tolerance {
		t.Errorf("Total duration off. Expected ~%v, got %v", expectedTotal, totalDuration)
	}

	t.Logf("SafeCube compliance test passed: %d requests in %v", 15, totalDuration)
}
