package auth

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RefreshBedrockCredentials calls `aws configure
// export-credentials --format env` to obtain fresh AWS
// credentials. Returns the access key, secret key,
// session token, and any error.
//
// Extracted from internal/gateway/refresh.go per
// design.md D2 to share between gateway and ollama-proxy.
func RefreshBedrockCredentials(
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
	envVars := ParseEnvExport(string(out))
	ak := envVars["AWS_ACCESS_KEY_ID"]
	sk := envVars["AWS_SECRET_ACCESS_KEY"]
	st := envVars["AWS_SESSION_TOKEN"]

	if ak == "" || sk == "" {
		// Try JSON format as fallback.
		ak, sk, st, err = ParseAWSCredentialsJSON(out)
		if err != nil {
			return "", "", "", fmt.Errorf(
				"could not parse AWS credentials. "+
					"Re-authenticate: aws sso login")
		}
	}

	return ak, sk, st, nil
}

// ParseEnvExport parses `export KEY=VALUE` lines into a map.
func ParseEnvExport(output string) map[string]string {
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

// ParseAWSCredentialsJSON parses JSON-format AWS credentials.
func ParseAWSCredentialsJSON(data []byte) (string, string, string, error) {
	var creds struct {
		AccessKeyID    string `json:"AccessKeyId"`
		SecretAccessKey string `json:"SecretAccessKey"`
		SessionToken   string `json:"SessionToken"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", "", "", err
	}
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return "", "", "", fmt.Errorf("missing required credential fields")
	}
	return creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, nil
}
