package gateway

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strings"
)

// sseFilterReader wraps an io.ReadCloser and drops SSE
// events whose `event:` field matches a filtered type.
// Used to remove `vertex_event` and `ping` events from
// Vertex streaming responses (FR-007, FR-008).
//
// Design decision: Implemented as an io.ReadCloser
// wrapper applied in ModifyResponse, keeping the filter
// composable and testable — it operates on io.Reader
// without knowledge of HTTP (per research.md R1).
type sseFilterReader struct {
	source   io.ReadCloser
	scanner  *bufio.Scanner
	buf      bytes.Buffer
	filtered map[string]bool
	done     bool
}

// newSSEFilterReader creates a filter that drops events
// with the given event types. Sets scanner buffer to 1MB
// to handle large SSE data lines (per research.md R1).
func newSSEFilterReader(source io.ReadCloser, filtered map[string]bool) *sseFilterReader {
	s := bufio.NewScanner(source)
	// Increase buffer to 1MB to handle large SSE data
	// lines (content blocks can be large).
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &sseFilterReader{
		source:   source,
		scanner:  s,
		filtered: filtered,
	}
}

// Read implements io.Reader. It accumulates SSE lines
// until a blank-line event boundary, then either drops
// or forwards the complete event.
func (r *sseFilterReader) Read(p []byte) (int, error) {
	// If buffer has data from a previous event, drain it.
	if r.buf.Len() > 0 {
		return r.buf.Read(p)
	}

	if r.done {
		return 0, io.EOF
	}

	// Accumulate lines until we see a blank line (event
	// boundary). An SSE event is one or more field lines
	// followed by a blank line.
	for {
		var event strings.Builder
		var eventType string
		foundEvent := false

		for r.scanner.Scan() {
			line := r.scanner.Text()
			if line == "" {
				// Blank line = end of event.
				foundEvent = true
				if eventType != "" && r.filtered[eventType] {
					// Drop this event — reset and try next.
					break
				}
				// Forward this event.
				event.WriteString("\n") // blank line separator
				r.buf.WriteString(event.String())
				return r.buf.Read(p)
			}
			event.WriteString(line)
			event.WriteString("\n")
			if strings.HasPrefix(line, "event: ") {
				eventType = strings.TrimPrefix(line, "event: ")
			}
		}

		if !foundEvent {
			// Scanner exhausted — return remaining data
			// or EOF.
			r.done = true
			if event.Len() > 0 {
				r.buf.WriteString(event.String())
				return r.buf.Read(p)
			}
			return 0, io.EOF
		}

		// If we get here, the event was filtered. Loop
		// to try the next event.
	}
}

// Close delegates to the underlying source's Close.
func (r *sseFilterReader) Close() error {
	return r.source.Close()
}

// vertexSSEFilter returns a ModifyResponse function for
// httputil.ReverseProxy that filters Vertex-specific SSE
// events from streaming responses.
//
// Only applied when Content-Type starts with
// text/event-stream (FR-009 — non-streaming responses
// pass through unchanged).
func vertexSSEFilter() func(*http.Response) error {
	return func(resp *http.Response) error {
		ct := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "text/event-stream") {
			return nil // Non-streaming — pass through.
		}

		filtered := map[string]bool{
			"vertex_event": true,
			"ping":         true,
		}
		resp.Body = newSSEFilterReader(resp.Body, filtered)
		// Remove Content-Length since filtering changes
		// the body size (chunked encoding handles this).
		resp.Header.Del("Content-Length")
		resp.ContentLength = -1
		return nil
	}
}
