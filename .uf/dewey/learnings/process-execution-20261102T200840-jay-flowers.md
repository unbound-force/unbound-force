---
tag: process-execution
author: jay-flowers
category: gotcha
created_at: 2026-11-02T20:08:40Z
identity: process-execution-20261102T200840-jay-flowers
tier: draft
---

In child_process.execFile, a timeout kill and a maxBuffer overflow have different callback-error shapes: a timeout kill sets killed=true, signal='SIGTERM', code=null; a maxBuffer overflow sets code='ERR_CHILD_PROCESS_STDIO_MAXBUFFER' and leaves killed/signal unset (observed on Node 24.11 and Bun 1.4.4). It is NOT 'ENOBUFS' -- a review-loop fix in uf-workflow's runProcessNoShell (uf_preflight_declared_checks) branched on err.code === 'ENOBUFS', which never matches; the original killed+signal timeout check was already correct for overflow. Verify error shapes empirically on every target runtime before encoding them, and classify overflow by ERR_CHILD_PROCESS_STDIO_MAXBUFFER (see processResultFromExecFile, pinned by a table test of observed shapes plus a real flood round trip with a small maxBuffer).
