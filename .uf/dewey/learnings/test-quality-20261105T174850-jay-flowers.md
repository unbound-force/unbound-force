---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:48:50Z
identity: test-quality-20261105T174850-jay-flowers
tier: draft
---

In-loop stream bounds (TOCTOU read guards that cap a live stream while stat stays small) are deterministically testable by mocking the file-open boundary: vi.mock("node:fs/promises") replaces open with a synthetic FileHandle whose stat() reports a small size while read() serves fixed-size chunks indefinitely. Asserting the exact read count at which the guard must fire (e.g. 1025 for a 64 MiB cap at 64 KiB per read — 1024 full reads land exactly on the cap, the 1025th trips the strictly-greater check) proves the in-loop branch fired rather than the pre-read stat branch, and bounding the mock's total read budget (e.g. 2048 calls) keeps a removed guard from hanging the suite — the subsequent JSON-parse failure then fails the message assertion instead. Shipped example: .opencode/test/read-dispatch-artifact-stream-guard.test.ts for readDispatchArtifact (RD-FR-057).
