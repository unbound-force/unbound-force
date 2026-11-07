---
tag: feedback-triage
author: jay-flowers
category: decision
created_at: 2026-11-07T17:06:16Z
identity: feedback-triage-20261107T170616-jay-flowers
tier: draft
---

Feedback-triage split rule for composite review comments (PR 957 item 4 precedent, twin npm pin): when one feedback item bundles (a) a structural remedy the repo's conventions already prescribe — a drift test asserting the Makefile PLUGIN_TEST_NPM pin and the ci_security.yml osv_plugin_transitive pin match, the TestEmbeddedAssets_MatchSource pattern — with (b) an accepted-risk documentation extension (recording npm@11.6.2 dist.integrity, accepted-risk class A7), split the disposition rather than forcing one scalar verdict over both halves. ACCEPT (a): the seam lives in the change's own added lines, the remedy is precedented and cheap (findProjectRoot-based parity test), and a comment-coupled constant pair has no failing signal on a one-sided bump. Leave (b) to author discretion under the accepted-risk record; if adopted, prefer an enforced or well-anchored hash record (resolution command + re-resolution-on-bump obligation) over a bare comment hash.
