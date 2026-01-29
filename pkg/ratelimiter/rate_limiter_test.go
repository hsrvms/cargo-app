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

	if rl.TokensAvailable() != 10 {
		t.Errorf("Expected 10 tokens available initially, got %d", rl.TokensAvailable())
	}
}

func TestSafeCubeAPIRateLimiter_Allow(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	for i := 0; i < 10; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should be allowed in burst, but was denied", i+1)
		}
	}

	if rl.Allow() {
		t.Error("11th request should be denied after burst exhausted")
	}
}

func TestSafeCubeAPIRateLimiter_Wait(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 10; i++ {
		err := rl.Wait(ctx)
		if err != nil {
			t.Errorf("Wait %d failed: %v", i+1, err)
		}
	}
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Errorf("Burst requests took too long: %v", elapsed)
	}

	start = time.Now()
	err := rl.Wait(ctx)
	elapsed = time.Since(start)

	if err != nil {
		t.Errorf("Wait after burst failed: %v", err)
	}

	if elapsed < 900*time.Millisecond || elapsed > 1100*time.Millisecond {
		t.Errorf("Expected wait time ~1s, got %v", elapsed)
	}
}

func TestSafeCubeAPIRateLimiter_WaitWithCancelledContext(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	for i := 0; i < 10; i++ {
		rl.Allow()
	}

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

	for i := 0; i < 10; i++ {
		if !rl.Allow() {
			break
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := rl.Wait(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
	}

	if elapsed < 80*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("Expected timeout around 100ms, got %v", elapsed)
	}
}

func TestSafeCubeAPIRateLimiter_Reserve(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	reservation := rl.Reserve()
	if reservation == nil {
		t.Fatal("Expected reservation, got nil")
	}

	if reservation.Delay() > 0 {
		t.Errorf("Expected no delay for first reservation, got %v", reservation.Delay())
	}

	for i := 0; i < 10; i++ {
		rl.Reserve()
	}

	reservation = rl.Reserve()
	delay := reservation.Delay()

	if delay <= 0 {
		t.Error("Expected positive delay after burst exhausted")
	}

	reservation.Cancel()
}

func TestSafeCubeAPIRateLimiter_SetBurst(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	rl.SetBurst(5)

	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("Request %d should be allowed with burst=5", i+1)
		}
	}

	if rl.Allow() {
		t.Error("6th request should be denied with burst=5")
	}
}

func TestSafeCubeAPIRateLimiter_SetLimit(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	rl.SetLimit(rate.Every(500 * time.Millisecond))

	for i := 0; i < 10; i++ {
		rl.Allow()
	}

	ctx := context.Background()
	start := time.Now()
	err := rl.Wait(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}

	if elapsed < 400*time.Millisecond || elapsed > 600*time.Millisecond {
		t.Errorf("Expected wait time ~500ms, got %v", elapsed)
	}
}

func TestSafeCubeAPIRateLimiter_TokensAvailable(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()

	if tokens := rl.TokensAvailable(); tokens != 10 {
		t.Errorf("Expected 10 tokens initially, got %d", tokens)
	}

	for i := 0; i < 3; i++ {
		rl.Allow()
	}

	if tokens := rl.TokensAvailable(); tokens != 7 {
		t.Errorf("Expected 7 tokens after using 3, got %d", tokens)
	}
}

func TestSafeCubeAPIRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()
	ctx := context.Background()

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

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

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

func TestSafeCubeAPIRateLimiter_SafeCubeCompliance(t *testing.T) {
	rl := NewSafeCubeAPIRateLimiter()
	ctx := context.Background()

	requestTimes := make([]time.Time, 15)

	start := time.Now()
	for i := 0; i < 15; i++ {
		err := rl.Wait(ctx)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		requestTimes[i] = time.Now()
	}
	totalDuration := time.Since(start)

	burstDuration := requestTimes[9].Sub(start)
	if burstDuration > 100*time.Millisecond {
		t.Errorf("Burst of 10 requests took too long: %v", burstDuration)
	}

	for i := 10; i < 15; i++ {
		expectedTime := start.Add(time.Duration(i-9) * time.Second)
		actualTime := requestTimes[i]

		tolerance := 200 * time.Millisecond
		if actualTime.Before(expectedTime.Add(-tolerance)) || actualTime.After(expectedTime.Add(tolerance)) {
			t.Errorf("Request %d timing off. Expected around %v, got %v (diff: %v)",
				i+1, expectedTime.Sub(start), actualTime.Sub(start), actualTime.Sub(expectedTime))
		}
	}

	expectedTotal := 5 * time.Second
	tolerance := 300 * time.Millisecond

	if totalDuration < expectedTotal-tolerance || totalDuration > expectedTotal+tolerance {
		t.Errorf("Total duration off. Expected ~%v, got %v", expectedTotal, totalDuration)
	}

	t.Logf("SafeCube compliance test passed: %d requests in %v", 15, totalDuration)
}
