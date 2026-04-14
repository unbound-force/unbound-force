## Why

`uf sandbox start --mode direct` mounts only
`~/.config/gcloud/application_default_credentials.json`
into the container. This file is type `authorized_user`
(from `gcloud auth application-default login`) which
contains a refresh token but not a usable access token.

The Google Auth library needs additional files to
complete the OAuth2 token flow:
- `access_tokens.db` — cached short-lived access tokens
- `credentials.db` — credential metadata
- The ability to write back refreshed tokens

Because the ADC file is mounted read-only and the
token databases are not mounted at all, OpenCode inside
the container cannot authenticate to Google Vertex AI.
The Anthropic-via-Vertex integration fails silently.

## What Changes

### Modified Capabilities

- `googleCloudCredentialMounts()` in
  `internal/sandbox/config.go`: Change Strategy 2 from
  mounting a single ADC file read-only to mounting the
  entire `~/.config/gcloud/` directory read-write.

### New Capabilities

None.

### Removed Capabilities

None.

## Impact

- 1 file modified: `internal/sandbox/config.go`
- 1 test file updated: `internal/sandbox/sandbox_test.go`
- No new Go logic — changing a single-file mount to a
  directory mount

## Constitution Alignment

All N/A — this is a volume mount path change.
