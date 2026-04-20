package gateway

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// refreshMinute is the duration of one minute, used as
// the base unit for refresh intervals. Defined as a
// variable (not const) so tests can override it.
var refreshMinute = time.Minute

// refreshLoop runs a generic refresh function on a ticker
// interval. Cancels when the context is done. Logs errors
// but does not stop on failure (the next tick will retry).
//
// Design decision: Generic refresh loop shared by Vertex
// and Bedrock providers. The refreshFn closure captures
// the provider-specific logic (per DRY principle).
func refreshLoop(ctx context.Context, interval time.Duration, refreshFn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := refreshFn(); err != nil {
				log.Error("credential refresh failed", "error", err)
			}
		}
	}
}

// refreshVertexToken calls `gcloud auth application-default
// print-access-token` to obtain a fresh OAuth token.
// Returns the token string or an error with a clear
// message suggesting re-authentication (US3 scenario 2).
func refreshVertexToken(
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

// refreshBedrockCredentials calls `aws configure
// export-credentials --format env` to obtain fresh AWS
// credentials. Returns the access key, secret key,
// session token, and any error.
func refreshBedrockCredentials(
	execCmd func(string, ...string) ([]byte, error),
) (accessKey, secretKey, sessionToken string, err error) {
	out, err := execCmd("aws", "configure", "export-credentials",
		"--format", "env")
	if err != nil {
		return "", "", "", fmt.Errorf(
			"failed to get AWS credentials from aws CLI. "+
				"Re-authenticate: aws sso login\n"+
				"Error: %s", strings.TrimSpace(string(out)))
	}

	// The output is in env format:
	//   export AWS_ACCESS_KEY_ID=...
	//   export AWS_SECRET_ACCESS_KEY=...
	//   export AWS_SESSION_TOKEN=...
	// Parse the key=value pairs.
	envVars := parseEnvExport(string(out))
	ak := envVars["AWS_ACCESS_KEY_ID"]
	sk := envVars["AWS_SECRET_ACCESS_KEY"]
	st := envVars["AWS_SESSION_TOKEN"]

	if ak == "" || sk == "" {
		// Try JSON format as fallback.
		ak, sk, st, err = parseAWSCredentialsJSON(out)
		if err != nil {
			return "", "", "", fmt.Errorf(
				"could not parse AWS credentials. "+
					"Re-authenticate: aws sso login")
		}
	}

	return ak, sk, st, nil
}

// parseEnvExport parses `export KEY=VALUE` lines into a map.
func parseEnvExport(output string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		// Remove surrounding quotes if present.
		value = strings.Trim(value, "\"'")
		result[key] = value
	}
	return result
}

// parseAWSCredentialsJSON parses JSON-format AWS credentials.
func parseAWSCredentialsJSON(data []byte) (string, string, string, error) {
	var creds struct {
		AccessKeyID     string `json:"AccessKeyId"`
		SecretAccessKey string `json:"SecretAccessKey"`
		SessionToken    string `json:"SessionToken"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", "", "", err
	}
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return "", "", "", fmt.Errorf("missing required credential fields")
	}
	return creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, nil
}

// --- SigV4 Signing ---

// signV4 signs an HTTP request with AWS Signature Version 4.
// This is a minimal implementation using crypto/hmac +
// crypto/sha256, covering the subset needed for Bedrock
// invoke requests (~200 lines per research.md R3).
//
// The algorithm follows the AWS documentation:
// https://docs.aws.amazon.com/IAM/latest/UserGuide/
// reference_sigv.html
func signV4(
	req *http.Request,
	region, service string,
	accessKey, secretKey, sessionToken string,
) error {
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	// Set required headers.
	req.Header.Set("X-Amz-Date", amzDate)
	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	// Ensure Host header is set.
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}

	// Step 1: Create canonical request.
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalQueryString := req.URL.Query().Encode()

	// Build signed headers (sorted, lowercase).
	signedHeaders, canonicalHeaders := buildCanonicalHeaders(req)

	// Hash the payload.
	payloadHash, err := hashPayload(req)
	if err != nil {
		return fmt.Errorf("hash payload: %w", err)
	}
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	// Step 2: Create string to sign.
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request",
		dateStamp, region, service)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	// Step 3: Calculate signature.
	signingKey := deriveSigningKey(secretKey, dateStamp, region, service)
	signature := hmacSHA256Hex(signingKey, []byte(stringToSign))

	// Step 4: Add Authorization header.
	authHeader := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)

	return nil
}

// buildCanonicalHeaders builds the canonical headers string
// and signed headers list for SigV4. Headers are sorted by
// lowercase name.
func buildCanonicalHeaders(req *http.Request) (signedHeaders, canonicalHeaders string) {
	type header struct {
		key   string
		value string
	}

	var headers []header
	for k, vals := range req.Header {
		lk := strings.ToLower(k)
		// Only include headers that should be signed.
		switch lk {
		case "host", "content-type", "x-amz-date",
			"x-amz-security-token", "x-amz-content-sha256":
			headers = append(headers, header{
				key:   lk,
				value: strings.TrimSpace(vals[0]),
			})
		}
	}

	// Add Host if not already in headers (it may be
	// implicit from req.URL.Host).
	hasHost := false
	for _, h := range headers {
		if h.key == "host" {
			hasHost = true
			break
		}
	}
	if !hasHost {
		headers = append(headers, header{
			key:   "host",
			value: req.URL.Host,
		})
	}

	sort.Slice(headers, func(i, j int) bool {
		return headers[i].key < headers[j].key
	})

	var canonicalBuf strings.Builder
	var signedBuf strings.Builder
	for i, h := range headers {
		canonicalBuf.WriteString(h.key)
		canonicalBuf.WriteString(":")
		canonicalBuf.WriteString(h.value)
		canonicalBuf.WriteString("\n")
		if i > 0 {
			signedBuf.WriteString(";")
		}
		signedBuf.WriteString(h.key)
	}

	return signedBuf.String(), canonicalBuf.String()
}

// hashPayload returns the SHA256 hex digest of the request
// body. Replaces the body so it can be read again by the
// proxy.
func hashPayload(req *http.Request) (string, error) {
	if req.Body == nil {
		return sha256Hex([]byte("")), nil
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	return sha256Hex(body), nil
}

// deriveSigningKey derives the SigV4 signing key from the
// secret key, date, region, and service.
func deriveSigningKey(secretKey, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}

// sha256Hex returns the SHA256 hex digest of data.
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// hmacSHA256 computes HMAC-SHA256.
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// hmacSHA256Hex computes HMAC-SHA256 and returns the hex
// digest.
func hmacSHA256Hex(key, data []byte) string {
	return hex.EncodeToString(hmacSHA256(key, data))
}
