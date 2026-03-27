## Context

`uf init` (Go binary) scaffolds `.opencode/command/uf-init.md`.
The user then runs `/uf-init` in an OpenCode session to
apply project-specific customizations to third-party
tool files (OpenSpec skills and commands).

## Goals / Non-Goals

### Goals
- Single `/uf-init` command applies all customizations
- Idempotent -- safe to run multiple times
- Resilient to OpenSpec CLI updates (LLM reasons about
  file structure, not hardcoded line numbers)
- Clear error reporting with fix suggestions
- Reports what was found, applied, skipped, and errors

### Non-Goals
- Replacing `uf init` (Go binary)
- Customizing non-OpenSpec tool files (v1)
- Providing an undo mechanism
- Modifying Go code or running tests

## Decisions

### Command File Structure

The command file has 4 sections:

1. **Prerequisites**: Check `.opencode/` exists (uf init
   ran). Check each target file exists. Error with fix
   suggestion for missing files.

2. **Customization 1: Branch Enforcement**: For each of
   7 target files, check if branch content is present,
   insert if not.

3. **Customization 2: Dewey Context**: For each of 5
   target files, check if Dewey query instructions are
   present, insert if not.

4. **Customization 3: 3-Tier Degradation**: For each of
   3 target skill files, check if degradation pattern
   is present, insert if not.

### Target Files and Customizations Matrix

| Target File | Branch | Dewey | Degradation |
|-------------|:------:|:-----:|:-----------:|
| `skills/openspec-propose/SKILL.md` | Create branch | Query before proposal | 3-tier |
| `skills/openspec-apply-change/SKILL.md` | Validate branch | Query before impl | 3-tier |
| `skills/openspec-archive-change/SKILL.md` | Cleanup branch | -- | -- |
| `skills/openspec-explore/SKILL.md` | -- | Query for investigation | 3-tier |
| `command/opsx-propose.md` | Create branch | Query before proposal | -- |
| `command/opsx-apply.md` | Validate branch | Query before impl | -- |
| `command/opsx-archive.md` | Cleanup branch | -- | -- |

Total: 14 potential insertions across 7 files.

### Idempotency Strategy

For each customization, the LLM checks semantically:
- **Branch**: "Does this file already describe creating/
  validating/cleaning up an `opsx/` branch?"
- **Dewey**: "Does this file already reference
  `dewey_semantic_search` or Dewey context queries?"
- **Degradation**: "Does this file already describe a
  3-tier degradation pattern (Full Dewey, Graph-only,
  No Dewey)?"

If the concept is already present (even with different
wording), skip with `⊘ already present`.

### Error Handling

If a target file doesn't exist:
- Report: `❌ <path>: file not found`
- Explain: "This file should have been created by
  `openspec init` which runs as part of `uf init`."
- Suggest: "Run `uf setup` to install OpenSpec, then
  `uf init` to scaffold files, then re-run `/uf-init`."
- Continue with other files (don't stop entirely).

### Output Format

```
## /uf-init: Project Customizations

### Prerequisites
  ✅ .opencode/ exists
  ✅ OpenSpec skills found (4 files)
  ✅ OpenSpec commands found (3 files)

### Branch Enforcement
  ✅ openspec-propose/SKILL.md: inserted
  ⊘ opsx-propose.md: already present (skipped)
  ✅ openspec-apply-change/SKILL.md: inserted
  ⊘ opsx-apply.md: already present (skipped)
  ...

### Dewey Context
  ✅ openspec-propose/SKILL.md: inserted
  ...

### 3-Tier Degradation (Skills only)
  ✅ openspec-propose/SKILL.md: inserted
  ...

### Summary
Applied: 8 | Already present: 4 | Errors: 0
```

### Scaffold Asset

The command file is deployed by `uf init` as a
tool-owned scaffold asset at
`internal/scaffold/assets/opencode/command/uf-init.md`.
This means it's overwritten on re-scaffold, so
updates to the customization instructions flow to
users on the next `uf init`.

## Risks / Trade-offs

**LLM reliability**: The command relies on the LLM
correctly identifying insertion points. Mitigation:
the customization descriptions are specific and the
idempotency check prevents double-insertion.

**OpenSpec updates**: When the user updates the
OpenSpec CLI, skill files are overwritten. The user
must re-run `/uf-init`. The command's description
notes this: "Run after `uf init`, `uf setup`, or
updating the OpenSpec CLI."

**Command file size**: ~250-300 lines of Markdown.
This is consistent with other large command files
(`speckit.implement.md`, `review-council.md`).
