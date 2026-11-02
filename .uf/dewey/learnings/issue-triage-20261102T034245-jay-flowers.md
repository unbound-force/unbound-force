---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:42:45Z
identity: issue-triage-20261102T034245-jay-flowers
tier: draft
---

Triage of agent-boilerplate skill extraction (unbound-force #660, proposed divisor-common skill) surfaced three delivery-path facts that change the assessment beyond the command-prose consolidation pattern: (1) Divisor agents carry `bash: deny` in their frontmatter, so the shared Review Scope section instructing `git diff`/`git status` is unexecutable by its own carriers and already superseded — the dispatch pipeline (capture_review_diff / fetch_pr_review_context) delivers diffs directly into prompts. A "no behavioral change" extraction would preserve dead, tool-superseded instruction as the mechanism of record; audit extracted boilerplate for executability under the agent's own permission block, not just content equality. (2) build_review_prompts assembles self-contained prompts from agent definitions, convention packs, AGENTS.md, and constitution — it does not read .opencode/skills/ — so an agent-file skill reference resolves only via runtime `skill` tool loads in child sessions (which restore the full boilerplate to each child's context, making token-savings claims file-footprint/DRY-only, not per-dispatch context savings) or via prompt-builder changes (scope expansion beyond the issue). Efficiency claims for agent-file refactors must state their delivery-path basis per DE-004. (3) Divisor agents are scaffolded to hero repos by the UF CLI (per gaze's AGENTS.md), so any new agent-adjacent skill must join the scaffold manifest and drift-detection test coverage or hero repos diverge from the meta repo. Also: the nominally "shared" agent sections are parameterized skeletons (Source Documents varies per agent in AGENTS.md focus, constitution focus, and Dewey query domain), so extraction needs a defined placeholder set, not just one {constraint_label} slot.
