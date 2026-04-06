## 1. Governance Documents

- [x] 1.1 Add "Gatekeeping Integrity" bullet to Development
  Workflow section in `.specify/memory/constitution.md`
- [x] 1.2 Add "Gatekeeping Value Protection" subsection to
  Behavioral Constraints in `AGENTS.md` with protected
  value categories and "what to do instead" instruction

## 2. Divisor Review Agents

- [x] 2.1 Add "Gatekeeping Integrity" checklist item to
  Code Review Mode audit section in
  `.opencode/agents/divisor-guard.md`
- [x] 2.2 Add "Gatekeeping Integrity" checklist item to
  Spec Review Mode audit section in
  `.opencode/agents/divisor-guard.md`
- [x] 2.3 Add "Gate Tampering" checklist item to
  Security Checks section in
  `.opencode/agents/divisor-adversary.md`

## 3. Developer Agent

- [x] 3.1 Add "Gatekeeping Integrity" behavioral
  constraint to `.opencode/agents/cobalt-crush-dev.md`

## 4. Scaffold Asset Sync

- [x] 4.1 Copy updated `divisor-guard.md` to
  `internal/scaffold/assets/opencode/agents/divisor-guard.md`
- [x] 4.2 Copy updated `divisor-adversary.md` to
  `internal/scaffold/assets/opencode/agents/divisor-adversary.md`
- [x] 4.3 Copy updated `cobalt-crush-dev.md` to
  `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md`

## 5. Documentation

- [x] 5.1 Add Recent Changes entry to `AGENTS.md`

## 6. Verification

- [x] 6.1 Run `go test -race -count=1 ./internal/scaffold/...`
  to verify drift detection passes
- [x] 6.2 Run `go build ./...` to confirm clean compilation
- [x] 6.3 Verify constitution alignment: gatekeeping
  integrity strengthens Observable Quality (protects
  metrics from tampering) and Autonomous Collaboration
  (preserves artifact quality standards)
