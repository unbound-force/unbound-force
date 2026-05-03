package auth

import (
	"fmt"
	"strings"
	"testing"
)

func TestRefreshVertexToken_Success(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("ya29.fresh-token\n"), nil
	}

	token, err := RefreshVertexToken(execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "ya29.fresh-token" {
		t.Errorf("token: got %q, want ya29.fresh-token", token)
	}
}

func TestRefreshVertexToken_Failure(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("ERROR: not authenticated"), fmt.Errorf("exit 1")
	}

	_, err := RefreshVertexToken(execCmd)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Re-authenticate") {
		t.Errorf("expected re-auth suggestion, got: %s", err.Error())
	}
}

func TestRefreshVertexToken_EmptyOutput(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("   \n"), nil
	}

	_, err := RefreshVertexToken(execCmd)
	if err == nil {
		t.Fatal("expected error for empty token output")
	}
	if !strings.Contains(err.Error(), "empty token") {
		t.Errorf("expected empty token error, got: %s", err.Error())
	}
}

func TestRefreshVertexToken_TrimsWhitespace(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("  ya29.trimmed  \n"), nil
	}

	token, err := RefreshVertexToken(execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "ya29.trimmed" {
		t.Errorf("token: got %q, want ya29.trimmed", token)
	}
}

func TestRefreshVertexToken_CommandArgs(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	execCmd := func(name string, args ...string) ([]byte, error) {
		capturedName = name
		capturedArgs = args
		return []byte("ya29.test\n"), nil
	}

	_, _ = RefreshVertexToken(execCmd)

	if capturedName != "gcloud" {
		t.Errorf("command: got %q, want gcloud", capturedName)
	}
	expectedArgs := []string{"auth", "application-default", "print-access-token"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("args count: got %d, want %d", len(capturedArgs), len(expectedArgs))
	}
	for i, arg := range expectedArgs {
		if capturedArgs[i] != arg {
			t.Errorf("arg[%d]: got %q, want %q", i, capturedArgs[i], arg)
		}
	}
}
