## 1. Implementation

- [x] 1.1 In `internal/sandbox/config.go`, modify
  Strategy 2 of `googleCloudCredentialMounts()`:
  change from mounting single ADC file read-only to
  mounting entire `~/.config/gcloud/` directory
  read-write. Keep Strategy 1 (service account key)
  unchanged.

## 2. Tests

- [x] 2.1 Update tests in
  `internal/sandbox/sandbox_test.go` that check for
  the ADC file mount — verify they now check for the
  directory mount instead

## 3. Verification

- [x] 3.1 Run `go build ./...` and
  `go test -race -count=1 ./internal/sandbox/...`
