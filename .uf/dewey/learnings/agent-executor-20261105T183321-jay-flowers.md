---
tag: agent-executor
author: jay-flowers
category: pattern
created_at: 2026-11-05T18:33:21Z
identity: agent-executor-20261105T183321-jay-flowers
tier: draft
---

When several async abort sources can end the same run (wall-clock timer, idle watchdog, parent cancellation), record the abort cause *before* triggering the abort and make the first cause win: guard late timers with `if (abortCause === null)`, and re-check the cause plus a `settled` flag after every `await` inside watchdog callbacks before acting on results. A probe that resolves after the child's prompt returned must neither abort a finished child nor relabel its result. (From divisor-adversary review of PR #950 delegation-timeout-budgets; implemented in executeAgentSession.)
