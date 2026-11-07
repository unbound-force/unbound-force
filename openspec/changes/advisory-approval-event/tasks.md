## 1. Command change

- [x] 1.1 Update the Step 8c mapping table in
  `.opencode/commands/uf.review-pr.md` to key advisory verdicts on
  severity.
- [x] 1.2 Add the severity rule and the human-judgment note.
- [x] 1.3 Make the self-authored rule and the gate options follow
  the mapped event.
- [x] 1.4 Mirror the change in
  `internal/scaffold/assets/opencode/commands/uf.review-pr.md`.

## 2. Verification

- [x] 2.1 Run the scaffold drift tests.
- [x] 2.2 Run `openspec validate advisory-approval-event --strict`.
