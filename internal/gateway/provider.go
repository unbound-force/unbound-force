package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

// vertexTokenLifetime is the assumed lifetime of a gcloud
// OAuth token. Set to 55 minutes (tokens typically expire
// at 60 minutes, with 5 minutes of safety margin).
const vertexTokenLifetime = 55 * time.Minute

// proactiveRefreshWindow is the window before token expiry
// during which PrepareRequest will attempt a proactive
// refresh. If the token expires within this window, a
// non-blocking refresh is attempted before forwarding the
// request.
const proactiveRefreshWindow = 5 * time.Minute

// bedrockCredLifetime is the assumed lifetime of AWS
// session credentials. Set to 50 minutes (session tokens
// typically expire in 1-12 hours; 50 minutes matches the
// existing refresh interval).
const bedrockCredLifetime = 50 * time.Minute

// Provider abstracts the upstream cloud provider's
// authentication and URL rewriting strategy.
//
// Design decision: Strategy pattern per SOLID Open/Closed
// Principle. Adding a new provider (e.g., OpenAI-compatible)
// requires only a new implementation, not modification of
// the gateway core. This matches the Backend interface
// pattern in internal/sandbox/backend.go.
type Provider interface {
	// Name returns the provider identifier
	// ("anthropic", "vertex", "bedrock").
	Name() string

	// PrepareRequest modifies the outbound request before
	// it is forwarded to the upstream provider.
	// Responsibilities:
	//   - Set the upstream URL (scheme, host, path)
	//   - Inject authentication headers
	//   - Set req.Host to the upstream host
	PrepareRequest(req *http.Request) error

	// Start initializes the provider (e.g., acquire
	// initial OAuth token for Vertex). Called once at
	// gateway startup. Returns error if credentials
	// are not available.
	Start(ctx context.Context) error

	// Stop cleans up provider resources (e.g., stop
	// token refresh goroutine). Called on gateway
	// shutdown. Must be idempotent.
	Stop()
}

// DetectProvider auto-detects the cloud provider from
// environment variables. Priority order (per data-model.md):
//
//  1. CLAUDE_CODE_USE_VERTEX=1 + ANTHROPIC_VERTEX_PROJECT_ID → Vertex
//  2. CLAUDE_CODE_USE_BEDROCK=1 → Bedrock
//  3. ANTHROPIC_API_KEY present → Anthropic
//  4. None matched → error listing supported providers
//
// Vertex is checked first because a developer may have
// both ANTHROPIC_API_KEY and Vertex env vars set (FR-003).
func DetectProvider(
	getenv func(string) string,
	execCmd func(string, ...string) ([]byte, error),
) (Provider, error) {
	// Priority 1: Vertex AI.
	if getenv("CLAUDE_CODE_USE_VERTEX") == "1" &&
		getenv("ANTHROPIC_VERTEX_PROJECT_ID") != "" {
		return newVertexProvider(getenv, execCmd)
	}

	// Priority 2: Bedrock.
	if getenv("CLAUDE_CODE_USE_BEDROCK") == "1" {
		return newBedrockProvider(getenv, execCmd), nil
	}

	// Priority 3: Direct Anthropic.
	if getenv("ANTHROPIC_API_KEY") != "" {
		return newAnthropicProvider(getenv), nil
	}

	// No provider detected.
	return nil, fmt.Errorf(
		"no cloud provider detected. Set one of:\n" +
			"  - ANTHROPIC_API_KEY (direct Anthropic)\n" +
			"  - CLAUDE_CODE_USE_VERTEX=1 + ANTHROPIC_VERTEX_PROJECT_ID (Vertex AI)\n" +
			"  - CLAUDE_CODE_USE_BEDROCK=1 (AWS Bedrock)")
}

// NewProviderByName creates a provider by explicit name.
// Returns an error listing valid names for invalid input
// (FR-009).
func NewProviderByName(
	name string,
	getenv func(string) string,
	execCmd func(string, ...string) ([]byte, error),
) (Provider, error) {
	switch name {
	case "anthropic":
		return newAnthropicProvider(getenv), nil
	case "vertex":
		return newVertexProvider(getenv, execCmd)
	case "bedrock":
		return newBedrockProvider(getenv, execCmd), nil
	default:
		return nil, fmt.Errorf(
			"unknown provider %q. Valid providers: anthropic, vertex, bedrock",
			name)
	}
}

// --- Anthropic Provider ---

// AnthropicProvider forwards requests to the Anthropic
// Messages API with an API key header.
type AnthropicProvider struct {
	apiKey string
	getenv func(string) string
}

func newAnthropicProvider(getenv func(string) string) *AnthropicProvider {
	return &AnthropicProvider{getenv: getenv}
}

// Name returns "anthropic".
func (p *AnthropicProvider) Name() string { return "anthropic" }

// Start reads the API key from the environment. Returns
// an error if ANTHROPIC_API_KEY is empty.
func (p *AnthropicProvider) Start(_ context.Context) error {
	p.apiKey = p.getenv("ANTHROPIC_API_KEY")
	if p.apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}
	return nil
}

// PrepareRequest sets the upstream URL to api.anthropic.com
// and adds the x-api-key header (FR-004).
func (p *AnthropicProvider) PrepareRequest(req *http.Request) error {
	req.URL.Scheme = "https"
	req.URL.Host = "api.anthropic.com"
	req.Host = "api.anthropic.com"
	// Path is preserved as-is (/v1/messages or
	// /v1/messages/count_tokens).
	req.Header.Set("x-api-key", p.apiKey)
	return nil
}

// Stop is a no-op for Anthropic (no refresh needed).
func (p *AnthropicProvider) Stop() {}

// --- Vertex AI Provider ---

// VertexProvider forwards requests to the Vertex AI
// rawPredict endpoint with an OAuth bearer token.
type VertexProvider struct {
	projectID      string
	region         string
	token          string
	tokenExpiry    time.Time
	tokenMu        sync.RWMutex
	tokenRefreshing sync.Mutex
	cancel         context.CancelFunc
	execCmd        func(string, ...string) ([]byte, error)
	getenv         func(string) string
}

func newVertexProvider(
	getenv func(string) string,
	execCmd func(string, ...string) ([]byte, error),
) (*VertexProvider, error) {
	// Region resolution priority:
	//   1. ANTHROPIC_VERTEX_REGION (Claude Code convention)
	//   2. VERTEX_LOCATION (Google Cloud convention)
	//   3. CLOUD_ML_REGION (legacy)
	//   4. Default: us-east5
	// "global" returns an error — Vertex rawPredict
	// requires a specific regional endpoint.
	region := getenv("ANTHROPIC_VERTEX_REGION")
	if region == "" {
		region = getenv("VERTEX_LOCATION")
	}
	if region == "" {
		region = getenv("CLOUD_ML_REGION")
	}
	if region == "global" {
		return nil, fmt.Errorf(
			"vertex region %q is not supported for "+
				"rawPredict/streamRawPredict. These "+
				"endpoints require a specific region "+
				"(e.g., us-east5, europe-west1). Set "+
				"ANTHROPIC_VERTEX_REGION to override "+
				"VERTEX_LOCATION and CLOUD_ML_REGION",
			region)
	}
	if region == "" {
		region = "us-east5"
	}
	return &VertexProvider{
		projectID: getenv("ANTHROPIC_VERTEX_PROJECT_ID"),
		region:    region,
		execCmd:   execCmd,
		getenv:    getenv,
	}, nil
}

// Name returns "vertex".
func (p *VertexProvider) Name() string { return "vertex" }

// Start acquires the initial OAuth token via gcloud and
// starts the refresh goroutine (FR-005).
func (p *VertexProvider) Start(ctx context.Context) error {
	token, err := refreshVertexToken(p.execCmd)
	if err != nil {
		return fmt.Errorf("vertex AI token acquisition failed: %w", err)
	}
	p.tokenMu.Lock()
	p.token = token
	p.tokenExpiry = time.Now().Add(vertexTokenLifetime)
	p.tokenMu.Unlock()

	// Start background refresh goroutine (50-minute
	// interval per research.md R2).
	refreshCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	go refreshLoop(refreshCtx, 50*refreshMinute, func() error {
		newToken, err := refreshVertexToken(p.execCmd)
		if err != nil {
			log.Error("vertex token refresh failed", "error", err)
			// Invalidate stale token so PrepareRequest
			// returns a clear error instead of forwarding
			// an expired token silently.
			p.tokenMu.Lock()
			p.token = ""
			p.tokenExpiry = time.Time{}
			p.tokenMu.Unlock()
			return err
		}
		p.tokenMu.Lock()
		p.token = newToken
		p.tokenExpiry = time.Now().Add(vertexTokenLifetime)
		p.tokenMu.Unlock()
		log.Info("vertex token refreshed")
		return nil
	})

	return nil
}

// PrepareRequest sets the upstream URL to the Vertex
// rawPredict or streamRawPredict endpoint and adds the
// Authorization header with the current OAuth token.
//
// Extended by Spec 034 to:
//   - Transform the request body (remove model, inject
//     anthropic_version) via transformVertexBody (FR-001,
//     FR-002, FR-003, FR-014)
//   - Select rawPredict vs streamRawPredict based on the
//     stream field (FR-005, FR-006)
//   - Strip anthropic-beta and anthropic-version HTTP
//     headers (FR-004)
func (p *VertexProvider) PrepareRequest(req *http.Request) error {
	token, err := p.validToken()
	if err != nil {
		return err
	}

	// Detect count_tokens path before body transformation
	// changes the URL. count_tokens always uses rawPredict.
	isCountTokens := strings.Contains(req.URL.Path, "count_tokens")

	// Transform the request body: remove model, inject
	// anthropic_version, detect stream flag (FR-001,
	// FR-002, FR-003, FR-014).
	model, stream, _ := transformVertexBody(req)

	// Select endpoint action based on stream flag.
	// count_tokens always uses rawPredict (FR-006).
	action := "rawPredict"
	if stream && !isCountTokens {
		action = "streamRawPredict"
	}

	// Build the Vertex endpoint URL.
	// Format: https://{region}-aiplatform.googleapis.com/v1/
	//   projects/{project}/locations/{region}/publishers/
	//   anthropic/models/{model}:{action}
	req.URL.Scheme = "https"
	req.URL.Host = fmt.Sprintf("%s-aiplatform.googleapis.com", p.region)
	req.Host = req.URL.Host
	req.URL.Path = fmt.Sprintf(
		"/v1/projects/%s/locations/%s/publishers/anthropic/models/%s:%s",
		p.projectID, p.region, model, action)

	// Strip SDK-injected headers that Vertex rawPredict
	// does not accept (FR-004).
	req.Header.Del("anthropic-beta")
	req.Header.Del("anthropic-version")

	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// validToken reads the current token and expiry under
// the read lock, triggers a proactive refresh if the
// token is near expiry, and returns an error if the
// token is empty or expired. Extracted from
// PrepareRequest to reduce cyclomatic complexity.
func (p *VertexProvider) validToken() (string, error) {
	p.tokenMu.RLock()
	token := p.token
	expiry := p.tokenExpiry
	p.tokenMu.RUnlock()

	if token == "" {
		return "", fmt.Errorf(
			"vertex AI token unavailable. Re-authenticate: " +
				"gcloud auth application-default login")
	}

	if !expiry.IsZero() &&
		time.Now().Add(proactiveRefreshWindow).After(expiry) {
		p.tryProactiveRefresh()
		p.tokenMu.RLock()
		token = p.token
		expiry = p.tokenExpiry
		p.tokenMu.RUnlock()
	}

	if !expiry.IsZero() && time.Now().After(expiry) {
		return "", fmt.Errorf(
			"vertex AI token expired. Re-authenticate: " +
				"gcloud auth application-default login")
	}

	return token, nil
}

// Stop cancels the refresh goroutine.
func (p *VertexProvider) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

// proactiveRefreshTimeout is the maximum time to wait
// for a proactive refresh subprocess. Prevents a hung
// gcloud or aws CLI from blocking HTTP requests.
const proactiveRefreshTimeout = 5 * time.Second

// tryProactiveRefresh attempts a non-blocking token
// refresh with a 5-second timeout. Uses TryLock to
// deduplicate concurrent attempts — if another goroutine
// is already refreshing, this call returns immediately
// without blocking.
//
// On failure (including timeout), the existing token is
// preserved (it may still be valid until expiry). Only
// the background refresh loop clears the token on failure.
func (p *VertexProvider) tryProactiveRefresh() {
	if !p.tokenRefreshing.TryLock() {
		return // another goroutine is already refreshing
	}
	defer p.tokenRefreshing.Unlock()

	type result struct {
		token string
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		t, err := refreshVertexToken(p.execCmd)
		ch <- result{t, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			log.Warn("vertex proactive token refresh failed",
				"error", r.err)
			return
		}
		p.tokenMu.Lock()
		p.token = r.token
		p.tokenExpiry = time.Now().Add(vertexTokenLifetime)
		p.tokenMu.Unlock()
	case <-time.After(proactiveRefreshTimeout):
		log.Warn("vertex proactive token refresh timed out")
		return
	}
}

// --- Bedrock Provider ---

// BedrockProvider forwards requests to the AWS Bedrock
// invoke endpoint with SigV4 signing.
type BedrockProvider struct {
	region         string
	accessKey      string
	secretKey      string
	sessionToken   string
	credExpiry     time.Time
	credMu         sync.RWMutex
	credRefreshing sync.Mutex
	cancel         context.CancelFunc
	execCmd        func(string, ...string) ([]byte, error)
	getenv         func(string) string
}

func newBedrockProvider(
	getenv func(string) string,
	execCmd func(string, ...string) ([]byte, error),
) *BedrockProvider {
	region := getenv("AWS_REGION")
	if region == "" {
		region = getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}
	return &BedrockProvider{
		region:  region,
		execCmd: execCmd,
		getenv:  getenv,
	}
}

// Name returns "bedrock".
func (p *BedrockProvider) Name() string { return "bedrock" }

// Start acquires initial AWS credentials via the aws CLI
// and starts the refresh goroutine.
func (p *BedrockProvider) Start(ctx context.Context) error {
	ak, sk, st, err := refreshBedrockCredentials(p.execCmd)
	if err != nil {
		return fmt.Errorf("bedrock credential acquisition failed: %w", err)
	}
	p.credMu.Lock()
	p.accessKey = ak
	p.secretKey = sk
	p.sessionToken = st
	p.credExpiry = time.Now().Add(bedrockCredLifetime)
	p.credMu.Unlock()

	// Start background refresh goroutine (50-minute
	// interval — session tokens typically expire in
	// 1-12 hours).
	refreshCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	go refreshLoop(refreshCtx, 50*refreshMinute, func() error {
		newAK, newSK, newST, err := refreshBedrockCredentials(p.execCmd)
		if err != nil {
			log.Error("bedrock credential refresh failed", "error", err)
			// Invalidate stale credentials so
			// PrepareRequest returns a clear error
			// instead of forwarding expired credentials.
			p.credMu.Lock()
			p.accessKey = ""
			p.secretKey = ""
			p.sessionToken = ""
			p.credExpiry = time.Time{}
			p.credMu.Unlock()
			return err
		}
		p.credMu.Lock()
		p.accessKey = newAK
		p.secretKey = newSK
		p.sessionToken = newST
		p.credExpiry = time.Now().Add(bedrockCredLifetime)
		p.credMu.Unlock()
		log.Info("bedrock credentials refreshed")
		return nil
	})

	return nil
}

// PrepareRequest sets the upstream URL to the Bedrock
// invoke endpoint and signs the request with SigV4.
func (p *BedrockProvider) PrepareRequest(req *http.Request) error {
	ak, sk, st, err := p.validCredentials()
	if err != nil {
		return err
	}

	// Extract model from request body. Default to
	// claude-sonnet-4-20250514.
	model := extractModelFromBody(req)

	// Build the Bedrock invoke URL.
	// Format: https://bedrock-runtime.{region}.amazonaws.com/
	//   model/{model}/invoke
	req.URL.Scheme = "https"
	req.URL.Host = fmt.Sprintf("bedrock-runtime.%s.amazonaws.com", p.region)
	req.Host = req.URL.Host
	req.URL.Path = fmt.Sprintf("/model/%s/invoke", model)

	// Sign the request with SigV4.
	if err := signV4(req, p.region, "bedrock-runtime", ak, sk, st); err != nil {
		return fmt.Errorf("sigv4 signing failed: %w", err)
	}
	return nil
}

// validCredentials reads the current credentials and
// expiry under the read lock, triggers a proactive
// refresh if near expiry, and returns an error if
// credentials are empty or expired. Extracted from
// PrepareRequest to reduce cyclomatic complexity.
func (p *BedrockProvider) validCredentials() (ak, sk, st string, err error) {
	p.credMu.RLock()
	ak = p.accessKey
	sk = p.secretKey
	st = p.sessionToken
	expiry := p.credExpiry
	p.credMu.RUnlock()

	if ak == "" || sk == "" {
		return "", "", "", fmt.Errorf(
			"bedrock credentials unavailable. Re-authenticate: aws sso login")
	}

	if !expiry.IsZero() &&
		time.Now().Add(proactiveRefreshWindow).After(expiry) {
		p.tryProactiveRefreshBedrock()
		p.credMu.RLock()
		ak = p.accessKey
		sk = p.secretKey
		st = p.sessionToken
		expiry = p.credExpiry
		p.credMu.RUnlock()
	}

	if !expiry.IsZero() && time.Now().After(expiry) {
		return "", "", "", fmt.Errorf(
			"bedrock credentials expired. Re-authenticate: aws sso login")
	}

	return ak, sk, st, nil
}

// Stop cancels the refresh goroutine.
func (p *BedrockProvider) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

// tryProactiveRefreshBedrock attempts a non-blocking
// credential refresh with a 5-second timeout. Uses
// TryLock to deduplicate concurrent attempts — if
// another goroutine is already refreshing, this call
// returns immediately.
//
// On failure (including timeout), existing credentials
// are preserved (they may still be valid until expiry).
// Only the background refresh loop clears credentials
// on failure.
func (p *BedrockProvider) tryProactiveRefreshBedrock() {
	if !p.credRefreshing.TryLock() {
		return // another goroutine is already refreshing
	}
	defer p.credRefreshing.Unlock()

	type result struct {
		ak, sk, st string
		err        error
	}
	ch := make(chan result, 1)
	go func() {
		ak, sk, st, err := refreshBedrockCredentials(p.execCmd)
		ch <- result{ak, sk, st, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			log.Warn("bedrock proactive credential refresh failed",
				"error", r.err)
			return
		}
		p.credMu.Lock()
		p.accessKey = r.ak
		p.secretKey = r.sk
		p.sessionToken = r.st
		p.credExpiry = time.Now().Add(bedrockCredLifetime)
		p.credMu.Unlock()
	case <-time.After(proactiveRefreshTimeout):
		log.Warn("bedrock proactive credential refresh timed out")
		return
	}
}

// --- Helpers ---

// transformVertexBody reads the request body, removes the
// `model` field, injects `anthropic_version` if absent,
// and replaces the body. Returns the extracted model name
// and stream flag.
//
// Uses map[string]any to preserve all unknown fields
// without requiring a struct definition for the full
// Anthropic Messages API body (per research.md R7).
func transformVertexBody(req *http.Request) (model string, stream bool, err error) {
	defaultModel := "claude-sonnet-4-20250514"

	if req.Body == nil {
		return defaultModel, false, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return defaultModel, false, nil
	}

	if len(body) == 0 {
		req.Body = io.NopCloser(bytes.NewReader(body))
		return defaultModel, false, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		// Malformed JSON — forward unchanged (edge case).
		req.Body = io.NopCloser(bytes.NewReader(body))
		return defaultModel, false, nil
	}

	// Extract model (FR-001).
	if m, ok := payload["model"].(string); ok && m != "" {
		model = m
	} else {
		model = defaultModel
	}
	delete(payload, "model")

	// Extract stream flag (FR-005, FR-006).
	if s, ok := payload["stream"].(bool); ok {
		stream = s
	}

	// Inject anthropic_version if absent (FR-002, FR-003).
	if _, ok := payload["anthropic_version"]; !ok {
		payload["anthropic_version"] = "vertex-2023-10-16"
	}

	// Re-encode.
	newBody, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		// Marshal failed — forward original body.
		req.Body = io.NopCloser(bytes.NewReader(body))
		return model, stream, nil
	}

	req.Body = io.NopCloser(bytes.NewReader(newBody))
	req.ContentLength = int64(len(newBody))
	return model, stream, nil
}

// extractModelFromBody reads the request body to extract
// the "model" field, then replaces the body so it can be
// read again by the proxy. Returns a default model if the
// body cannot be parsed.
func extractModelFromBody(req *http.Request) string {
	defaultModel := "claude-sonnet-4-20250514"

	if req.Body == nil {
		return defaultModel
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return defaultModel
	}
	// Replace the body so the proxy can read it.
	req.Body = io.NopCloser(bytes.NewReader(body))

	var payload struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.Model == "" {
		return defaultModel
	}

	// Bedrock uses a different model ID format — strip
	// the "anthropic." prefix if present.
	model := payload.Model
	if strings.HasPrefix(model, "anthropic.") {
		return model
	}

	return model
}
