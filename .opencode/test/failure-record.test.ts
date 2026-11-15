import { mkdir, mkdtemp, readFile, readdir, rm, stat, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import { DISPATCH_ARTIFACT_DIRECTORY } from "../lib/review-dispatch-artifact.js"
import {
  ReviewDispatchFailurePayloadSchema,
  buildFailureRecordEnvelope,
  createFinalizationDependencies,
  finalizeReviewDispatch,
  persistFailureRecord,
  type FinalizationDependencies,
  type FinalizeReviewDispatchInput,
  type ReviewDispatchFailureEnvelope,
  type ReviewDispatchFailurePayload,
  type ReviewDispatchPayload,
} from "../plugins/review-dispatch/index.js"
import samplePayload from "../../schemas/review-dispatch/samples/sample-review-dispatch.json"

// RD-FR-066 sanity bounds (design D3): a payload at or below the bound is
// carried whole; a payload beyond it is rejected — never shortened.
const ERROR_CHARS_BOUND = 131_072
const ERROR_ENTRIES_BOUND = 65_536

const RECORD_ID = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
const DISPATCH_ID = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
const TEMPORARY_ID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
const COLLISION_ID = "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee"
const SECOND_TEMPORARY_ID = "ffffffff-ffff-4fff-8fff-ffffffffffff"
// The sample dispatch payload's correlation id — UUID-usable, so a record
// for a failed finalization of the sample is keyed by it (design D3).
const SAMPLE_DISPATCH_ID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
const COMMIT = "2222222222222222222222222222222222222222"
const TIMESTAMP = "2026-11-15T12:00:00.000Z"

function failurePayload(
  overrides: Partial<ReviewDispatchFailurePayload> = {},
): ReviewDispatchFailurePayload {
  return {
    dispatch_correlation_id: null,
    status: "failed",
    operation_verdict: "INCONCLUSIVE",
    errors: ["persist review dispatch artifact: disk full"],
    ...overrides,
  }
}

function provenance(): { branch: string; commit: string; workflow_id: string } {
  return {
    branch: "opsx/review-dispatch-errors-cap-loses-unrecov",
    commit: COMMIT,
    workflow_id: "uf.review-council",
  }
}

/** The schema-valid sample dispatch payload the finalize tests build on. */
function dispatchPayload(): ReviewDispatchPayload {
  return structuredClone(samplePayload) as unknown as ReviewDispatchPayload
}

const scratchDirectories: string[] = []

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

async function scratchDirectory(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-failure-record-"))
  scratchDirectories.push(directory)
  return directory
}

// Real-filesystem dependencies over a scratch root (the t.TempDir
// equivalent of the Go suite) with a deterministic clock and a queued UUID
// sequence, mirroring the persistEnvelope test harness.
function dependencies(projectRoot: string, identifiers: readonly string[] = [TEMPORARY_ID]): FinalizationDependencies {
  const base = createFinalizationDependencies(projectRoot)
  const queue = [...identifiers]
  return {
    ...base,
    now: (): Date => new Date(TIMESTAMP),
    uuid: (): string => queue.shift() ?? SECOND_TEMPORARY_ID,
  }
}

async function readWrittenRecord(path: string): Promise<ReviewDispatchFailureEnvelope> {
  return JSON.parse(await readFile(path, "utf8")) as ReviewDispatchFailureEnvelope
}

describe("failure record payload schema boundaries (RD-FR-066)", () => {
  it("accepts a minimal payload with no dispatch identity", () => {
    const parsed = ReviewDispatchFailurePayloadSchema.safeParse({
      dispatch_correlation_id: null,
      status: "failed",
      operation_verdict: "INCONCLUSIVE",
      errors: ["validate review dispatch input: garbage"],
    })
    expect(parsed.success).toBe(true)
    if (parsed.success) {
      expect(parsed.data.dispatch_correlation_id).toBeNull()
      expect(parsed.data.command).toBeUndefined()
      expect(parsed.data.mode).toBeUndefined()
      expect(parsed.data.status).toBe("failed")
    }
  })

  it("accepts a complete payload with command, mode, and a dispatch correlation id", () => {
    const parsed = ReviewDispatchFailurePayloadSchema.safeParse({
      dispatch_correlation_id: DISPATCH_ID,
      command: "review-council",
      mode: "code",
      status: "failed",
      operation_verdict: "UNAVAILABLE",
      errors: ["first", "second"],
    })
    expect(parsed.success).toBe(true)
    if (parsed.success) {
      expect(parsed.data.command).toBe("review-council")
      expect(parsed.data.mode).toBe("code")
      expect(parsed.data.operation_verdict).toBe("UNAVAILABLE")
      expect(parsed.data.errors).toEqual(["first", "second"])
    }
  })

  it("accepts an error entry at the 131,072-character bound and rejects one beyond it", () => {
    const atBound = ReviewDispatchFailurePayloadSchema.safeParse(
      failurePayload({ errors: ["x".repeat(ERROR_CHARS_BOUND)] }),
    )
    expect(atBound.success).toBe(true)

    const beyondBound = ReviewDispatchFailurePayloadSchema.safeParse(
      failurePayload({ errors: ["x".repeat(ERROR_CHARS_BOUND + 1)] }),
    )
    expect(beyondBound.success).toBe(false)
  })

  it("accepts 65,536 error entries and rejects 65,537", () => {
    const atBound = ReviewDispatchFailurePayloadSchema.safeParse(
      failurePayload({ errors: Array.from({ length: ERROR_ENTRIES_BOUND }, (_, index) => `error ${index}`) }),
    )
    expect(atBound.success).toBe(true)

    const beyondBound = ReviewDispatchFailurePayloadSchema.safeParse(
      failurePayload({ errors: Array.from({ length: ERROR_ENTRIES_BOUND + 1 }, (_, index) => `error ${index}`) }),
    )
    expect(beyondBound.success).toBe(false)
  })

  it("rejects an empty errors array and empty error strings", () => {
    expect(ReviewDispatchFailurePayloadSchema.safeParse(failurePayload({ errors: [] })).success).toBe(false)
    expect(ReviewDispatchFailurePayloadSchema.safeParse(failurePayload({ errors: [""] })).success).toBe(false)
  })

  it("rejects a status other than failed", () => {
    const parsed = ReviewDispatchFailurePayloadSchema.safeParse({
      dispatch_correlation_id: null,
      status: "success",
      operation_verdict: "INCONCLUSIVE",
      errors: ["one"],
    })
    expect(parsed.success).toBe(false)
  })

  it("rejects a dispatch correlation id that is neither null nor a UUID", () => {
    expect(
      ReviewDispatchFailurePayloadSchema.safeParse(failurePayload({ dispatch_correlation_id: "not-a-uuid" }))
        .success,
    ).toBe(false)
    expect(
      ReviewDispatchFailurePayloadSchema.safeParse({
        status: "failed",
        operation_verdict: "INCONCLUSIVE",
        errors: ["one"],
      }).success,
    ).toBe(false)
  })

  it("rejects command, mode, or operation verdict outside their enums", () => {
    expect(
      ReviewDispatchFailurePayloadSchema.safeParse(failurePayload({ command: "review-pr" })).success,
    ).toBe(false)
    expect(ReviewDispatchFailurePayloadSchema.safeParse(failurePayload({ mode: "spec" })).success).toBe(false)
    expect(
      ReviewDispatchFailurePayloadSchema.safeParse(failurePayload({ operation_verdict: "APPROVED" })).success,
    ).toBe(false)
  })

  it("rejects unknown payload keys", () => {
    const parsed = ReviewDispatchFailurePayloadSchema.safeParse({
      ...failurePayload(),
      artifact_path: "/tmp/rogue.json",
    })
    expect(parsed.success).toBe(false)
  })
})

describe("failure record envelope builder (RD-FR-066 D3)", () => {
  it("emits the review-dispatch-failure Hero Interface Contract envelope", () => {
    const envelope = buildFailureRecordEnvelope(
      failurePayload({
        dispatch_correlation_id: DISPATCH_ID,
        command: "review-council",
        mode: "code",
        operation_verdict: "UNAVAILABLE",
      }),
      provenance(),
      RECORD_ID,
      TIMESTAMP,
    )
    expect(envelope.hero).toBe("the-divisor")
    expect(envelope.version).toBe("1.0.0")
    expect(envelope.timestamp).toBe(TIMESTAMP)
    expect(envelope.artifact_type).toBe("review-dispatch-failure")
    expect(envelope.schema_version).toBe("1.0.0")
    expect(Object.keys(envelope)).toEqual([
      "hero",
      "version",
      "timestamp",
      "artifact_type",
      "schema_version",
      "context",
      "payload",
    ])
    expect(Object.keys(envelope.context)).toEqual(["branch", "commit", "correlation_id", "workflow_id"])
  })

  it("copies schema-valid provenance verbatim and keys the context by the record id", () => {
    const envelope = buildFailureRecordEnvelope(failurePayload(), provenance(), RECORD_ID, TIMESTAMP)
    expect(envelope.context.branch).toBe("opsx/review-dispatch-errors-cap-loses-unrecov")
    expect(envelope.context.commit).toBe(COMMIT)
    expect(envelope.context.workflow_id).toBe("uf.review-council")
    // The record's own correlation id is the context key; the dispatch's
    // (possibly null) identity lives in the payload (design D3).
    expect(envelope.context.correlation_id).toBe(RECORD_ID)
    expect(envelope.payload.dispatch_correlation_id).toBeNull()
  })

  it("keeps the record correlation id distinct from the dispatch correlation id", () => {
    const envelope = buildFailureRecordEnvelope(
      failurePayload({ dispatch_correlation_id: DISPATCH_ID }),
      provenance(),
      RECORD_ID,
      TIMESTAMP,
    )
    expect(envelope.context.correlation_id).toBe(RECORD_ID)
    expect(envelope.payload.dispatch_correlation_id).toBe(DISPATCH_ID)
  })

  it("degrades each unusable provenance field to unknown, independently", () => {
    const partial = buildFailureRecordEnvelope(
      failurePayload(),
      { branch: "opsx/some-branch", commit: 42, workflow_id: "bad\nid" },
      RECORD_ID,
      TIMESTAMP,
    )
    expect(partial.context.branch).toBe("opsx/some-branch")
    expect(partial.context.commit).toBe("unknown")
    expect(partial.context.workflow_id).toBe("unknown")

    const garbage = buildFailureRecordEnvelope(failurePayload(), "not-an-object", RECORD_ID, TIMESTAMP)
    expect(garbage.context).toEqual({
      branch: "unknown",
      commit: "unknown",
      correlation_id: RECORD_ID,
      workflow_id: "unknown",
    })
  })

  it("carries the payload keys in the contract order, omitting unknown command and mode", () => {
    const complete = buildFailureRecordEnvelope(
      failurePayload({ command: "triage-issue", mode: "triage" }),
      provenance(),
      RECORD_ID,
      TIMESTAMP,
    )
    expect(Object.keys(complete.payload)).toEqual([
      "dispatch_correlation_id",
      "command",
      "mode",
      "status",
      "operation_verdict",
      "errors",
    ])

    const minimal = buildFailureRecordEnvelope(failurePayload(), provenance(), RECORD_ID, TIMESTAMP)
    expect(Object.keys(minimal.payload)).toEqual([
      "dispatch_correlation_id",
      "status",
      "operation_verdict",
      "errors",
    ])
  })

  it("carries errors verbatim, whole, up to the sanity bound", () => {
    const composed = `validate review dispatch input: ${"issue; ".repeat(2048)}`
    const envelope = buildFailureRecordEnvelope(
      failurePayload({ errors: [composed, "second"] }),
      provenance(),
      RECORD_ID,
      TIMESTAMP,
    )
    expect(envelope.payload.errors).toEqual([composed, "second"])

    const atBound = "x".repeat(ERROR_CHARS_BOUND)
    const bounded = buildFailureRecordEnvelope(
      failurePayload({ errors: [atBound] }),
      provenance(),
      RECORD_ID,
      TIMESTAMP,
    )
    expect(bounded.payload.errors).toEqual([atBound])
  })

  it("fails closed on a payload beyond the sanity bounds", () => {
    expect(() =>
      buildFailureRecordEnvelope(
        failurePayload({ errors: ["x".repeat(ERROR_CHARS_BOUND + 1)] }),
        provenance(),
        RECORD_ID,
        TIMESTAMP,
      ),
    ).toThrowError(/validate failure record payload/)
  })

  it("fails closed on a record correlation id that is not a UUID", () => {
    expect(() =>
      buildFailureRecordEnvelope(failurePayload(), provenance(), "not-a-uuid", TIMESTAMP),
    ).toThrowError(/failure record correlation id is not a UUID/)
  })
})

describe("failure record persistence (RD-FR-066 D3)", () => {
  it("writes a restricted atomic failure record keyed by the record id", async () => {
    const directory = await scratchDirectory()

    const reference = await persistFailureRecord(
      failurePayload({
        dispatch_correlation_id: DISPATCH_ID,
        command: "review-council",
        mode: "code",
        errors: ["first", "second"],
      }),
      provenance(),
      RECORD_ID,
      dependencies(directory),
    )

    const artifactDirectory = join(directory, DISPATCH_ARTIFACT_DIRECTORY)
    expect(reference).toEqual({ correlation_id: RECORD_ID, path: join(artifactDirectory, `${RECORD_ID}.failure.json`) })
    const raw = await readFile(reference.path, "utf8")
    expect(raw.endsWith("\n")).toBe(true)
    const record = await readWrittenRecord(reference.path)
    expect(record.hero).toBe("the-divisor")
    expect(record.timestamp).toBe(TIMESTAMP)
    expect(record.artifact_type).toBe("review-dispatch-failure")
    // The record's own id keys the envelope context; the dispatch's rides
    // in the payload (design D3 identity contract, on disk).
    expect(record.context.correlation_id).toBe(RECORD_ID)
    expect(record.payload.dispatch_correlation_id).toBe(DISPATCH_ID)
    expect(record.payload.errors).toEqual(["first", "second"])
    expect((await stat(artifactDirectory)).mode & 0o777).toBe(0o750)
    expect((await stat(reference.path)).mode & 0o777).toBe(0o600)
    expect((await readdir(artifactDirectory)).some((name) => name.endsWith(".tmp"))).toBe(false)
  })

  it("flushes, closes, and chmods before exclusive publication", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    const operations: string[] = []
    const ordered: FinalizationDependencies = {
      ...base,
      open: async (path, flags, mode) => {
        operations.push("open")
        const handle = await base.open(path, flags, mode)
        return {
          writeFile: async (data, encoding): Promise<void> => {
            operations.push("write")
            await handle.writeFile(data, encoding)
          },
          sync: async (): Promise<void> => {
            operations.push("sync")
            await handle.sync()
          },
          close: async (): Promise<void> => {
            operations.push("close")
            await handle.close()
          },
        }
      },
      chmod: async (path, mode): Promise<void> => {
        if (path.endsWith(".tmp")) {
          operations.push("chmod")
        }
        await base.chmod(path, mode)
      },
      link: async (existingPath, newPath): Promise<void> => {
        operations.push("link")
        await base.link(existingPath, newPath)
      },
      unlink: async (path): Promise<void> => {
        operations.push("unlink")
        await base.unlink(path)
      },
    }

    const reference = await persistFailureRecord(failurePayload(), provenance(), RECORD_ID, ordered)

    expect(reference.correlation_id).toBe(RECORD_ID)
    expect(operations).toEqual(["open", "write", "sync", "close", "chmod", "link", "unlink"])
  })

  it("uses exclusive publication to preserve a real colliding record", async () => {
    const directory = await scratchDirectory()
    const artifactDirectory = join(directory, DISPATCH_ARTIFACT_DIRECTORY)
    await mkdir(artifactDirectory, { recursive: true })
    const existingPath = join(artifactDirectory, `${RECORD_ID}.failure.json`)
    await writeFile(existingPath, "existing\n", "utf8")

    const reference = await persistFailureRecord(
      failurePayload({ dispatch_correlation_id: DISPATCH_ID }),
      provenance(),
      RECORD_ID,
      dependencies(directory, [TEMPORARY_ID, COLLISION_ID, SECOND_TEMPORARY_ID]),
    )

    // A re-failed dispatch appends rather than overwrites (design D3): the
    // colliding record survives byte-for-byte and the fresh record carries
    // the dispatch's identity in its payload.
    expect(reference.correlation_id).toBe(COLLISION_ID)
    expect(reference.path).toBe(join(artifactDirectory, `${COLLISION_ID}.failure.json`))
    expect(await readFile(existingPath, "utf8")).toBe("existing\n")
    const appended = await readWrittenRecord(reference.path)
    expect(appended.context.correlation_id).toBe(COLLISION_ID)
    expect(appended.payload.dispatch_correlation_id).toBe(DISPATCH_ID)
    expect((await readdir(artifactDirectory)).sort()).toEqual([
      `${RECORD_ID}.failure.json`,
      `${COLLISION_ID}.failure.json`,
    ])
  })

  it("retries a collision detected at atomic publication", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory, [TEMPORARY_ID, COLLISION_ID, SECOND_TEMPORARY_ID])
    let linkAttempts = 0
    const racing: FinalizationDependencies = {
      ...base,
      link: async (existingPath, newPath): Promise<void> => {
        linkAttempts += 1
        if (linkAttempts === 1) {
          throw Object.assign(new Error("late collision"), { code: "EEXIST" })
        }
        await base.link(existingPath, newPath)
      },
    }

    const reference = await persistFailureRecord(failurePayload(), provenance(), RECORD_ID, racing)

    expect(reference.correlation_id).toBe(COLLISION_ID)
    expect(linkAttempts).toBe(2)
    expect(await readdir(join(directory, DISPATCH_ARTIFACT_DIRECTORY))).toEqual([`${COLLISION_ID}.failure.json`])
  })

  it("fails closed when every collision retry is exhausted", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    let linkAttempts = 0
    const alwaysColliding: FinalizationDependencies = {
      ...base,
      link: async (): Promise<void> => {
        linkAttempts += 1
        throw Object.assign(new Error("persistent collision"), { code: "EEXIST" })
      },
    }

    await expect(persistFailureRecord(failurePayload(), provenance(), RECORD_ID, alwaysColliding)).rejects.toThrow(
      /persistent collision/,
    )
    // The same bounded retry count as success artifacts
    // (MAX_CORRELATION_ATTEMPTS): one initial attempt plus 31 re-mints.
    expect(linkAttempts).toBe(32)
    // Every attempt cleaned its temporary file; no record was published.
    expect(await readdir(join(directory, DISPATCH_ARTIFACT_DIRECTORY))).toEqual([])
  })

  it("fails closed on write failure and cleans temporary files", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    const failing: FinalizationDependencies = {
      ...base,
      open: async (path, flags, mode) => {
        const handle = await base.open(path, flags, mode)
        return {
          ...handle,
          writeFile: async (): Promise<void> => {
            throw new Error("write denied")
          },
        }
      },
    }

    await expect(persistFailureRecord(failurePayload(), provenance(), RECORD_ID, failing)).rejects.toThrow(
      /write denied/,
    )
    expect(await readdir(join(directory, DISPATCH_ARTIFACT_DIRECTORY))).toEqual([])
  })

  it("prioritizes the write failure when the handle also fails to close", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    const failing: FinalizationDependencies = {
      ...base,
      open: async (path, flags, mode) => {
        const handle = await base.open(path, flags, mode)
        return {
          ...handle,
          writeFile: async (): Promise<void> => {
            throw new Error("write denied")
          },
          close: async (): Promise<void> => {
            throw new Error("close denied")
          },
        }
      },
    }

    await expect(persistFailureRecord(failurePayload(), provenance(), RECORD_ID, failing)).rejects.toThrow(
      /write denied/,
    )
    expect(await readdir(join(directory, DISPATCH_ARTIFACT_DIRECTORY))).toEqual([])
  })

  it("fails closed on link failure and removes the flushed temporary file", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    const failing: FinalizationDependencies = {
      ...base,
      link: async (): Promise<void> => {
        throw new Error("link denied")
      },
    }

    await expect(persistFailureRecord(failurePayload(), provenance(), RECORD_ID, failing)).rejects.toThrow(
      /link denied/,
    )
    expect(await readdir(join(directory, DISPATCH_ARTIFACT_DIRECTORY))).toEqual([])
  })

  it("returns the published record when temporary cleanup succeeds on retry", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    let unlinkAttempts = 0
    const flaky: FinalizationDependencies = {
      ...base,
      unlink: async (path): Promise<void> => {
        if (path.endsWith(".tmp")) {
          unlinkAttempts += 1
          if (unlinkAttempts === 1) {
            throw Object.assign(new Error("transient unlink failure"), { code: "EIO" })
          }
        }
        await base.unlink(path)
      },
    }

    const reference = await persistFailureRecord(failurePayload(), provenance(), RECORD_ID, flaky)

    // Publication already succeeded (link returned): a transient temp
    // cleanup failure that the cleanup retry clears still yields the record.
    expect(reference.correlation_id).toBe(RECORD_ID)
    expect(unlinkAttempts).toBe(2)
    expect((await readWrittenRecord(reference.path)).artifact_type).toBe("review-dispatch-failure")
    expect((await readdir(join(directory, DISPATCH_ARTIFACT_DIRECTORY))).some((name) => name.endsWith(".tmp"))).toBe(
      false,
    )
  })

  it("reports a cleanup failure alongside the original error", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    const failing: FinalizationDependencies = {
      ...base,
      unlink: async (path): Promise<void> => {
        if (path.endsWith(".tmp")) {
          throw Object.assign(new Error("transient unlink failure"), { code: "EIO" })
        }
        await base.unlink(path)
      },
    }

    await expect(persistFailureRecord(failurePayload(), provenance(), RECORD_ID, failing)).rejects.toThrow(
      /transient unlink failure; temporary cleanup failed: transient unlink failure/,
    )
    expect((await readdir(join(directory, DISPATCH_ARTIFACT_DIRECTORY))).some((name) => name.endsWith(".tmp"))).toBe(
      true,
    )
  })

  it("fails closed on an unbuildable record before touching the filesystem", async () => {
    const directory = await scratchDirectory()
    const deps = dependencies(directory)

    await expect(
      persistFailureRecord(
        failurePayload({ errors: ["x".repeat(ERROR_CHARS_BOUND + 1)] }),
        provenance(),
        RECORD_ID,
        deps,
      ),
    ).rejects.toThrow(/validate failure record payload/)
    await expect(persistFailureRecord(failurePayload(), provenance(), "not-a-uuid", deps)).rejects.toThrow(
      /failure record correlation id is not a UUID/,
    )
    // The envelope validates before the first filesystem effect (design D3
    // fail-closed): not even the artifact directory is created.
    expect(await readdir(join(directory, DISPATCH_ARTIFACT_DIRECTORY)).catch(() => null)).toBeNull()
  })
})

describe("failure record wiring into finalizeReviewDispatch (RD-FR-066 D3/D4)", () => {
  it("writes a record keyed by the dispatch id on structural rejection with a usable id", async () => {
    const directory = await scratchDirectory()
    const payload = { ...dispatchPayload(), unexpected: true }

    const result = await finalizeReviewDispatch(
      { payload, provenance: provenance() } as unknown as FinalizeReviewDispatchInput,
      dependencies(directory),
    )

    expect(result.status).toBe("failed")
    expect(result.correlation_id).toBeNull()
    expect(result.artifact_path).toBeNull()
    expect(result.errors).toHaveLength(1)
    expect(result.errors[0]).toContain('Unrecognized key: "unexpected"')
    // The record is keyed by the dispatch's UUID-usable correlation id even
    // though the rejected result itself surfaces none (design D3).
    expect(result.failure_record).toEqual({
      correlation_id: SAMPLE_DISPATCH_ID,
      path: join(directory, DISPATCH_ARTIFACT_DIRECTORY, `${SAMPLE_DISPATCH_ID}.failure.json`),
    })
    const record = await readWrittenRecord(result.failure_record!.path)
    expect(record.artifact_type).toBe("review-dispatch-failure")
    expect(record.context.correlation_id).toBe(SAMPLE_DISPATCH_ID)
    expect(record.context.branch).toBe("opsx/review-dispatch-errors-cap-loses-unrecov")
    expect(record.payload.dispatch_correlation_id).toBe(SAMPLE_DISPATCH_ID)
    expect(record.payload.command).toBe("review-council")
    expect(record.payload.mode).toBe("code")
    expect(record.payload.status).toBe("failed")
    expect(record.payload.operation_verdict).toBe("INCONCLUSIVE")
    // Complete errors verbatim: the composed structural error rides whole.
    expect(record.payload.errors).toEqual(result.errors)
  })

  it("mints a record correlation id for a garbage input with no usable dispatch id", async () => {
    const directory = await scratchDirectory()
    const MINTED_ID = "99999999-9999-4999-8999-999999999999"

    const result = await finalizeReviewDispatch(
      {
        // Structurally rejected garbage: no usable correlation id, a lone
        // valid command, and no mode at all.
        payload: { correlation_id: "not-a-uuid", command: "review-council" },
        provenance: provenance(),
      } as unknown as FinalizeReviewDispatchInput,
      dependencies(directory, [MINTED_ID]),
    )

    expect(result.status).toBe("failed")
    expect(result.correlation_id).toBeNull()
    expect(result.failure_record?.correlation_id).toBe(MINTED_ID)
    const record = await readWrittenRecord(result.failure_record!.path)
    expect(record.context.correlation_id).toBe(MINTED_ID)
    expect(record.payload.dispatch_correlation_id).toBeNull()
    // Best-effort facts from the failed input (design D3): a schema-valid
    // command rides along, an absent mode stays absent, and provenance is
    // projected from the rejected input itself.
    expect(record.payload.command).toBe("review-council")
    expect(record.payload).not.toHaveProperty("mode")
    expect(record.context.branch).toBe("opsx/review-dispatch-errors-cap-loses-unrecov")
    expect(record.payload.errors).toEqual(result.errors)
  })

  it("carries the complete semantic errors verbatim on semantic rejection", async () => {
    const directory = await scratchDirectory()
    const payload = dispatchPayload()
    payload.run_counts.success = 0

    const result = await finalizeReviewDispatch(
      { payload, provenance: provenance() },
      dependencies(directory),
    )

    expect(result.status).toBe("failed")
    expect(result.correlation_id).toBe(SAMPLE_DISPATCH_ID)
    expect(result.errors.length).toBeGreaterThan(0)
    expect(result.failure_record).not.toBeNull()
    const record = await readWrittenRecord(result.failure_record!.path)
    expect(record.payload.dispatch_correlation_id).toBe(SAMPLE_DISPATCH_ID)
    expect(record.payload.command).toBe("review-council")
    expect(record.payload.mode).toBe("code")
    expect(record.payload.operation_verdict).toBe("INCONCLUSIVE")
    expect(record.payload.errors).toEqual(result.errors)
  })

  it("writes the record when only the artifact write fails", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    let linkAttempts = 0
    const flakyArtifact: FinalizationDependencies = {
      ...base,
      link: async (existingPath, newPath): Promise<void> => {
        linkAttempts += 1
        if (linkAttempts === 1) {
          throw Object.assign(new Error("EIO: artifact link failure"), { code: "EIO" })
        }
        await base.link(existingPath, newPath)
      },
    }

    const result = await finalizeReviewDispatch(
      { payload: dispatchPayload(), provenance: provenance() },
      flakyArtifact,
    )

    expect(result.status).toBe("failed")
    expect(linkAttempts).toBe(2)
    // The surviving record write adds no error and loses nothing: exactly
    // the persistence error, once, plus the record locator (design D4).
    expect(result.errors).toEqual(["persist review dispatch artifact: EIO: artifact link failure"])
    expect(result.failure_record).toEqual({
      correlation_id: SAMPLE_DISPATCH_ID,
      path: join(directory, DISPATCH_ARTIFACT_DIRECTORY, `${SAMPLE_DISPATCH_ID}.failure.json`),
    })
    const record = await readWrittenRecord(result.failure_record!.path)
    expect(record.payload.dispatch_correlation_id).toBe(SAMPLE_DISPATCH_ID)
    expect(record.payload.errors).toEqual(["persist review dispatch artifact: EIO: artifact link failure"])
  })

  it("degrades without masking when the record write also fails (design D4)", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    const failing: FinalizationDependencies = {
      ...base,
      link: async (): Promise<void> => {
        throw Object.assign(new Error("EIO: disk full"), { code: "EIO" })
      },
    }

    const result = await finalizeReviewDispatch(
      { payload: dispatchPayload(), provenance: provenance() },
      failing,
    )

    expect(result.status).toBe("failed")
    expect(result.failure_record).toBeNull()
    // The finalization failure survives whole: the persistence error stays
    // first and exactly one bounded record-write degradation joins it.
    expect(result.errors).toEqual([
      "persist review dispatch artifact: EIO: disk full",
      "persist failure record: EIO: disk full",
    ])
    expect(await readdir(join(directory, DISPATCH_ARTIFACT_DIRECTORY))).toEqual([])
  })
})
