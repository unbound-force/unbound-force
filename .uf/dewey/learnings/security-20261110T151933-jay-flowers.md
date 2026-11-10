---
tag: security
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:19:33Z
identity: security-20261110T151933-jay-flowers
tier: draft
---

Reference implementation of the shell-free gh subprocess pattern for intent-level workflow tools (unbound-force PR #1007, .opencode/lib/feedback-replies.ts): (1) injectable executor typed as (args: string[], cwd: string) => string backed by execFileSync("gh", args, { timeout }) — argument array, never a shell, so untrusted text can never become shell code; (2) every untrusted body reaches gh only through a 0o600 file inside a mkdtemp 0o700 directory, consumed via --input/--body-file and removed in a finally block; (3) scalar inputs that must reach argv (repo, title) are constrained by zod strict schemas (regex ^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$, length caps, non-blank) at the tool boundary with safeParse re-validation inside execute; (4) gh failures classified into a structured error contract (gh_not_found/timeout/rate_limited/gh_auth_failed/api_error with retryable flags) by inspecting err.code, err.status === 4, stderr text; (5) per-item side effects persisted to a 0o600 state file (atomic temp+rename, symlinked targets refused) before the next mutation so re-runs resume. This retires the older prose pattern of `gh ... --title "$(cat "$TITLE_FILE")"` which re-parsed file contents through a shell.
