package ollamaproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
)

// defaultMaxTokens is the max_tokens value sent to the
// Anthropic Messages API. 4096 is a reasonable default
// for generation tasks (per spec requirement).
const defaultMaxTokens = 4096

// ollamaGenerateRequest is the Ollama generate request
// format. Dewey sends:
// {"model": "<name>", "prompt": "...", "stream": false}.
type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaGenerateResponse is the Ollama generate response
// format. Dewey expects:
// {"model": "<name>", "response": "..."}.
type ollamaGenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
}

// anthropicMessage is a single message in the Anthropic
// Messages API format.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicRequest is the Anthropic Messages API request
// format.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

// anthropicContentBlock is a single content block in the
// Anthropic response.
type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// anthropicResponse is the Anthropic Messages API response
// format (simplified — only the fields we need).
type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
}

// handleGenerate handles POST /api/generate. It translates
// Ollama generate requests to Anthropic Messages API
// requests and delegates to the gateway (per design.md D4).
func (ps *proxyServer) handleGenerate(w http.ResponseWriter, r *http.Request) {
	// Enforce max request body size (10MB).
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			writeOllamaError(w, http.StatusRequestEntityTooLarge,
				"request body too large (max 10MB)")
			return
		}
		writeOllamaError(w, http.StatusBadRequest,
			fmt.Sprintf("failed to read request body: %s", err))
		return
	}

	var req ollamaGenerateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOllamaError(w, http.StatusBadRequest,
			fmt.Sprintf("invalid JSON: %s", err))
		return
	}

	// Validate model name for safe characters (D5).
	if err := validateModelName(req.Model); err != nil {
		writeOllamaError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Map Ollama model name to Claude model (D5).
	cloudModel, ok := mapModelName(req.Model)
	if !ok {
		writeOllamaError(w, http.StatusBadRequest,
			fmt.Sprintf("unknown model %q: not in model mapping table. "+
				"Available models: granite-embedding:30m, "+
				"granite-embedding-small-english-r2, llama3.2:3b",
				req.Model))
		return
	}

	// Build Anthropic Messages request.
	anthropicReq := anthropicRequest{
		Model:     cloudModel,
		MaxTokens: defaultMaxTokens,
		Messages: []anthropicMessage{
			{Role: "user", Content: req.Prompt},
		},
	}

	anthropicBody, err := json.Marshal(anthropicReq)
	if err != nil {
		writeOllamaError(w, http.StatusInternalServerError,
			fmt.Sprintf("failed to marshal Anthropic request: %s", err))
		return
	}

	// POST to gateway URL (per design.md D4).
	gatewayMessagesURL := ps.opts.GatewayURL + "/v1/messages"

	httpReq, err := http.NewRequestWithContext(r.Context(),
		http.MethodPost, gatewayMessagesURL,
		bytes.NewReader(anthropicBody))
	if err != nil {
		writeOllamaError(w, http.StatusInternalServerError,
			fmt.Sprintf("failed to create gateway request: %s", err))
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := ps.opts.HTTPClient.Do(httpReq)
	if err != nil {
		// Gateway unavailable — return clear error with
		// instructions (per spec requirement).
		log.Error("gateway request failed", "error", err)
		writeOllamaError(w, http.StatusBadGateway,
			"gateway not available — run `uf gateway start` "+
				"to enable generation. The gateway handles "+
				"Anthropic/Vertex translation for /api/generate")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeOllamaError(w, http.StatusBadGateway,
			fmt.Sprintf("failed to read gateway response: %s", err))
		return
	}

	// Handle non-200 responses from gateway.
	if resp.StatusCode != http.StatusOK {
		log.Error("gateway error",
			"status", resp.StatusCode,
			"body_length", len(respBody))
		writeOllamaError(w, http.StatusBadGateway,
			fmt.Sprintf("gateway error (HTTP %d): %s",
				resp.StatusCode, string(respBody)))
		return
	}

	// Parse Anthropic response.
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		writeOllamaError(w, http.StatusBadGateway,
			fmt.Sprintf("failed to parse gateway response: %s", err))
		return
	}

	// Extract content[0].text — handle empty content
	// array gracefully (no panic, per task 5.6).
	var responseText string
	if len(anthropicResp.Content) > 0 {
		responseText = anthropicResp.Content[0].Text
	} else {
		writeOllamaError(w, http.StatusBadGateway,
			"gateway returned empty content — no text "+
				"in response")
		return
	}

	// Build Ollama response.
	ollamaResp := ollamaGenerateResponse{
		Model:    req.Model,
		Response: responseText,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ollamaResp)
}
