---
tag: reviewer-registration
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:07:22Z
identity: reviewer-registration-20261107T160722-jay-flowers
tier: draft
---

In unbound-force, adding a repo-local reviewer persona (one that depends on repo-local tooling, e.g. divisor-linguist on the instruction linter) has a deliberately split registration surface: register it in the repository's .uf/reviewer-capabilities.yaml extension copy (validated by TestReviewerCapabilities_RepositoryManifestExtendsCanonical, which requires every canonical entry unchanged plus an agent definition for each addition), bump the AGENTS.md agent count, and add the agent file to knownNonEmbeddedFiles in scaffold_test.go. Do NOT add it to the scaffolded command rosters (uf.review-council.md's "Known Divisor Persona Roles" table, uf.triage-issue.md's "six known review-capable personas", uf.address-feedback.md's known-review list) or to .uf/review-matrix.yaml's advisor membership lists: those files are embedded scaffold mirrors (byte-identity enforced) shipped to every hero repo, which never receives the repo-local agent file — a roster row there would dangle in generated projects. The commands' escape hatches ("Additional personas may run only when the validated reviewer manifest and returned plan classify them as review-capable and include them") are the designed absorption mechanism for repo-local personas, and prompts are assembled from the agent definition itself, so dispatch quality is unaffected. Verified on PR #958: a reviewer initially reading the stale "six known personas" text as a registration gap should check whether the roster file is a scaffolded mirror before flagging.
