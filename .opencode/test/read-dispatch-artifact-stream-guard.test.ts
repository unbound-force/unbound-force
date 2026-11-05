import { describe, expect, it, vi } from "vitest"

// RD-FR-057: the reader bounds the LIVE read stream against the 64 MiB cap,
// not just the pre-read stat. The stat branch (sparse-truncate before the
// call) is covered by read-dispatch-artifact.test.ts; this file exercises
// the in-loop TOCTOU guard deterministically by serving a synthetic
// FileHandle whose stat stays small while its read stream grows past the
// cap. A real fixture that grows mid-read races the stat check, so the
// handle is mocked at the module boundary (node:fs/promises) instead.

const openMock = vi.hoisted(() => vi.fn())

vi.mock("node:fs/promises", async (importOriginal) => {
  const actual = await importOriginal<typeof import("node:fs/promises")>()
  return { ...actual, open: openMock }
})

import { readDispatchArtifact } from "../plugins/review-dispatch/index.js"
import samplePayload from "../../schemas/review-dispatch/samples/sample-review-dispatch.json"

const CORRELATION = samplePayload.correlation_id
const MAX_BYTES = 64 * 1024 * 1024

describe("read_dispatch_artifact in-loop size guard (RD-FR-057)", () => {
  it("rejects a stream that grows past the 64 MiB cap mid-read while stat stays small", async () => {
    const statSize = 4096 // far below the cap: the pre-read stat check cannot fire
    let reads = 0
    openMock.mockImplementationOnce(async () => ({
      stat: async () => ({ isFile: () => true, size: statSize }),
      read: async (buffer: Buffer) => {
        reads += 1
        // Serve 64 KiB per call. 2048 calls cover 128 MiB, so a removed
        // guard drains the stream instead of hanging, and the JSON parse
        // error that follows fails the message assertion below.
        if (reads <= 2048) {
          buffer.fill(0x78)
          return { bytesRead: buffer.length, buffer }
        }
        return { bytesRead: 0, buffer }
      },
      close: async () => undefined,
    }))

    const result = await readDispatchArtifact({ correlation_id: CORRELATION, section: "summary" }, "/stream-guard-root")

    expect(result).toMatchObject({
      status: "error",
      code: "invalid_artifact",
      message: `artifact exceeds ${MAX_BYTES} bytes`,
    })
    // The guard fired inside the read loop: 64 MiB is exactly 1024 full
    // 64 KiB reads, so the 1025th read must be the one that trips it.
    expect(reads).toBe(1025)
    // Evidence the pre-read stat branch cannot have produced the error.
    expect(statSize).toBeLessThan(MAX_BYTES)
  })
})
