## Context

The `googleCloudCredentialMounts()` function in
`config.go` has two strategies:

1. If `GOOGLE_APPLICATION_CREDENTIALS` is set: mount
   that specific file read-only (service account key).
2. If not set: mount
   `~/.config/gcloud/application_default_credentials.json`
   read-only (gcloud ADC fallback).

Strategy 1 works because service account keys are
self-contained (contain a private key). Strategy 2
fails because `authorized_user` ADC credentials need
the token refresh infrastructure (access_tokens.db,
credentials.db) to authenticate.

## Goals / Non-Goals

### Goals

- Mount the entire `~/.config/gcloud/` directory so
  the Google Auth library can read credentials AND
  write refreshed access tokens
- Keep Strategy 1 unchanged (service account key mount
  works correctly)

### Non-Goals

- No changes to Strategy 1 (explicit
  GOOGLE_APPLICATION_CREDENTIALS)
- No changes to CDE backend credential handling

## Decisions

### D1: Mount entire directory, not individual files

Mounting individual files (ADC + credentials.db +
access_tokens.db) is brittle — the auth library may
need additional files (configs, legacy_credentials).
Mounting the directory is simpler and handles all
credential types.

### D2: Read-write mount (no :ro)

The auth library writes to `access_tokens.db` when
refreshing tokens. A read-only mount would prevent
token refresh, causing auth to fail after the cached
token expires (~1 hour).

### D3: Strategy 1 unchanged

Service account keys (Strategy 1) are self-contained
and work with a single-file read-only mount. No change
needed.
