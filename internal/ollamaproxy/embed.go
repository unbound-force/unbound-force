package ollamaproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
)

// ollamaEmbedRequest is the Ollama embed request format.
// Dewey sends: {"model": "<name>", "input": ["text"]}.
type ollamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// ollamaEmbedResponse is the Ollama embed response format.
// Dewey expects: {"model": "<name>", "embeddings": [[float...]]}.
type ollamaEmbedResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
}

// vertexEmbedInstance is a single instance in the Vertex
// AI embedding predict request.
type vertexEmbedInstance struct {
	Content string `json:"content"`
}

// vertexEmbedRequest is the Vertex AI embedding predict
// request format.
type vertexEmbedRequest struct {
	Instances []vertexEmbedInstance `json:"instances"`
}

// vertexEmbedResponse is the Vertex AI embedding predict
// response format.
type vertexEmbedResponse struct {
	Predictions []vertexEmbedPrediction `json:"predictions"`
}

// vertexEmbedPrediction is a single prediction in the
// Vertex AI response.
type vertexEmbedPrediction struct {
	Embeddings vertexEmbedValues `json:"embeddings"`
}

// vertexEmbedValues holds the embedding vector from
// Vertex AI.
type vertexEmbedValues struct {
	Values []float64 `json:"values"`
}

// handleEmbed handles POST /api/embed. It translates
// Ollama embed requests to Vertex AI embedding predict
// requests and returns Ollama-format responses.
//
// Design decision: Direct HTTP call to Vertex AI (not
// reverse proxy) because the proxy must translate between
// two different API formats (per design.md D3).
func (ps *proxyServer) handleEmbed(w http.ResponseWriter, r *http.Request) {
	// Enforce max request body size (10MB).
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		// MaxBytesReader returns a specific error type
		// when the limit is exceeded.
		if err.Error() == "http: request body too large" {
			writeOllamaError(w, http.StatusRequestEntityTooLarge,
				"request body too large (max 10MB)")
			return
		}
		writeOllamaError(w, http.StatusBadRequest,
			fmt.Sprintf("failed to read request body: %s", err))
		return
	}

	var req ollamaEmbedRequest
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

	// Map Ollama model name to Vertex model (D5).
	cloudModel, ok := mapModelName(req.Model)
	if !ok {
		writeOllamaError(w, http.StatusBadRequest,
			fmt.Sprintf("unknown model %q: not in model mapping table. "+
				"Available models: granite-embedding:30m, "+
				"granite-embedding-small-english-r2, llama3.2:3b",
				req.Model))
		return
	}

	// Handle empty input gracefully.
	if len(req.Input) == 0 {
		writeOllamaError(w, http.StatusBadRequest,
			"input is required and must not be empty")
		return
	}

	// Build Vertex AI request with batch support.
	instances := make([]vertexEmbedInstance, len(req.Input))
	for i, text := range req.Input {
		instances[i] = vertexEmbedInstance{Content: text}
	}
	vertexReq := vertexEmbedRequest{Instances: instances}

	vertexBody, err := json.Marshal(vertexReq)
	if err != nil {
		writeOllamaError(w, http.StatusInternalServerError,
			fmt.Sprintf("failed to marshal Vertex request: %s", err))
		return
	}

	// Construct Vertex AI predict URL (per design.md D9).
	region := ps.opts.Getenv("ANTHROPIC_VERTEX_REGION")
	if region == "" {
		region = ps.opts.Getenv("VERTEX_LOCATION")
	}
	if region == "" {
		region = ps.opts.Getenv("CLOUD_ML_REGION")
	}
	if region == "" {
		region = "us-east5"
	}

	projectID := ps.opts.Getenv("ANTHROPIC_VERTEX_PROJECT_ID")
	if projectID == "" {
		projectID = ps.opts.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	// URL construction uses fmt.Sprintf with validated
	// model name (safe characters only, per D5).
	vertexURL := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/"+
			"locations/%s/publishers/google/models/%s:predict",
		region, projectID, region, cloudModel)

	// Get OAuth token from TokenManager.
	token, err := ps.tokenMgr.Token()
	if err != nil {
		writeOllamaError(w, http.StatusServiceUnavailable,
			"authentication unavailable — credential refresh required")
		return
	}

	// Call Vertex AI endpoint.
	httpReq, err := http.NewRequestWithContext(r.Context(),
		http.MethodPost, vertexURL, bytes.NewReader(vertexBody))
	if err != nil {
		writeOllamaError(w, http.StatusInternalServerError,
			fmt.Sprintf("failed to create Vertex request: %s", err))
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := ps.opts.HTTPClient.Do(httpReq)
	if err != nil {
		writeOllamaError(w, http.StatusBadGateway,
			fmt.Sprintf("Vertex AI request failed: %s",
				redactToken(err.Error())))
		return
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writeOllamaError(w, http.StatusBadGateway,
			fmt.Sprintf("failed to read Vertex response: %s", err))
		return
	}

	// Handle non-200 responses from Vertex AI.
	if resp.StatusCode != http.StatusOK {
		// Redact any tokens that Vertex may have echoed
		// in the error response (per D8a).
		redacted := redactToken(string(respBody))
		log.Error("vertex AI embedding error",
			"status", resp.StatusCode,
			"body_length", len(respBody))
		writeOllamaError(w, http.StatusBadGateway,
			fmt.Sprintf("Vertex AI error (HTTP %d): %s",
				resp.StatusCode, redacted))
		return
	}

	// Parse Vertex response.
	var vertexResp vertexEmbedResponse
	if err := json.Unmarshal(respBody, &vertexResp); err != nil {
		writeOllamaError(w, http.StatusBadGateway,
			fmt.Sprintf("failed to parse Vertex response: %s", err))
		return
	}

	// Build Ollama response: transform
	// predictions[].embeddings.values to embeddings[].
	embeddings := make([][]float64, len(vertexResp.Predictions))
	for i, pred := range vertexResp.Predictions {
		embeddings[i] = pred.Embeddings.Values
	}

	ollamaResp := ollamaEmbedResponse{
		Model:      req.Model,
		Embeddings: embeddings,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ollamaResp)
}
