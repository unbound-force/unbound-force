# Quickstart: Unleash Command

## Prerequisites

1. A Speckit feature branch exists (`NNN-*`)
2. `spec.md` exists in the feature directory
3. OpenCode is running with the Swarm plugin

## Happy Path Verification

```bash
# 1. Create a spec (human step)
/speckit.specify "Add a health check endpoint"

# 2. Unleash the swarm
/unleash

# Expected output sequence:
#   Step 1/8: Clarifying spec...
#     Dewey answered 2/2 questions automatically.
#   Step 2/8: Generating plan...
#     Created plan.md
#   Step 3/8: Generating tasks...
#     Created tasks.md (15 tasks, 8 phases)
#   Step 4/8: Reviewing specs...
#     Review council: APPROVE (5/5 reviewers)
#   Step 5/8: Implementing...
#     Phase 1: Setup (3 tasks, sequential)
#     Phase 2: Foundational (2 tasks, sequential)
#     Phase 3: US1 (4 tasks, parallel)  ← workers
#     Phase 4: Polish (2 tasks)
#   Step 6/8: Reviewing code...
#     CI: PASS | Gaze: quality report | Divisor: APPROVE
#   Step 7/8: Retrospective...
#     Stored 3 learnings to semantic memory
#   Step 8/8: Demo instructions
#
#   ## What Was Built
#   [summary from spec user stories]
#
#   ## How to Verify
#   [verification commands]
#
#   ## Next Steps
#   - /finale to merge and release
#   - /speckit.clarify to refine and re-run /unleash
```

## Exit Path Verification

```bash
# 1. Create a spec with an unanswerable question
/speckit.specify "Add OAuth2 with [company-specific IdP]"

# 2. Unleash
/unleash

# Expected: exits at clarify step
#   ## /unleash paused at: clarify
#   **Reason**: 1 question requires human input
#   Q1: Which identity provider should be used?
#   ### What to do next
#   Answer the question in the spec, then re-run /unleash.

# 3. Answer the question in spec.md
# 4. Re-run /unleash -- should resume at plan step
```

## Resume Verification

```bash
# 1. Run /unleash on a spec that already has plan.md
/unleash

# Expected: skips clarify and plan, resumes at tasks
#   Detected: clarify ✓ plan ✓ tasks ✗
#   Resuming at step 3/8: Generating tasks...
```
