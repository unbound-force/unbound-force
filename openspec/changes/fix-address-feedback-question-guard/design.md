## Context

Two runs of `/uf.address-feedback 159` on `unbound-force/dewey#159` (18 items from review `5820364580`) ended mid-walkthrough with no commits, pushes, or replies. Run 1 stopped at item 12 of 18; run 2 around items 7–9. At item 12 the agent emitted the `question` tool invocation as plain pseudo-XML in its reply text instead of making a tool call:

```
<invoke name="question"><parameter name="questions">[{"question": "Item 12 of 18 — [LOW] … Decide:", …}]</parameter></invoke>
```

No question was raised, so neither the auto-question plugin nor a human could answer. The session idle pattern is the same one seen in #853 (question-tool presentation contract across `uf.triage-issue`, `uf.address-feedback`, `uf.review-pr`). The fix must not re-introduce the failure silently.

Per #1013's issue comment, the deterministic repair belongs in `jflowers/opencode-supervisor#65` (detect last-message tool-call-as-text on idle and re-prompt to make the real call) plus `jflowers/gh-hud#9` (flag idle-waiting runs). Those live out-of-repo and are not part of this change. This change is the small in-repo leftover noted in that comment: tighten wording in `.opencode/commands/uf.address-feedback.md` so each per-item decision says plainly to call the `question` tool.

Current wording already says "Present the **question tool**" and carries a visibility directive ("the printed text and the question tool MUST present the same details and options"). The failure shows that phrasing can be read as "print something that looks like a question" rather than "make a tool call."

## Goals / Non-Goals

### Goals
- Make the per-item triage decision in Section 3.2 unambiguously instruct the agent to **call** the `question` tool, with an explicit "do not write `<invoke>` as text" guard.
- Apply the same tightening to the follow-up collection prompts in the Decision table (Modify / Reject / Ask) and to the Section 3.4 triage-summary confirmation and Phase 4 confirmation gates where a question tool call is required.
- Keep the visibility directive (printed text ↔ question tool parity) unchanged.
- Preserve scaffold sync: `.opencode/commands/uf.address-feedback.md` and `internal/scaffold/assets/opencode/commands/uf.address-feedback.md` stay byte-identical.

### Non-Goals
- Supervisor idle-detection / re-prompt logic (`opencode-supervisor#65`) — out-of-repo.
- gh-hud idle-flagging (`gh-hud#9`) — out-of-repo.
- Any change to the four-phase workflow, dispatch, caching (`state.json`), posting, or `feedback-triage` schema.
- A new deterministic tool or artifact produced by this change.

## Decisions

### D1: Wording-only in-repo fix; deterministic repair out-of-repo

Do not add a tool or state machine to `uf.address-feedback.md` for this failure mode. The existing wording is the only in-repo surface to tighten; the real recovery (parse last message for `<invoke name="question">` text and force a tool call) must run after the turn has ended, which a command template cannot do. Placing that logic in `opencode-supervisor` keeps a single guard that covers every command that asks questions.

**Alternatives considered**: (1) Add a post-step loop inside the command that re-checks "did I call the tool?" — rejected as brittle prompt-level retry that still cannot fire after the turn ended. (2) Duplicate detector text inside the command — rejected; detector belongs in the supervisor where it can run on idle for any command.

### D2: Imperative phrasing + negative instruction

Change "Present the **question tool**" → "Call the `question` tool … — do not write the call as text/`<invoke>` in your reply" at each decision site. Retain the existing visibility directive paragraph immediately after. The positive imperative addresses tool-call omission; the negative instruction directly targets the observed pseudo-XML emission (`<invoke name="question">…` in reply text). Both are needed — positive alone was already present in paraphrase and still failed.

### D3: Scope to observed gates only

Touch Section 3.2 (initial decision), the Decision-table follow-ups, Section 3.4 confirmation, and Phase 4 gates. Do not rewrite the entire file or refactor the presentation contract — that contract is tracked in #853 and should be changed once consistently if it is changed.

## Risks / Trade-offs

- **Wording may still be ignored by some models**: A prose tightening is inherently best-effort. Accepted: the supervisor is the backstop and the risk is already mitigated there. This change only reduces likelihood, not eliminates it.
- **Over-tightening could introduce redundant repetition**: Repeating the "do not write as text" clause at every gate could be noisy. Accepted: repetition at the actual call sites is cheap relative to a silent stall that loses up to 12 Tier 2 panel outcomes and blocks a PR.
- **Scaffold drift**: Two copies must stay in sync. Mitigated by existing `TestEmbeddedAssets_MatchSource` / `TestScaffoldDrift` checks — no new test harness required for this wording change.

## Constitution Alignment Reference

Per proposal: I N/A, II N/A, III N/A, IV N/A, V N/A, VI PASS — the deterministic procedure lives in plugin code (opencode-supervisor) rather than prose, satisfying Principle VI; this repo change is intentionally minimal wording.
