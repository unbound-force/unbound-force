package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenManager_InitialAcquisition(t *testing.T) {
	tm := NewTokenManager(TokenManagerOpts{
		RefreshFn: func() (string, error) {
			return "initial-token", nil
		},
		Lifetime:        55 * time.Minute,
		ProactiveWindow: 5 * time.Minute,
		Interval:        50 * time.Minute,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tm.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer tm.Stop()

	token, err := tm.Token()
	if err != nil {
		t.Fatalf("Token failed: %v", err)
	}
	if token != "initial-token" {
		t.Errorf("token: got %q, want initial-token", token)
	}
}

func TestTokenManager_InitialAcquisitionFails(t *testing.T) {
	tm := NewTokenManager(TokenManagerOpts{
		RefreshFn: func() (string, error) {
			return "", fmt.Errorf("gcloud not authenticated")
		},
		Lifetime:        55 * time.Minute,
		ProactiveWindow: 5 * time.Minute,
		Interval:        50 * time.Minute,
	})

	err := tm.Start(context.Background())
	if err == nil {
		t.Fatal("expected error when initial acquisition fails")
	}
	if !strings.Contains(err.Error(), "gcloud not authenticated") {
		t.Errorf("expected gcloud error, got: %s", err.Error())
	}
}

func TestTokenManager_BackgroundRefresh(t *testing.T) {
	var callCount atomic.Int32
	tm := NewTokenManager(TokenManagerOpts{
		RefreshFn: func() (string, error) {
			n := callCount.Add(1)
			return fmt.Sprintf("token-%d", n), nil
		},
		Lifetime:        55 * time.Minute,
		ProactiveWindow: 5 * time.Minute,
		Interval:        1 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())

	if err := tm.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for at least one background refresh cycle.
	time.Sleep(100 * time.Millisecond)

	cancel()
	tm.Stop()

	if callCount.Load() < 2 {
		t.Errorf("expected at least 2 calls (initial + refresh), got: %d",
			callCount.Load())
	}

	token, err := tm.Token()
	if err != nil {
		t.Fatalf("Token failed: %v", err)
	}
	if token == "" {
		t.Error("expected token to be set after refresh")
	}
}

func TestTokenManager_BackgroundRefreshFailure_ClearsToken(t *testing.T) {
	// First call succeeds (initial), second call fails
	// (background refresh).
	var callCount atomic.Int32
	tm := NewTokenManager(TokenManagerOpts{
		RefreshFn: func() (string, error) {
			n := callCount.Add(1)
			if n == 1 {
				return "initial-token", nil
			}
			return "", fmt.Errorf("refresh failed")
		},
		Lifetime:        55 * time.Minute,
		ProactiveWindow: 5 * time.Minute,
		Interval:        1 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())

	if err := tm.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for the background refresh to fail.
	time.Sleep(100 * time.Millisecond)

	cancel()
	tm.Stop()

	// Token should be cleared after background refresh
	// failure (atomic invalidation per learning/gateway-3).
	_, err := tm.Token()
	if err == nil {
		t.Fatal("expected error after background refresh failure clears token")
	}
	if !strings.Contains(err.Error(), "unavailable") {
		t.Errorf("expected 'unavailable' error, got: %s", err.Error())
	}
}

func TestTokenManager_ProactiveRefresh(t *testing.T) {
	tm := NewTokenManager(TokenManagerOpts{
		RefreshFn: func() (string, error) {
			return "refreshed-token", nil
		},
		Lifetime:        55 * time.Minute,
		ProactiveWindow: 5 * time.Minute,
		// Long interval so background refresh doesn't
		// interfere with the proactive refresh test.
		Interval: 1 * time.Hour,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tm.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer tm.Stop()

	// Manually set token expiry within the proactive
	// window to trigger proactive refresh.
	tm.tokenMu.Lock()
	tm.token = "old-token"
	tm.tokenExpiry = time.Now().Add(3 * time.Minute)
	tm.tokenMu.Unlock()

	token, err := tm.Token()
	if err != nil {
		t.Fatalf("Token failed: %v", err)
	}
	if token != "refreshed-token" {
		t.Errorf("token: got %q, want refreshed-token", token)
	}
}

func TestTokenManager_ProactiveRefreshFails_PreservesToken(t *testing.T) {
	// First call succeeds (initial), subsequent calls fail.
	var callCount atomic.Int32
	tm := NewTokenManager(TokenManagerOpts{
		RefreshFn: func() (string, error) {
			n := callCount.Add(1)
			if n == 1 {
				return "initial-token", nil
			}
			return "", fmt.Errorf("gcloud not found")
		},
		Lifetime:        55 * time.Minute,
		ProactiveWindow: 5 * time.Minute,
		Interval:        1 * time.Hour,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tm.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer tm.Stop()

	// Set token expiry within the proactive window.
	tm.tokenMu.Lock()
	tm.tokenExpiry = time.Now().Add(3 * time.Minute)
	tm.tokenMu.Unlock()

	// Proactive refresh fails, but the existing token
	// should be preserved (not cleared).
	token, err := tm.Token()
	if err != nil {
		t.Fatalf("Token should succeed with still-valid token: %v", err)
	}
	if token != "initial-token" {
		t.Errorf("token should be preserved on proactive refresh failure, got: %q", token)
	}
}

func TestTokenManager_ContextCancellation(t *testing.T) {
	var callCount atomic.Int32
	tm := NewTokenManager(TokenManagerOpts{
		RefreshFn: func() (string, error) {
			callCount.Add(1)
			return "token", nil
		},
		Lifetime:        55 * time.Minute,
		ProactiveWindow: 5 * time.Minute,
		Interval:        1 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())

	if err := tm.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Let it run a few cycles.
	time.Sleep(50 * time.Millisecond)

	// Cancel should stop the background refresh.
	cancel()
	tm.Stop()

	countAtStop := callCount.Load()
	time.Sleep(50 * time.Millisecond)
	countAfterStop := callCount.Load()

	// No new calls should happen after cancellation.
	if countAfterStop > countAtStop+1 {
		t.Errorf("refresh continued after cancellation: %d → %d",
			countAtStop, countAfterStop)
	}
}

func TestTokenManager_ConcurrentAccess(t *testing.T) {
	// Verify that concurrent Token() calls with near-expiry
	// tokens result in exactly 1 proactive refresh
	// invocation (TryLock deduplication).
	var refreshCount atomic.Int32
	tm := NewTokenManager(TokenManagerOpts{
		RefreshFn: func() (string, error) {
			refreshCount.Add(1)
			// Brief sleep to simulate real gcloud latency.
			time.Sleep(10 * time.Millisecond)
			return "refreshed-token", nil
		},
		Lifetime:        55 * time.Minute,
		ProactiveWindow: 5 * time.Minute,
		Interval:        1 * time.Hour,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tm.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer tm.Stop()

	// Reset refresh count after initial acquisition.
	refreshCount.Store(0)

	// Set token expiry within the proactive window.
	tm.tokenMu.Lock()
	tm.token = "near-expiry-token"
	tm.tokenExpiry = time.Now().Add(3 * time.Minute)
	tm.tokenMu.Unlock()

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Use a barrier to synchronize all goroutines.
	barrier := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-barrier // Wait for barrier release.
			_, _ = tm.Token()
		}()
	}

	// Release all goroutines simultaneously.
	close(barrier)
	wg.Wait()

	count := refreshCount.Load()
	if count != 1 {
		t.Errorf("expected exactly 1 refresh invocation "+
			"(TryLock deduplication), got: %d", count)
	}
}

func TestTokenManager_EmptyToken(t *testing.T) {
	tm := &TokenManager{
		refreshFn:       func() (string, error) { return "", nil },
		lifetime:        55 * time.Minute,
		proactiveWindow: 5 * time.Minute,
		interval:        50 * time.Minute,
	}
	// Token is empty (never started).
	_, err := tm.Token()
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	if !strings.Contains(err.Error(), "unavailable") {
		t.Errorf("expected 'unavailable' error, got: %s", err.Error())
	}
}

func TestTokenManager_ExpiredToken(t *testing.T) {
	tm := &TokenManager{
		token:           "expired-token",
		tokenExpiry:     time.Now().Add(-10 * time.Minute),
		refreshFn:       func() (string, error) { return "", fmt.Errorf("fail") },
		lifetime:        55 * time.Minute,
		proactiveWindow: 5 * time.Minute,
		interval:        50 * time.Minute,
	}
	_, err := tm.Token()
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected 'expired' error, got: %s", err.Error())
	}
}

func TestTokenManager_StopIdempotent(t *testing.T) {
	tm := NewTokenManager(TokenManagerOpts{
		RefreshFn: func() (string, error) { return "t", nil },
		Lifetime:  55 * time.Minute,
		Interval:  50 * time.Minute,
	})
	// Stop without Start — should not panic.
	tm.Stop()
	tm.Stop()
}
