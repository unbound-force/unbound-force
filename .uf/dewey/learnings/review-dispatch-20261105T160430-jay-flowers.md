---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:04:30Z
identity: review-dispatch-20261105T160430-jay-flowers
tier: draft
---

Race-window guards that re-check a bound inside a read loop (stat-then-read TOCTOU defense) cannot be tested deterministically with real files: growth before the call is visible to the pre-read stat, and concurrent mid-read appends race the drain loop and flake. The deterministic equivalent is a synthetic file handle behind a module mock: vi.mock("node:fs/promises", importOriginal) overriding only open (the unleash-* suite's node:fs mock precedent), with stat reporting a small regular file and read reporting fixed bytesRead per call until the accumulated total crosses the cap (1025 x 64 KiB reads for a 64 MiB cap, ~64 MiB transient buffer memory, sub-second). Assert the error code plus evidence the stat stayed under the cap (mock-reported size or read call count >= 1025), proving the in-loop branch — not the stat branch — produced the rejection; that assertion is what makes the test fail when the guard is deleted, turning it into genuine regression protection rather than a characterization test. Keep the mocked test in its own file so the file-level vi.mock leaves real-fs tests untouched.
