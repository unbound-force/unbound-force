## 1. /uf-init Customization

- [x] 1.1 Add "OpenSpec Command Guardrails" section to
  `.opencode/command/uf-init.md` — instructions for
  the AI agent to inject a `## Guardrails` section
  into `.opencode/command/opsx-propose.md` if not
  already present. Idempotent (check for existing
  section before appending).
- [x] 1.2 Copy updated `.opencode/command/uf-init.md`
  to `internal/scaffold/assets/opencode/command/`

## 2. Verification

- [x] 2.1 Run `go build ./...` to verify clean
  compilation
- [x] 2.2 Run `go test -race -count=1
  -run TestEmbeddedAssets_MatchSource
  ./internal/scaffold/...` to verify drift detection
