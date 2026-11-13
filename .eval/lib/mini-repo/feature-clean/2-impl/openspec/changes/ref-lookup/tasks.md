## 1. Ref resolution

- [x] 1.1 Add `parseObjectName` and `gitEnv` with table tests
  (RR-FR-002, RR-FR-004)
- [x] 1.2 Add `refs.Lookup` using argv-form git with
  `--end-of-options` (RR-FR-001, RR-FR-003, RR-FR-005)
- [x] 1.3 Test Lookup against temporary repositories for every
  scenario in specs/ref-resolution/spec.md

## 2. Provenance

- [x] 2.1 Change `provenance.Stamp` to `(dir, ref) (Record,
  error)` and add `Record.Commit` (AP-FR-001, AP-FR-002)
- [x] 2.2 Test Stamp success, failure and JSON shape (AP-FR-003)

## 3. Verification

- [x] 3.1 Run CI locally: `go vet ./...` and
  `go test -race -count=1 ./...`
- [x] 3.2 Verify constitution alignment against proposal.md
  (I-VI) and that no coverage threshold changed
- [x] 3.3 Documentation gate: AGENTS.md already lists
  `internal/refs` and `internal/provenance`; the change is an
  internal package change, so no documentation issue is needed
