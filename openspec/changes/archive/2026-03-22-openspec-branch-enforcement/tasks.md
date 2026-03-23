## 1. Update Propose Command and Skill

- [x] 1.1 Add branch creation step to `.opencode/command/opsx-propose.md`: after `openspec new change`, run `git checkout -b opsx/<name>`. Add guard: if already on an `opsx/*` branch, error and stop.
- [x] 1.2 Add same branch creation step to `.opencode/skills/openspec-propose/SKILL.md`

## 2. Update Apply Command and Skill

- [x] 2.1 Add branch validation step to `.opencode/command/opsx-apply.md`: before implementation, run `git rev-parse --abbrev-ref HEAD` and verify it equals `opsx/<change-name>`. If not, error with checkout hint.
- [x] 2.2 Add same branch validation to `.opencode/skills/openspec-apply-change/SKILL.md`

## 3. Update Archive Command and Skill

- [x] 3.1 Add branch cleanup step to `.opencode/command/opsx-archive.md`: after moving to archive, run `git checkout main`.
- [x] 3.2 Add same branch cleanup to `.opencode/skills/openspec-archive-change/SKILL.md`

## 4. Update Cobalt-Crush Command

- [x] 4.1 Add branch validation to the OpenSpec detection path in `.opencode/command/cobalt-crush.md`: when an active OpenSpec change is detected, validate the current branch is `opsx/<change-name>` before delegating.

## 5. Update Documentation

- [x] 5.1 Add OpenSpec branch convention note to `AGENTS.md` in the "Strategic vs Tactical: Boundary Guidelines" section, documenting that OpenSpec changes use `opsx/<name>` branches.

## 6. Verify

- [x] 6.1 Build passes: `go build ./...`
- [x] 6.2 Full test suite passes: `go test -race -count=1 ./...` (scaffold drift test caught unsync'd cobalt-crush.md -- fixed)
