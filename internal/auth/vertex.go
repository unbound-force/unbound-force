package auth

import (
	"fmt"
	"strings"
)

// RefreshVertexToken calls `gcloud auth application-default
// print-access-token` to obtain a fresh OAuth token.
// Returns the token string or an error with a clear
// message suggesting re-authentication.
//
// Extracted from internal/gateway/refresh.go per
// design.md D2 to share between gateway and ollama-proxy.
func RefreshVertexToken(
	execCmd func(string, ...string) ([]byte, error),
) (string, error) {
	out, err := execCmd("gcloud", "auth", "application-default",
		"print-access-token")
	if err != nil {
		return "", fmt.Errorf(
			"failed to get Vertex AI token from gcloud. "+
				"Re-authenticate: gcloud auth application-default login\n"+
				"Error: %s", strings.TrimSpace(string(out)))
	}
	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", fmt.Errorf(
			"gcloud returned empty token. "+
				"Re-authenticate: gcloud auth application-default login")
	}
	return token, nil
}
