package auth

import (
	"fmt"
	"strings"
	"testing"
)

func TestRefreshBedrockCredentials_EnvFormat(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte(
			"export AWS_ACCESS_KEY_ID=AKIATEST\n" +
				"export AWS_SECRET_ACCESS_KEY=secrettest\n" +
				"export AWS_SESSION_TOKEN=sessiontest\n"), nil
	}

	ak, sk, st, err := RefreshBedrockCredentials(execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ak != "AKIATEST" {
		t.Errorf("accessKey: got %q, want AKIATEST", ak)
	}
	if sk != "secrettest" {
		t.Errorf("secretKey: got %q, want secrettest", sk)
	}
	if st != "sessiontest" {
		t.Errorf("sessionToken: got %q, want sessiontest", st)
	}
}

func TestRefreshBedrockCredentials_JSONFormat(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte(`{
			"AccessKeyId": "AKIAJSON",
			"SecretAccessKey": "secretjson",
			"SessionToken": "sessionjson"
		}`), nil
	}

	ak, sk, st, err := RefreshBedrockCredentials(execCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ak != "AKIAJSON" {
		t.Errorf("accessKey: got %q, want AKIAJSON", ak)
	}
	if sk != "secretjson" {
		t.Errorf("secretKey: got %q, want secretjson", sk)
	}
	if st != "sessionjson" {
		t.Errorf("sessionToken: got %q, want sessionjson", st)
	}
}

func TestRefreshBedrockCredentials_Failure(t *testing.T) {
	execCmd := func(name string, args ...string) ([]byte, error) {
		return []byte("Unable to locate credentials"), fmt.Errorf("exit 1")
	}

	_, _, _, err := RefreshBedrockCredentials(execCmd)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Re-authenticate") {
		t.Errorf("expected re-auth suggestion, got: %s", err.Error())
	}
}

func TestRefreshBedrockCredentials_CommandArgs(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	execCmd := func(name string, args ...string) ([]byte, error) {
		capturedName = name
		capturedArgs = args
		return []byte(
			"export AWS_ACCESS_KEY_ID=AK\n" +
				"export AWS_SECRET_ACCESS_KEY=SK\n"), nil
	}

	_, _, _, _ = RefreshBedrockCredentials(execCmd)

	if capturedName != "aws" {
		t.Errorf("command: got %q, want aws", capturedName)
	}
	expectedArgs := []string{"configure", "export-credentials", "--format", "env"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("args count: got %d, want %d", len(capturedArgs), len(expectedArgs))
	}
	for i, arg := range expectedArgs {
		if capturedArgs[i] != arg {
			t.Errorf("arg[%d]: got %q, want %q", i, capturedArgs[i], arg)
		}
	}
}

func TestParseEnvExport_Standard(t *testing.T) {
	output := "export AWS_ACCESS_KEY_ID=AKIA123\n" +
		"export AWS_SECRET_ACCESS_KEY=secret456\n" +
		"export AWS_SESSION_TOKEN=session789\n"

	result := ParseEnvExport(output)
	if result["AWS_ACCESS_KEY_ID"] != "AKIA123" {
		t.Errorf("AWS_ACCESS_KEY_ID: got %q, want AKIA123",
			result["AWS_ACCESS_KEY_ID"])
	}
	if result["AWS_SECRET_ACCESS_KEY"] != "secret456" {
		t.Errorf("AWS_SECRET_ACCESS_KEY: got %q, want secret456",
			result["AWS_SECRET_ACCESS_KEY"])
	}
	if result["AWS_SESSION_TOKEN"] != "session789" {
		t.Errorf("AWS_SESSION_TOKEN: got %q, want session789",
			result["AWS_SESSION_TOKEN"])
	}
}

func TestParseEnvExport_QuotedValues(t *testing.T) {
	output := "export KEY=\"quoted-value\"\n" +
		"export KEY2='single-quoted'\n"

	result := ParseEnvExport(output)
	if result["KEY"] != "quoted-value" {
		t.Errorf("KEY: got %q, want quoted-value", result["KEY"])
	}
	if result["KEY2"] != "single-quoted" {
		t.Errorf("KEY2: got %q, want single-quoted", result["KEY2"])
	}
}

func TestParseEnvExport_EmptyInput(t *testing.T) {
	result := ParseEnvExport("")
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestParseEnvExport_NoExportPrefix(t *testing.T) {
	output := "KEY=value\n"
	result := ParseEnvExport(output)
	if result["KEY"] != "value" {
		t.Errorf("KEY: got %q, want value", result["KEY"])
	}
}

func TestParseAWSCredentialsJSON_Valid(t *testing.T) {
	data := []byte(`{
		"AccessKeyId": "AKIA",
		"SecretAccessKey": "secret",
		"SessionToken": "session"
	}`)
	ak, sk, st, err := ParseAWSCredentialsJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ak != "AKIA" {
		t.Errorf("AccessKeyId: got %q, want AKIA", ak)
	}
	if sk != "secret" {
		t.Errorf("SecretAccessKey: got %q, want secret", sk)
	}
	if st != "session" {
		t.Errorf("SessionToken: got %q, want session", st)
	}
}

func TestParseAWSCredentialsJSON_MissingFields(t *testing.T) {
	data := []byte(`{"AccessKeyId":"AKIA","SecretAccessKey":""}`)
	_, _, _, err := ParseAWSCredentialsJSON(data)
	if err == nil {
		t.Fatal("expected error for missing SecretAccessKey")
	}
}

func TestParseAWSCredentialsJSON_InvalidJSON(t *testing.T) {
	data := []byte(`not json`)
	_, _, _, err := ParseAWSCredentialsJSON(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseAWSCredentialsJSON_NoSessionToken(t *testing.T) {
	data := []byte(`{
		"AccessKeyId": "AKIA",
		"SecretAccessKey": "secret"
	}`)
	ak, sk, st, err := ParseAWSCredentialsJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ak != "AKIA" || sk != "secret" {
		t.Errorf("credentials: got (%q, %q), want (AKIA, secret)", ak, sk)
	}
	if st != "" {
		t.Errorf("SessionToken: got %q, want empty", st)
	}
}
