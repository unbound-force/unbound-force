---
description: Run the five-reviewer governance council to audit codebase or spec compliance.
---
<!-- scaffolded by gaze v1.2.7 -->

# Command: /review-council

## User Input

```text
$ARGUMENTS
```

## Description

Review the current codebase **or** SpecKit artifacts for compliance with the Behavioral Constraints in `AGENTS.md` using the review council (The Adversary, The Architect, The Guard, The Tester, The Operator).

## Determine Review Mode

Inspect `$ARGUMENTS` to select the review mode:

- If arguments contain the word **"specs"**: use **Spec Review Mode**
- Otherwise: use **Code Review Mode** (default)

---

## Code Review Mode

Review the current codebase for compliance with the Behavioral Constraints in `AGENTS.md`.

### Instructions

1. Delegate the review to all five council agents in parallel using the Task tool:
   - `reviewer-adversary` — audits for security, resilience, efficiency, and constraint violations
   - `reviewer-architect` — audits for architectural alignment, coding conventions, and plan adherence
   - `reviewer-guard` — audits for intent drift, neighborhood impact, and zero-waste compliance
   - `reviewer-testing` — audits for test architecture, coverage strategy, assertion quality, and testing convention compliance
   - `reviewer-sre` — audits for deployment readiness, release pipeline integrity, dependency health, operational observability, and upgrade paths

   For each agent, instruct it to review the current changes and return its verdict (**APPROVE** or **REQUEST CHANGES**) along with all findings.

2. Collect all **REQUEST CHANGES** findings from the five reviewers. If all five return **APPROVE**, report the result and stop.

3. If there are **REQUEST CHANGES**, address the findings by making the necessary code fixes. Then re-run all five reviewers to verify the fixes. Repeat this loop until all five return **APPROVE** or the process has exceeded 3 iterations.

4. If 3 iterations are exceeded, ask the user whether to continue or stop.

5. Provide a final report to the user:
   - What was found in each iteration
   - What was fixed
   - If stopped early, the current set of outstanding **REQUEST CHANGES**
   - If there were persistent circular **REQUEST CHANGES** (fixes for one reviewer cause failures in another), report those with additional detail so the user can make an informed decision

---

## Spec Review Mode

Review all SpecKit artifacts under `specs/` for quality, consistency, and alignment with the project constitution.

### Instructions

1. Delegate the review to all five council agents in parallel using the Task tool:
   - `reviewer-adversary` — audits specs for completeness, testability, ambiguity, security gaps, dependency risks, and cross-spec consistency
   - `reviewer-architect` — audits specs for structural consistency, plan-to-spec alignment, task coverage, data model coherence, tech stack feasibility, and research quality
   - `reviewer-guard` — audits specs for intent fidelity, scope creep, inter-feature conflicts, status accuracy, user value, and constitution alignment
   - `reviewer-testing` — audits specs for testability of requirements, coverage strategy definition, fixture feasibility, and contract surface clarity
   - `reviewer-sre` — audits specs for deployment feasibility, operational requirements, configuration management, dependency risks, maintenance burden, and cross-hero operational impact

   For each agent, instruct it to **operate in Spec Review Mode**: review all SpecKit artifacts under `specs/` (not code), plus `.specify/memory/constitution.md` and `AGENTS.md`. Instruct the agent to return its verdict (**APPROVE** or **REQUEST CHANGES**) along with all findings.

2. Collect all **REQUEST CHANGES** findings from the five reviewers. If all five return **APPROVE**, report the result and stop.

3. If there are **REQUEST CHANGES**, apply the **hybrid fix policy**:

   **Auto-fix (LOW and MEDIUM findings)** — Apply these fixes directly to the spec files:
   - Formatting and template compliance issues
   - Status field updates (e.g., "Draft" on a completed feature)
   - Terminology inconsistencies (same concept named differently across specs)
   - Missing or stale cross-references between spec, plan, and tasks
   - Coverage gaps with obvious fixes (e.g., a requirement with zero tasks when the task is clearly implied by the plan)
   - Stale or incorrect metadata (dates, branch names, prerequisite lists)

   **Report only (HIGH and CRITICAL findings)** — Do NOT attempt to fix these. Report them with full context and recommendations so the user can make an informed decision:
   - Missing user stories or acceptance criteria
   - Scope creep or under-specification
   - Design-level security gaps or unaddressed failure modes
   - Inter-feature conflicts or architectural misalignment
   - Constitution violations
   - Ambiguous requirements that require human judgment to resolve

4. After applying LOW/MEDIUM fixes, re-run all five reviewers to verify. Repeat this loop until all five return **APPROVE** (considering only remaining HIGH/CRITICAL findings as blocking) or the process has exceeded 3 iterations.

5. If 3 iterations are exceeded, ask the user whether to continue or stop.

6. Provide a final report to the user:
   - What was found in each iteration
   - What was auto-fixed (LOW/MEDIUM)
   - Outstanding HIGH/CRITICAL findings that require human decision, with full context and recommendations
   - The Architect's Alignment Score for spec quality (if provided)
   - If there were persistent circular findings, report those with additional detail
   - Suggested next steps (e.g., "Run `/speckit.clarify` on spec 007 to resolve the ambiguous credential migration behavior")

---

## Verdict

The council returns **APPROVE** only when all five reviewers return **APPROVE**. Any single **REQUEST CHANGES** means the council verdict is **REQUEST CHANGES**.

In Spec Review Mode, the council may return **APPROVE WITH ADVISORIES** when all LOW/MEDIUM findings have been auto-fixed but HIGH/CRITICAL findings remain that require human judgment. The advisories are the outstanding HIGH/CRITICAL findings.
