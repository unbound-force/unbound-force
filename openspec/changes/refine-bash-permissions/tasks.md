<!-- spec-review: passed -->
<!-- code-review: passed -->
<!-- Tasks marked [P] can run in parallel -->

## Phase 1: Implementation

- [x] **1.1** Replace the blanket `"gh api*": "ask"` rule
  in `opencode.json` with refined rules per D1-D4:
  `"gh api*": "allow"` followed by 11 mutation patterns
  (4 `-X` methods + 4 `--method` long forms + 3 data
  flags).
  - Files: `opencode.json`

## Phase 2: Verification

- [x] **2.1** Validate JSON syntax.
  - Run: `python3 -m json.tool opencode.json`
- [x] [P] **2.2** Verify rule count: exactly 26 rules
  (1 global default + 13 original ask rules + 1
  `gh api*` allow + 11 gh api mutation ask rules:
  4 `-X` + 4 `--method` + 3 data flags).
  - Files: `opencode.json`
- [x] [P] **2.3** Verify constitution alignment: all
  five principles assessed. Principles I-IV are N/A.
  Principle V (Security by Default) is Aligned — the
  change implements least privilege for bash commands.
  - Files: `proposal.md`
