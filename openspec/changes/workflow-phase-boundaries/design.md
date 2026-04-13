## Context

The Speckit pipeline has 8 phases that run in sequence:
specify → clarify → plan → tasks → analyze → checklist
→ implement → (review via /unleash). Each phase has a
corresponding slash command.

Spec 027 externalized Speckit scripts, templates, and
config from `uf`'s scaffold assets to `specify init`.
But the 9 `.opencode/command/speckit.*.md` command files
were left embedded in `uf`. These are OpenCode command
definitions that tell the agent what to do when someone
types `/speckit.specify`, etc. They should follow the
same externalization pattern — `specify init` creates
the upstream 5, `/uf-init` creates the UF-custom 4 and
applies project-specific customizations to all 9.

The `/unleash` command has a resumability detection
system that probes filesystem state. Spec review writes
a marker but code review has no equivalent.

Additionally, 74 stray files exist in subdirectories
from tools running in the wrong working directory.

## Goals / Non-Goals

### Goals

- Externalize 9 speckit command files from uf scaffold
- Add `/uf-init` guardrail injection for speckit commands
- Add `/uf-init` UF-custom command creation
- Add code-review marker to /unleash
- Add phase boundary rules to constitution + AGENTS.md
- Clean up 74 stray files + prevent recurrence
- Keep all changes in Markdown + Go test assertions

### Non-Goals

- No programmatic enforcement of phase boundaries
- No changes to OpenSpec tactical commands (opsx-*)
- No changes to the 6 UF-custom commands that stay
  embedded (cobalt-crush, constitution-check, finale,
  review-council, uf-init, unleash)

## Decisions

### D1: Externalize all 9 speckit.*.md files

Remove from `internal/scaffold/assets/opencode/command/`:
speckit.specify.md, speckit.clarify.md, speckit.plan.md,
speckit.tasks.md, speckit.analyze.md, speckit.checklist.md,
speckit.implement.md, speckit.constitution.md,
speckit.taskstoissues.md.

Of these 9, upstream `specify init` creates 5 (specify,
plan, tasks, implement, constitution). The remaining 4
(analyze, checklist, clarify, taskstoissues) are
UF-custom and created by `/uf-init`.

### D2: /uf-init AI-assisted guardrail injection

The `/uf-init` slash command includes instructions for
the AI agent to inject a `## Guardrails` section into
each speckit command file. The pattern is:

1. Read each `.opencode/command/speckit.*.md` file
2. Check if a `## Guardrails` section exists
3. If not present, append the guardrails block
4. If present, skip (idempotent)

The guardrails block states: this command may only
write to files within the `specs/NNN-*/` feature
directory, never source code, tests, or config.

### D3: /uf-init creates 4 UF-custom speckit commands

The `/uf-init` command includes instructions to create
`speckit.analyze.md`, `speckit.checklist.md`,
`speckit.clarify.md`, and `speckit.taskstoissues.md`
if they do not exist. The content for each is defined
inline in the `/uf-init` command file instructions.

### D4: Code-review marker in /unleash

After Step 6 (code review) approval, write:
```
<!-- code-review: passed -->
```
to the end of tasks.md, same location as the
spec-review marker.

Update Step 2 resumability detection to check for this
marker instead of inferring from CI pass/fail.

### D5: Constitution + AGENTS.md phase boundaries

Add Phase Discipline rule to constitution Development
Workflow section. Add Workflow Phase Boundaries
subsection to AGENTS.md Behavioral Constraints.

These are behavioral rules — the same enforcement
model as gatekeeping value protection.

### D6: Stray file prevention

Add `.gitignore` patterns that prevent `.opencode/`
and `.uf/` directories from being tracked inside
subdirectories like `internal/scaffold/` and
`cmd/unbound-force/`. The patterns must not affect
the repo-root `.opencode/` and `.uf/` directories.

Pattern approach: use path-specific ignores:
```gitignore
# Prevent stray tool directories in subdirectories
cmd/**/.opencode/
cmd/**/.uf/
internal/**/.opencode/
internal/**/.uf/
```

### D7: What stays embedded in uf

6 commands remain as uf scaffold assets because no
external tool creates them:

- `cobalt-crush.md` — UF developer persona
- `constitution-check.md` — UF alignment checker
- `finale.md` — UF branch finalization
- `review-council.md` — UF Divisor review system
- `uf-init.md` — UF project customizer
- `unleash.md` — UF autonomous pipeline

These are Unbound Force-specific commands, not Speckit
or OpenSpec commands.

## Risks / Trade-offs

### Risk: /uf-init not run after specify init

If an engineer runs `uf init` + `specify init` but
forgets `/uf-init`, the speckit commands exist without
guardrails. Mitigated by the "When to Re-run" section
in `/uf-init` and the constitution/AGENTS.md rules
(agents read those regardless of guardrails).

### Risk: AI-assisted injection variability

The AI agent may format the guardrails section slightly
differently each run. Mitigated by the idempotency
check — the section is only injected once. Subsequent
runs detect the existing section and skip.

### Trade-off: 4 commands defined inline in /uf-init

The 4 UF-custom commands are defined as instruction
text in `/uf-init`, not as separate template files.
This keeps the change simple but means updating them
requires editing `/uf-init` rather than a standalone
file. Acceptable because these commands change
infrequently.
