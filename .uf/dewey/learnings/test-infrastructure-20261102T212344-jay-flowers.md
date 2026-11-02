---
tag: test-infrastructure
author: jay-flowers
category: gotcha
created_at: 2026-11-02T21:23:44Z
identity: test-infrastructure-20261102T212344-jay-flowers
tier: draft
---

Node's execFile callback error object is an implicit runtime contract with two distinct kill shapes: a timeout kill surfaces as error.killed=true with error.signal set and error.code null, while a maxBuffer overflow surfaces as error.code='ERR_CHILD_PROCESS_STDIO_MAXBUFFER' with killed/signal unset (Node 24, Bun 1.4; not 'ENOBUFS'). When production code discriminates failure modes on these shapes (e.g. reporting timed_out vs output-limit), each branch needs its own direct test — a real-process case or a fake that reproduces the exact shape. The shapes are not documented as stable API, and a single mistyped property check silently misclassifies one failure mode as the other without failing any test that only exercises the other branch. In the preflight-declared change, the timeout branch got a real-process round-trip test (sleep + timeout_seconds: 1 → exit 124) but the output-limit branch shipped untested -- and keyed on 'ENOBUFS', a code neither runtime reports, which a real flood test would have exposed at once. Prefer real-process cases (or fakes copied from observed shapes) over shapes assumed from memory.
