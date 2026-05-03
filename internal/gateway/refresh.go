package gateway

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

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
