## Context

gh-hud dispatches `/uf.review-pr` on the author's PRs and
`/uf.address-feedback` when a PR has changes requested or unresolved
threads. Branch rules dismiss stale approvals, so every push needs a
fresh review. PR #957 showed how the two commands feed each other:
each round of fixes produced a push, each push a full review, and
each full review new findings, mostly LOW.

gh-hud now also caps auto-dispatch at 2 rounds per task per PR. That
stops the loop; this change makes the rounds that do run converge.

## Goals / Non-Goals

**Goals:**

- A feedback run changes code only where a finding is serious enough
  to block merge.
- A re-review that follows a small fix push can approve without a
  full council.
- One council per review and feedback cycle.
- No learnings commits on PR branches.

**Non-Goals:**

- Changing how severity is assigned, or the council itself.
- Changing `/uf.review-council` or `/uf.unleash` Step 6.
- Changing gh-hud. Its learnings PR and round limit already exist.

## Decisions

### Severity decides what changes code

MEDIUM and LOW items get a different option set (`File issue`,
`Reply`, `Ask`) with no code-change option. A finding with no
severity label counts as HIGH, so human comments keep the existing
authority matrix. MEDIUM and LOW items always stay at Tier 1: a
Divisor dispatch to assess a change that will not be made is wasted.

FILE and REPLY threads are resolved along with accepted ones. Left
open, gh-hud reads them as unresolved feedback and starts another
feedback run.

The feedback-triage schema has no FILE or REPLY decision. Rather than
bump the schema, those items are recorded as `reject` (no code change
in this PR) with a `decision_reasoning` prefix of `filed:` or
`reply:`.

### Drop the inner council, keep the outer one

The council inside `/uf.address-feedback` ran before the push; the
re-review ran after it. Keeping the re-review keeps the council
where the approval is decided, and the evaluator makes it cheaper
when the push is small. The feedback run keeps a build, test, and
lint gate before pushing, capped at 3 attempts.

### Light check conditions are conservative

The light check runs only when every condition holds: the reviewed
commit is still in the branch, the delta is at most 200 changed
lines, no file was added, removed, or renamed, no workflow,
`AGENTS.md`, constitution, convention pack, or auth/secret path
changed, the prior review raised no CRITICAL finding, and CI shows
no PR-caused failure. Anything else gets the council. A rebase
always gets the council, since the compare status is no longer
`ahead`.

The light check approves only when every earlier CRITICAL and HIGH
finding is resolved and local tools pass. It never decides a new
CRITICAL or HIGH finding by itself: it escalates to the council.

### Learnings go through the learnings PR

`uf_finale_stage` stages with a pathspec that excludes
`.uf/dewey/learnings` and then unstages anything already staged
there. Compiled articles still ride the feature PR: gh-hud's
learnings PR only collects `learnings/`.

## Risks / Trade-offs

- [A MEDIUM finding that should have been fixed is filed instead] ->
  The issue keeps it tracked, and the author can still raise its
  severity by replying; a human comment with no label is HIGH.
- [The light check misses a problem a council would find] -> It
  only runs on small, contained pushes to code a council already
  reviewed, and a rebase or sensitive path always gets the council.
- [The 200-line threshold and path list are judgment calls written
  as prose] -> A deterministic tool for the evaluator is a natural
  follow-up under the Determinism gate.
- [Without gh-hud, learnings stay untracked until someone opens a
  learnings PR] -> `AGENTS.md` says to open one by hand.
