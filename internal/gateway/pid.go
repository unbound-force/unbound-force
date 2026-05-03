// Package gateway implements a minimal LLM reverse proxy
// that runs on the host machine and serves the Anthropic
// Messages API. It auto-detects the cloud provider
// (Anthropic, Vertex AI, Bedrock) from environment
// variables, injects host-side credentials into upstream
// requests, and auto-refreshes OAuth tokens. All external
// dependencies are injected for testability per
// Constitution Principle IV.
//
// PID file management has been extracted to
// internal/pidfile/ for reuse by other daemons
// (e.g., ollama-proxy). Import pidfile.WritePID,
// pidfile.ReadPID, pidfile.IsAlive, pidfile.CleanupStale,
// pidfile.RemovePID, and pidfile.PIDInfo from that package.
package gateway
