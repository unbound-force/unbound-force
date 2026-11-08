<!--
  Retroactive change: the system shipped on PR #959's branch
  before this change was filed; tasks are recorded as completed.
  "Round 2" marks work done while addressing PR #959 review
  feedback.
-->

## 1. Persona

- [x] 1.1 Add .opencode/agents/divisor-determinism.md owning
  DE-001..DE-008 with graceful degradation and treat-as-data
  directives
- [x] 1.2 Enroll in .uf/reviewer-capabilities.yaml (scopes:
  documentation, standard)
- [x] 1.3 Bash permissions: wrapper and tool-inventory allows
  behind terminal metacharacter denies (round 2)
- [x] 1.4 Untrusted-input marking in Review Scope (round 2)

## 2. Evidence tooling

- [x] 2.1 .opencode/lib/tool-inventory.ts: inventory +
  --match + --seams --base (DE-007), JSON provenance stamp,
  text-mode control-character escaping (round 2)
- [x] 2.2 .opencode/lib/cli-common.ts: isMainModule entry
  detection, runMain crash-to-exit-2, escapeControlChars,
  null-safe spawnDetail (round 2)
- [x] 2.3 instruction-lint IL-042 (DE-008 delegation
  contract) and --base/--changed deltas (DE-004)
- [x] 2.4 Real-repo discoverability test pinning seam credit
  against the live registration-surface suite (round 2)

## 3. Re-scopes

- [x] 3.1 divisor-linguist defers DE findings and IL-040..042
  hints; keeps lint evidence and instruction clarity
- [x] 3.2 divisor-guard defers DE-008 to the enrolled
  determinism persona, deployment-neutral (canonical +
  scaffold mirror)
- [x] 3.3 divisor-curator bash allows hardened behind
  terminal denies (canonical + scaffold mirror) (round 2)

## 4. Docs and governance

- [x] 4.1 CHANGELOG entry with Spec path; AGENTS.md agent
  count
- [x] 4.2 docs/heroes.md + docs/usage.md 11-persona council
  notes; QUICKSTART persona counts (round 2)
- [x] 4.3 This retroactive change record (round 2 review
  feedback)
