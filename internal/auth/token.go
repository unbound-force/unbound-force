package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

// proactiveRefreshTimeout is the maximum time to wait
// for a proactive refresh subprocess. Prevents a hung
// gcloud or aws CLI from blocking HTTP requests.
const proactiveRefreshTimeout = 5 * time.Second

// TokenManagerOpts configures a TokenManager.
//
// Design decision: Options struct per SOLID Single
// Responsibility — the constructor receives all
// configuration in one struct, keeping the TokenManager
// fields focused on runtime state (per Cobalt-Crush
// engineering philosophy).
type TokenManagerOpts struct {
	// RefreshFn acquires a fresh token/credential string.
	// Called for initial acquisition and periodic refresh.
	RefreshFn func() (string, error)

	// Lifetime is the assumed validity duration of a
	// token. The expiry is set to now + Lifetime after
	// each successful refresh.
	Lifetime time.Duration

	// ProactiveWindow is the duration before expiry
	// during which Token() will attempt a non-blocking
	// proactive refresh.
	ProactiveWindow time.Duration

	// Interval is the background refresh ticker interval.
	// Replaces the mutable refreshMinute package variable
	// from the gateway (per design.md D2).
	Interval time.Duration
}

// TokenManager encapsulates the full token lifecycle:
// storage with sync.RWMutex, expiry tracking, background
// refresh loop, proactive refresh with TryLock() dedup,
// and atomic token invalidation on failure.
//
// Extracted from VertexProvider/BedrockProvider per
// design.md D2 to share defensive patterns between
// the gateway and ollama-proxy. The stale-token bug
// (learning/gateway-3) motivated this extraction —
// both services inherit all defensive patterns without
// re-implementation risk.
type TokenManager struct {
	token       string
	tokenMu     sync.RWMutex
	tokenExpiry time.Time
	proactiveMu sync.Mutex

	refreshFn       func() (string, error)
	lifetime        time.Duration
	proactiveWindow time.Duration
	interval        time.Duration
	cancel          context.CancelFunc
}

// NewTokenManager creates a TokenManager with the given
// options. Call Start() to acquire the initial token and
// launch the background refresh goroutine.
func NewTokenManager(opts TokenManagerOpts) *TokenManager {
	return &TokenManager{
		refreshFn:       opts.RefreshFn,
		lifetime:        opts.Lifetime,
		proactiveWindow: opts.ProactiveWindow,
		interval:        opts.Interval,
	}
}

// Start acquires the initial token via refreshFn and
// launches the background refresh goroutine. Returns an
// error if the initial token acquisition fails.
func (tm *TokenManager) Start(ctx context.Context) error {
	token, err := tm.refreshFn()
	if err != nil {
		return err
	}
	tm.tokenMu.Lock()
	tm.token = token
	tm.tokenExpiry = time.Now().Add(tm.lifetime)
	tm.tokenMu.Unlock()

	// Launch background refresh goroutine.
	refreshCtx, cancel := context.WithCancel(ctx)
	tm.cancel = cancel
	go tm.refreshLoop(refreshCtx)

	return nil
}

// Stop cancels the background refresh goroutine.
// Idempotent — safe to call multiple times.
func (tm *TokenManager) Stop() {
	if tm.cancel != nil {
		tm.cancel()
	}
}

// Token returns the current valid token. If the token is
// within the proactive refresh window, a non-blocking
// refresh is attempted using TryLock() deduplication.
// Returns an error if the token is empty or expired.
//
// Design decision: Proactive refresh failure preserves
// the existing token (it may still be valid until expiry).
// Only the background refresh loop clears the token on
// failure — this asymmetry prevents a single failed
// proactive refresh from immediately breaking all
// requests (per learning/gateway-3).
func (tm *TokenManager) Token() (string, error) {
	tm.tokenMu.RLock()
	token := tm.token
	expiry := tm.tokenExpiry
	tm.tokenMu.RUnlock()

	if token == "" {
		return "", fmt.Errorf(
			"token unavailable — credential refresh required")
	}

	// Proactive refresh: if within the window, attempt
	// a non-blocking refresh before returning.
	if !expiry.IsZero() &&
		time.Now().Add(tm.proactiveWindow).After(expiry) {
		tm.tryProactiveRefresh()
		// Re-read after proactive refresh attempt.
		tm.tokenMu.RLock()
		token = tm.token
		expiry = tm.tokenExpiry
		tm.tokenMu.RUnlock()
	}

	if !expiry.IsZero() && time.Now().After(expiry) {
		return "", fmt.Errorf(
			"token expired — credential refresh required")
	}

	return token, nil
}

// refreshLoop runs the refresh function on a ticker
// interval. On failure, the token is cleared atomically
// so Token() returns a clear error instead of forwarding
// stale credentials silently.
//
// Design decision: Generic refresh loop shared by all
// providers. The refreshFn closure captures the
// provider-specific logic (per DRY principle). On
// failure, the token is invalidated — this is the
// critical fix from learning/gateway-3.
func (tm *TokenManager) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(tm.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			newToken, err := tm.refreshFn()
			if err != nil {
				log.Error("credential refresh failed",
					"error", err)
				// Invalidate stale token so Token()
				// returns a clear error instead of
				// forwarding expired credentials.
				tm.tokenMu.Lock()
				tm.token = ""
				tm.tokenExpiry = time.Time{}
				tm.tokenMu.Unlock()
				continue
			}
			tm.tokenMu.Lock()
			tm.token = newToken
			tm.tokenExpiry = time.Now().Add(tm.lifetime)
			tm.tokenMu.Unlock()
			log.Info("credential refreshed")
		}
	}
}

// tryProactiveRefresh attempts a non-blocking token
// refresh with a 5-second timeout. Uses TryLock to
// deduplicate concurrent attempts — if another goroutine
// is already refreshing, this call returns immediately
// without blocking.
//
// On failure (including timeout), the existing token is
// preserved (it may still be valid until expiry). Only
// the background refresh loop clears the token on failure.
func (tm *TokenManager) tryProactiveRefresh() {
	if !tm.proactiveMu.TryLock() {
		return // another goroutine is already refreshing
	}
	defer tm.proactiveMu.Unlock()

	type result struct {
		token string
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		t, err := tm.refreshFn()
		ch <- result{t, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			log.Warn("proactive token refresh failed",
				"error", r.err)
			return
		}
		tm.tokenMu.Lock()
		tm.token = r.token
		tm.tokenExpiry = time.Now().Add(tm.lifetime)
		tm.tokenMu.Unlock()
	case <-time.After(proactiveRefreshTimeout):
		log.Warn("proactive token refresh timed out")
		return
	}
}
