# Design: plugin-lock-advisories

## Context

Issue #991 records eleven advisory findings from the PR #957
round-3 review council (`stop-shipping-plugin-lock`). That change
is still unarchived, so traceability fixes land in its artifacts in
place; this change carries the tests, the doctor fixes, and the
recorded decisions.

Gatekeeping: every edit here either adds enforcement (new sweep
surfaces, a new parity assertion, a new acceptance test) or records
a decision. No workflow file, pin, threshold, or CI flag changes.

## Decisions

### D1. Sweep the scheduled job, and require step parity

`TestPluginTestNpmPin_TwinsCISecurityPin` and
`TestOsvRecipeDoc_PinsMatchCanonicalJob` iterate over both
`ci_security.yml` (canonical) and `ci_scheduled.yml`. Each file must
declare the pin; the scheduled copy must equal the canonical one.

A new test, `TestOsvPluginTransitive_ScheduledStepsMatchCanonical`,
extracts the `osv_plugin_transitive` job's `steps:` block from both
workflows and requires byte equality. The two CI copies are run by
this repository, so unlike the recipe doc they can be held to full
structural parity (flags, step order, action SHAs). The job-level
header (name, comment, permissions) stays free to differ, because
the scheduled job documents why it exists.

Folding both copies into a reusable workflow was considered and
rejected for now: it changes CI structure (a governance surface)
for a gain the parity test already provides.

### D2. Locale-stable tracked-lock classification

`checkTrackedPluginLock` runs `env LC_ALL=C git -C <dir> ls-files
-- .opencode/package-lock.json`. Under `LC_ALL=C`, gettext ignores
`LANGUAGE` and git emits its untranslated messages. The skip
branch fires only when a line of the output starts with
`fatal: not a git repository`; a `TargetDir` that embeds the phrase
appears mid-line in other git errors and no longer matches.

`git rev-parse --is-inside-work-tree` was considered. It still
needs the same classification when it fails (exit 128 covers both
"not a repository" and real faults such as dubious ownership), so
it adds a process without removing the locale dependency.

### D3. Keep the remediation constants as twinned pairs

The four remediation strings (`pluginRuntimeRemediation` and
`pluginInstallRemediation` in `internal/scaffold`, their
`reviewPlugin*Remediation` twins in `internal/doctor`) stay where
they are. A shared internal package would make the compiler the
enforcement, but it would also retire the source-regex assertion in
`TestRemediationConstants_CarryNpm11EscapeHatch`, which is a drift
test; replacing a drift test is a governance-surface change outside
this follow-up's "only stricter" scope. The existing test already
fails closed on any byte divergence and on a missing constant.
Revisit when a third consumer appears.

### D4. GOTOOLCHAIN patch-staleness window

The `GOTOOLCHAIN=go1.27.0` floor fails closed when the osv-scanner
pin raises its Go floor, but it does not move on Go patch security
releases. The accepted staleness window is: until the next
deliberate osv-scanner bump, or until a Go security release names a
fix that affects building or running the scanner, whichever comes
first. Rationale: the toolchain only compiles a scanner inside an
ephemeral `contents: read` runner, nothing it builds is shipped, and
module downloads are verified through `sum.golang.org`. No
scheduled tripwire is added; the window is documented in
`docs/osv-transitive-scan.md` beside the pin.

### D5. Record the recipe sweep's scope

`TestOsvRecipeDoc_PinsMatchCanonicalJob` checks version pins only.
Flags, permissions, and step structure in the recipe doc are not
compared, because the doc's comments and guidance deliberately
differ from the CI copies and hero repos resolve their own action
SHAs. The scope is stated in the test's doc comment and in the
recipe's Gotchas. The two CI copies get full step parity (D1).

## Risks / Trade-offs

- **`env` on PATH.** The tracked-lock check now depends on `env`.
  It is POSIX on macOS and Linux, the only platforms doctor runs
  on. If it is missing, the check warns with the exec error rather
  than passing silently.
- **Step-parity strictness.** A legitimate edit to one CI copy now
  fails until the other is edited too. That is the intended
  coupling.
