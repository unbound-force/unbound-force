## Proposal: Refine Bash Permissions

PR #477 added a `permission.bash` block to
`opencode.json` gating 14 ask rules plus 1 global
allow default (15 rules total) behind `ask` prompts
for GitHub-mutating, git-write, and destructive
command patterns. PR #499 fixed the root cause of
DCP compression losing guardrails by adding
`<protect>` tags. Issue #531 proposed removing the
entire permission block as redundant, but member
yvonnedevlinrh refined the scope: keep the permission
block for defense-in-depth, but refine `gh api` rules
so read-only operations do not prompt.

### What Changes

Replace the blanket `"gh api*": "ask"` rule with a
refined set: default `gh api*` to `allow`, then
explicitly gate mutating methods in all flag forms
(`-X`, `-X` glued, `--method`, `--method=`, each in
uppercase and lowercase) and data-sending flags (`-f`,
`-F`, `--field`, `--field=`, `--raw-field`,
`--raw-field=`, `--input`) behind `ask`. All other
original mutation guards remain unchanged.

### Capabilities

- Read-only `gh api` calls (implicit or explicit GET)
  execute without prompting
- Mutating `gh api` calls require approval
- All original `gh issue`, `gh pr`, `git push`,
  `git commit`, and `rm` guards preserved

### Impact

- **Go code**: None (config-only change)
- **CI**: None
- **Dependencies**: None
- **Users**: Fewer unnecessary approval prompts during
  agent sessions for read-only API context gathering

### Constitution Alignment

Assessed against the Unbound Force org constitution
(v1.2.0, 5 principles).

#### I. Autonomous Collaboration

**Assessment**: N/A

No artifact communication patterns are affected.

#### II. Composability First

**Assessment**: N/A

No hero dependencies are introduced or modified.

#### III. Observable Quality

**Assessment**: N/A

No machine-parseable outputs are affected. The
permission block is project-level configuration
consumed by OpenCode at runtime.

#### IV. Testability

**Assessment**: N/A

This is a static JSON configuration change. No
testable components are added or modified. Verification
is by inspection and JSON syntax validation, which is
the appropriate approach for config-only changes.
Drift detection conventions apply to embedded assets,
not to runtime config files.

#### V. Security by Default

**Assessment**: Aligned

This change directly implements the least-privilege
sub-principle by gating mutating operations behind
approval prompts while allowing read-only operations.
The `permission.bash` block provides defense-in-depth
alongside the `<protect>` tags from PR #499: if DCP
compression strips `<protect>` tags, the permission
rules remain as a secondary guard. Mutation patterns
cover uppercase and lowercase method names, short
(`-X`), glued (`-XPOST`), long (`--method`), and
equals (`--method=`) forms, as well as data-sending
flags in both short (`-f`, `-F`) and long (`--field`,
`--raw-field`) forms plus `--input`. The allow-default
for read-only `gh api` calls follows least privilege:
read operations are the minimum permission needed for
context gathering.
