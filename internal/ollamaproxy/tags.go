package ollamaproxy

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"
)

// ollamaModel represents a single model in the Ollama
// tags response format.
type ollamaModel struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
}

// ollamaTagsResponse is the Ollama tags response format.
type ollamaTagsResponse struct {
	Models []ollamaModel `json:"models"`
}

// handleTags handles GET /api/tags. It returns a synthetic
// model list containing the configured Ollama model names
// from the model mapping table (per spec requirement).
//
// Design decision: Synthetic list from the model map
// rather than querying upstream, because the proxy only
// supports mapped models (unknown models are rejected
// per D5).
func (ps *proxyServer) handleTags(w http.ResponseWriter, _ *http.Request) {
	// Collect model names from the mapping table and
	// sort for deterministic output.
	names := make([]string, 0, len(defaultModelMap))
	for name := range defaultModelMap {
		names = append(names, name)
	}
	sort.Strings(names)

	models := make([]ollamaModel, len(names))
	for i, name := range names {
		models[i] = ollamaModel{
			Name:       name,
			ModifiedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Size:       0,
		}
	}

	resp := ollamaTagsResponse{Models: models}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
