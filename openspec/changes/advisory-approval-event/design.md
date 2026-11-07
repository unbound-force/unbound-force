## Context

`openspec/changes/review-council-github-posting/design.md` maps
`APPROVE WITH ADVISORIES` to `COMMENT` "because unresolved
HIGH/CRITICAL advisories should not unblock merge". The mapping
keys on the verdict, so advisories of any severity are treated the
same way.

## Goals / Non-Goals

**Goals:**

- Let an advisory verdict count as an approval when nothing in it
  needs a human.
- Keep the original guarantee for HIGH and CRITICAL advisories.

**Non-Goals:**

- Changing how verdicts are computed or how severity is assigned.
- Changing `/uf.review-council`.
- Changing what triggers a feedback run or a re-review.

## Decisions

### Key the mapping on severity as well as the verdict

`APPROVE WITH ADVISORIES` maps to `APPROVE` only when no displayed
finding is CRITICAL or HIGH. This keeps the stated reason for the
original mapping and removes only the cases it did not need.

### Fail towards COMMENT

A finding with no recognized severity counts as HIGH for this
mapping, so missing data cannot produce an approval.

### Options follow the mapped event

The gate's option sets are selected by the mapped event rather than
the verdict, which keeps the self-authored mapping correct too.

## Risks / Trade-offs

- [More PRs reach an approval without a human reading them] ->
  Branch rules still apply, CRITICAL and HIGH advisories still stop
  the approval, and the confirmation gate is unchanged.
- [Severity is assigned by the reviewing agents] -> A finding rated
  too low would no longer hold back an approval. This is the same
  trust the `APPROVE` verdict already places in those agents.
