// ---------------------------------------------------------------------------
// Human-decision record store and record_human_decision tool — SYNTH-006
// coverage (synthetic-compaction-continue-prompt-rec, tasks 2.2-2.5).
//
// Spec: openspec/changes/synthetic-compaction-continue-prompt-rec/specs/
// checkpoint-gate/spec.md (SYNTH-001, SYNTH-006).
//
// Covers the store contract of `.uf/progress/<key>/human-decisions.json`
// against real temp roots: append-only semantics with machine-stamped
// decision_id (UUID v4) and recorded_at; the sanitized key resolution with
// traversal rejection; symlink refusal on directory and record; 0o700/0o600
// modes; atomic temp-file-plus-rename writes leaving no lock or temp
// residue; the empty-record first-use case; fail-closed corrupted-record
// reads that never overwrite the audit trail; serialized concurrent
// appends; and the advisory-lock stale-break and acquire-timeout paths
// (exercised through the injected timing knobs, never real-time waits).
//
// The record_human_decision tool suite (task 2.3) covers the in-session
// question path's own contract: the happy path with provenance, the
// synthetic-prompt deny-list across normalization variants, SYNTH-006
// input bounds with byte semantics, machine-generated identity, and the
// fail-closed persist envelope. The envelope parameter wiring
// (uf_delegate_command) is covered by unleash-delegate.test.ts; the
// checkpoint-side reference enforcement is covered by the tasks 2.4-2.5
// suite at the bottom of this file (SYNTH-001 with SYNTH-006 lookup
// semantics), including the task 2.5 cases: the declared-step tripwire
// (override data riding `implement` without a ref) and the prior-session
// free-text citation that does not carry forward.
// ---------------------------------------------------------------------------

import { randomUUID } from "node:crypto"
import {
  mkdir,
  mkdtemp,
  readdir,
  readFile,
  rm,
  stat,
  symlink,
  utimes,
  writeFile,
} from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import {
  appendHumanDecisions,
  appendSpecReviewIterationToState,
  HUMAN_DECISIONS_FILENAME,
  readHumanDecisionRecord,
  readState,
  writeState,
  type HumanDecisionStoreOptions,
} from "../lib/progress-state.js"
import {
  HUMAN_DECISION_ID_PATTERN,
  HUMAN_DECISION_REF_MAX_LENGTH,
  SPEC_REVIEW_RECORD_KEY,
  SYNTHETIC_COMPACTION_PROMPT,
  type HumanDecisionAppendInput,
  type HumanDecisionEntry,
  type SpecReviewGateRecord,
} from "../lib/uf-workflow-types.js"
import {
  CHECKPOINT_STEP_NAME_MAX_LENGTH,
  createCheckpointTool,
  createRecordHumanDecisionTool,
  isDeclaredUnleashCheckpointStep,
  UNLEASH_DECLARED_STEP_ALLOWLIST,
  type GitExec,
  type ProgressState,
  type RecordHumanDecisionData,
  type ToolResult,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Temp roots created during a test, removed in afterEach (isolation). */
const ROOTS: string[] = []

/** Create an isolated project root holding no `.uf/progress/` yet. */
async function newRoot(): Promise<string> {
  const root = await mkdtemp(join(tmpdir(), "uf-human-decisions-"))
  ROOTS.push(root)
  return root
}

/** One orchestrator-relay append input, the envelope's write surface. */
function relay(
  question: string,
  answer: string,
): HumanDecisionAppendInput {
  return { source: "orchestrator-relay", question, answer }
}

/** One question-tool append input, the in-session write surface. */
function questionTool(
  question: string,
  answer: string,
): HumanDecisionAppendInput {
  return { source: "question-tool", question, answer }
}

/** The record path for a state key inside a root (test-side derivation). */
function recordPath(root: string, key: string): string {
  return join(root, ".uf/progress", key, HUMAN_DECISIONS_FILENAME)
}

/** Fast lock timing knobs so timeout paths run in milliseconds. */
const FAST_LOCK: HumanDecisionStoreOptions = {
  lockAcquireTimeoutMs: 80,
  lockRetryDelayMs: 5,
}

afterEach(async () => {
  await Promise.all(
    ROOTS.splice(0).map((root) => rm(root, { recursive: true, force: true })),
  )
})

// ---------------------------------------------------------------------------
// Append and read semantics
// ---------------------------------------------------------------------------

describe("human-decision record append and read", () => {
  it("treats a missing record as the empty-record case, not an error", async () => {
    const root = await newRoot()
    const entries = readHumanDecisionRecord(root, "opsx/change-name")
    expect(entries).toEqual([])
  })

  it("stamps machine fields and round-trips one entry through read", async () => {
    const root = await newRoot()
    const stamped = await appendHumanDecisions(root, "opsx/change-name", [
      relay("Proceed past the advisory gate?", "Yes, proceed to implement."),
    ])

    expect(stamped).toHaveLength(1)
    const entry = stamped[0] as HumanDecisionEntry
    expect(HUMAN_DECISION_ID_PATTERN.test(entry.decision_id)).toBe(true)
    expect(entry.source).toBe("orchestrator-relay")
    expect(entry.question).toBe("Proceed past the advisory gate?")
    expect(entry.answer).toBe("Yes, proceed to implement.")
    expect(new Date(entry.recorded_at).toString()).not.toBe("Invalid Date")

    const read = readHumanDecisionRecord(root, "opsx/change-name")
    expect(read).toEqual(stamped)
  })

  it("accepts both write surfaces as the entry source", async () => {
    const root = await newRoot()
    await appendHumanDecisions(root, "opsx/change-name", [
      relay("Relay question?", "Relay answer."),
      questionTool("Session question?", "Session answer."),
    ])
    const entries = readHumanDecisionRecord(root, "opsx/change-name")
    expect(entries.map((entry) => entry.source)).toEqual([
      "orchestrator-relay",
      "question-tool",
    ])
  })

  it("is append-only across calls: prior entries persist verbatim in order", async () => {
    const root = await newRoot()
    const first = await appendHumanDecisions(root, "opsx/change-name", [
      relay("First question?", "First answer."),
    ])
    const second = await appendHumanDecisions(root, "opsx/change-name", [
      relay("Second question?", "Second answer."),
      relay("Third question?", "Third answer."),
    ])

    const read = readHumanDecisionRecord(root, "opsx/change-name")
    expect(read).toHaveLength(3)
    expect(read[0]).toEqual(first[0])
    expect(read.slice(1)).toEqual(second)
    const ids = new Set(read.map((entry) => entry.decision_id))
    expect(ids.size).toBe(3)
  })

  it("resolves the key from a raw branch name exactly like the sanitized key", async () => {
    const root = await newRoot()
    await appendHumanDecisions(root, "opsx/change-name", [
      relay("Branch-keyed question?", "Branch-keyed answer."),
    ])
    // readSpecReviewGateRecord ergonomics: raw branch or state key both work.
    const viaSanitized = readHumanDecisionRecord(root, "opsx-change-name")
    const viaBranch = readHumanDecisionRecord(root, "opsx/change-name")
    expect(viaSanitized).toEqual(viaBranch)
    expect(viaSanitized).toHaveLength(1)
  })

  it("accepts a question or answer of exactly 16 KiB (inclusive bound)", async () => {
    const root = await newRoot()
    const edge = "x".repeat(16 * 1024)
    const stamped = await appendHumanDecisions(root, "opsx/change-name", [
      relay(edge, edge),
    ])
    expect(stamped).toHaveLength(1)
    expect(readHumanDecisionRecord(root, "opsx/change-name")).toHaveLength(1)
  })
})

// ---------------------------------------------------------------------------
// Store discipline: modes, atomic writes, lock residue
// ---------------------------------------------------------------------------

describe("human-decision record store discipline", () => {
  it("writes 0o600 records inside 0o700 directories with no lock or temp residue", async () => {
    const root = await newRoot()
    await appendHumanDecisions(root, "opsx/change-name", [
      relay("Question?", "Answer."),
    ])

    const keyDir = join(root, ".uf/progress", "opsx-change-name")
    const recordStat = await stat(recordPath(root, "opsx-change-name"))
    expect(recordStat.mode & 0o777).toBe(0o600)
    const dirStat = await stat(keyDir)
    expect(dirStat.mode & 0o777).toBe(0o700)

    // Atomic write + lock release leave exactly the record behind: no
    // *.tmp, no *.lock.
    const names = await readdir(keyDir)
    expect(names.sort()).toEqual([HUMAN_DECISIONS_FILENAME])

    // The record is a JSON array with a trailing newline (writeState style).
    // The raw branch key sanitizes to `opsx-change-name` (SYNTH-006: the
    // record path is derived through the same sanitized key resolution as
    // the progress state store), so the content check reads the sanitized
    // path this test already statted above.
    const raw = await readFile(recordPath(root, "opsx-change-name"), "utf8")
    expect(raw.endsWith("\n")).toBe(true)
    const parsed: unknown = JSON.parse(raw)
    expect(Array.isArray(parsed)).toBe(true)
  })
})

// ---------------------------------------------------------------------------
// Guards: traversal, symlinks, corruption
// ---------------------------------------------------------------------------

describe("human-decision record guards", () => {
  it("rejects a traversal-shaped key before touching the filesystem", async () => {
    const root = await newRoot()
    await expect(
      appendHumanDecisions(root, "../../etc", [relay("Question?", "Answer.")]),
    ).rejects.toThrow(/Path traversal blocked/)
    // readHumanDecisionRecord is synchronous by design — the codebase's
    // durable reads (readState, readSpecReviewGateRecord) are sync, and the
    // append's lock-held critical section calls it synchronously — so its
    // guards assert with the sync-throw form rather than .rejects.
    expect(() => readHumanDecisionRecord(root, "../../etc")).toThrow(
      /Path traversal blocked/,
    )
    // No file outside .uf/progress/ — in fact nothing at all — was created.
    expect(await readdir(root)).toEqual([])
  })

  it("rejects a Windows-style traversal key identically", async () => {
    const root = await newRoot()
    await expect(
      appendHumanDecisions(root, "..\\..\\etc", [relay("Question?", "Answer.")]),
    ).rejects.toThrow(/Path traversal blocked/)
    expect(await readdir(root)).toEqual([])
  })

  it("refuses to append into a symlinked key directory", async () => {
    const root = await newRoot()
    const target = join(root, "outside-target")
    await mkdir(target)
    await mkdir(join(root, ".uf/progress"), { recursive: true })
    await symlink(target, join(root, ".uf/progress", "opsx-change-name"))

    await expect(
      appendHumanDecisions(root, "opsx/change-name", [relay("Question?", "Answer.")]),
    ).rejects.toThrow(/Refusing to write into symlinked directory/)
    // Nothing escaped through the symlink.
    expect(await readdir(target)).toEqual([])
  })

  it("refuses to read or append a symlinked record file, leaving the target untouched", async () => {
    const root = await newRoot()
    const outside = join(root, "outside.json")
    await writeFile(outside, "[]")
    const keyDir = join(root, ".uf/progress", "opsx-change-name")
    await mkdir(keyDir, { recursive: true })
    await symlink(outside, recordPath(root, "opsx-change-name"))

    expect(() => readHumanDecisionRecord(root, "opsx/change-name")).toThrow(
      /Refusing to read symlinked human-decision record/,
    )
    await expect(
      appendHumanDecisions(root, "opsx/change-name", [relay("Question?", "Answer.")]),
    ).rejects.toThrow(/Refusing to read symlinked human-decision record/)
    // The append failed inside the lock's critical section, so the lock
    // was still released and the outside target never received a write.
    expect(await readFile(outside, "utf8")).toBe("[]")
    const names = (await readdir(keyDir)).sort()
    expect(names).toEqual([HUMAN_DECISIONS_FILENAME])
  })

  it("fails closed on corrupted JSON without overwriting the record", async () => {
    const root = await newRoot()
    const keyDir = join(root, ".uf/progress", "opsx-change-name")
    await mkdir(keyDir, { recursive: true })
    const corrupted = "{ not json"
    await writeFile(recordPath(root, "opsx-change-name"), corrupted)

    expect(() => readHumanDecisionRecord(root, "opsx/change-name")).toThrow(
      /corrupted \(malformed JSON\)/,
    )
    await expect(
      appendHumanDecisions(root, "opsx/change-name", [relay("Question?", "Answer.")]),
    ).rejects.toThrow(/corrupted/)
    // The audit trail is never overwritten from an unreadable state.
    expect(await readFile(recordPath(root, "opsx-change-name"), "utf8")).toBe(
      corrupted,
    )
  })

  it("fails closed on malformed entry shapes without overwriting the record", async () => {
    const root = await newRoot()
    const keyDir = join(root, ".uf/progress", "opsx-change-name")
    await mkdir(keyDir, { recursive: true })
    const malformed = JSON.stringify([
      {
        decision_id: "not-a-uuid",
        source: "orchestrator-relay",
        question: "Question?",
        answer: "Answer.",
        recorded_at: "2026-11-15T00:00:00.000Z",
      },
    ])
    await writeFile(recordPath(root, "opsx-change-name"), malformed)

    expect(() => readHumanDecisionRecord(root, "opsx/change-name")).toThrow(
      /corrupted \(malformed entry shape\)/,
    )
    await expect(
      appendHumanDecisions(root, "opsx/change-name", [relay("Question?", "Answer.")]),
    ).rejects.toThrow(/corrupted/)
    expect(await readFile(recordPath(root, "opsx-change-name"), "utf8")).toBe(
      malformed,
    )
  })

  it("rejects input-bound violations before path derivation, with bounded identifiers only", async () => {
    const root = await newRoot()
    // The oversize entry must be named before the traversal-shaped key is
    // even considered (SYNTH-006: bounds rejected before file path
    // derivation) — proving the ordering, not just both rejections.
    const oversize = await appendHumanDecisions(
      root,
      "../../etc",
      [relay("q".repeat(16 * 1024 + 1), "Answer.")],
    ).catch((err: unknown) => err)
    expect(oversize).toBeInstanceOf(Error)
    expect((oversize as Error).message).toBe(
      "human decision entry 0: question must be non-empty and at most 16 KiB UTF-8 bytes",
    )

    await expect(
      appendHumanDecisions(root, "opsx/change-name", [relay("Question?", "")]),
    ).rejects.toThrow(
      "human decision entry 0: answer must be non-empty and at most 16 KiB UTF-8 bytes",
    )
    await expect(
      appendHumanDecisions(root, "opsx/change-name", [
        { source: "carrier-pigeon" as unknown as "orchestrator-relay", question: "Q?", answer: "A." },
      ]),
    ).rejects.toThrow("human decision entry 0: unknown source")
    expect(await readdir(root)).toEqual([])
  })
})

// ---------------------------------------------------------------------------
// Concurrency: serialized appends and the advisory lock protocol
// ---------------------------------------------------------------------------

describe("human-decision record concurrency", () => {
  it("loses no entries across concurrent appends", async () => {
    const root = await newRoot()
    const count = 8
    const results = await Promise.all(
      Array.from({ length: count }, (_, index) =>
        appendHumanDecisions(root, "opsx/change-name", [
          relay(`Question ${index}?`, `Answer ${index}.`),
        ]),
      ),
    )

    const read = readHumanDecisionRecord(root, "opsx/change-name")
    expect(read).toHaveLength(count)
    const stampedIds = new Set(
      results.map((batch) => (batch[0] as HumanDecisionEntry).decision_id),
    )
    expect(stampedIds.size).toBe(count)
    for (const entry of read) {
      expect(stampedIds.has(entry.decision_id)).toBe(true)
    }
    // One atomic record file; no residue from the contended writers.
    const names = (await readdir(join(root, ".uf/progress", "opsx-change-name"))).sort()
    expect(names).toEqual([HUMAN_DECISIONS_FILENAME])
  })

  it("breaks a stale lock left by a crashed holder", async () => {
    const root = await newRoot()
    const lockPath = recordPath(root, "opsx-change-name") + ".lock"
    await mkdir(join(root, ".uf/progress", "opsx-change-name"), { recursive: true })
    await writeFile(lockPath, "")
    // A lock 60s old exceeds the 30s default stale threshold.
    const stale = new Date(Date.now() - 60_000)
    await utimes(lockPath, stale, stale)

    const stamped = await appendHumanDecisions(root, "opsx/change-name", [
      relay("Recovered question?", "Recovered answer."),
    ])
    expect(stamped).toHaveLength(1)
    expect(readHumanDecisionRecord(root, "opsx/change-name")).toHaveLength(1)
    const names = (await readdir(join(root, ".uf/progress", "opsx-change-name"))).sort()
    expect(names).toEqual([HUMAN_DECISIONS_FILENAME])
  })

  it("times out acquisition against a fresh lock without touching the record", async () => {
    const root = await newRoot()
    // The advisory lock lives beside the record, whose path is derived
    // through the sanitized key resolution (SYNTH-006): the raw branch key
    // `opsx/change-name` sanitizes to `opsx-change-name`, the directory this
    // test creates below and reads back in its final assertion.
    const lockPath = recordPath(root, "opsx-change-name") + ".lock"
    await mkdir(join(root, ".uf/progress", "opsx-change-name"), { recursive: true })
    await writeFile(lockPath, `held-by-${randomUUID()}`)

    await expect(
      appendHumanDecisions(root, "opsx/change-name", [relay("Question?", "Answer.")], FAST_LOCK),
    ).rejects.toThrow(/timeout acquiring the human-decision record lock/)

    // The fresh holder's lock is not the failed acquirer's to break, and no
    // record was created by the failed append.
    const names = (await readdir(join(root, ".uf/progress", "opsx-change-name"))).sort()
    expect(names).toEqual([`${HUMAN_DECISIONS_FILENAME}.lock`])
  })
})

// ---------------------------------------------------------------------------
// record_human_decision tool — SYNTH-006, task 2.3
// ---------------------------------------------------------------------------

/** Mock git exec reporting the raw branch the record key derives from. */
const branchExec: GitExec = () => "opsx/change-name"

/** Invoke the record_human_decision tool and parse its JSON envelope. */
async function invokeRecordTool(
  root: string,
  args: Record<string, unknown>,
): Promise<ToolResult<RecordHumanDecisionData>> {
  const toolInstance = createRecordHumanDecisionTool(branchExec, root)
  const result = await (
    toolInstance as {
      execute: (
        args: Record<string, unknown>,
        ctx: { sessionID: string },
      ) => Promise<{ output: string }>
    }
  ).execute(args, { sessionID: "test-session" })
  return JSON.parse(result.output) as ToolResult<RecordHumanDecisionData>
}

describe("record_human_decision tool", () => {
  it("records a question-tool answer with machine-stamped identity and provenance", async () => {
    const root = await newRoot()
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Proceed past the advisory gate?",
      answer: "Yes, proceed to implement.",
    })

    expect(result.status).toBe("ok")
    if (result.status !== "ok") return
    const entry = result.data as RecordHumanDecisionData
    expect(HUMAN_DECISION_ID_PATTERN.test(entry.decision_id)).toBe(true)
    expect(entry.source).toBe("question-tool")
    expect(entry.question).toBe("Proceed past the advisory gate?")
    expect(entry.answer).toBe("Yes, proceed to implement.")
    expect(new Date(entry.recorded_at).toString()).not.toBe("Invalid Date")

    // The record under the sanitized key round-trips the exact entry.
    expect(readHumanDecisionRecord(root, "opsx/change-name")).toEqual([entry])
  })

  it("honors an explicit key override like the checkpoint tools", async () => {
    const root = await newRoot()
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Question?",
      answer: "Answer.",
      key: "explicit-key",
    })

    expect(result.status).toBe("ok")
    expect(readHumanDecisionRecord(root, "explicit-key")).toHaveLength(1)
  })

  it("appends after orchestrator-relay entries without losing either surface's entry", async () => {
    const root = await newRoot()
    await appendHumanDecisions(root, "opsx/change-name", [
      relay("Relayed question?", "Relayed answer."),
    ])
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Session question?",
      answer: "Session answer.",
    })

    expect(result.status).toBe("ok")
    const read = readHumanDecisionRecord(root, "opsx/change-name")
    expect(read).toHaveLength(2)
    expect(read.map((entry) => entry.source)).toEqual([
      "orchestrator-relay",
      "question-tool",
    ])
  })

  it("refuses the synthetic compaction prompt verbatim with a machine-parseable error", async () => {
    const root = await newRoot()
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "How should I proceed?",
      answer: SYNTHETIC_COMPACTION_PROMPT,
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("synthetic_prompt_answer")
    expect(result.retryable).toBe(false)
    // Bounded identifiers only: the refusal never echoes the answer text.
    expect(result.message).not.toContain("Continue if you have next steps")
    // Nothing was recorded — the refusal precedes any store write.
    expect(readHumanDecisionRecord(root, "opsx/change-name")).toEqual([])
  })

  it.each([
    ["uppercased", SYNTHETIC_COMPACTION_PROMPT.toUpperCase()],
    [
      "whitespace-varied",
      `  Continue   if you have next steps,\nor\tstop and ask for clarification   if you are unsure how to proceed  `,
    ],
    [
      "double-quoted inside a larger justification",
      `User explicitly directed continuation via "${SYNTHETIC_COMPACTION_PROMPT}" with the advisory gate.`,
    ],
    [
      "single-quoted inside a larger justification (incident shape)",
      `User explicitly directed continuation via '${SYNTHETIC_COMPACTION_PROMPT}' — proceed.`,
    ],
    [
      "smart-quoted",
      `The human answered “${SYNTHETIC_COMPACTION_PROMPT}” so continue.`,
    ],
    [
      "case-, whitespace-, and quoting-varied together",
      `  sure? "CONTINUE IF YOU HAVE NEXT STEPS,\n  or   stop and ask for clarification if you are UNSURE how to proceed."  `,
    ],
  ])("refuses the synthetic prompt %s", async (_label, answer) => {
    const root = await newRoot()
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Question?",
      answer,
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("synthetic_prompt_answer")
    expect(readHumanDecisionRecord(root, "opsx/change-name")).toEqual([])
  })

  it.each([
    ["an ordinary affirmative answer", "Yes — proceed to implement phase 3."],
    [
      "a partial prompt quote (the incident's justification text)",
      "User explicitly directed continuation via 'Continue if you have next steps'",
    ],
    ["an unrelated continuation instruction", "Continue with the next task."],
  ])("accepts %s", async (_label, answer) => {
    // Containment of the FULL normalized prompt is the spec's match
    // definition (SYNTH-006): a partial quote does not contain the whole
    // prompt, so it passes the deny-list and remains governed by
    // SYNTH-001's reference requirement at checkpoint time (task 2.4).
    const root = await newRoot()
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Question?",
      answer,
    })

    expect(result.status).toBe("ok")
    expect(readHumanDecisionRecord(root, "opsx/change-name")).toHaveLength(1)
  })

  it.each([
    ["question", "question"],
    ["answer", "answer"],
  ] as const)("rejects an oversize %s before any store write (byte bounds)", async (_label, field) => {
    const root = await newRoot()
    // 8,193 two-byte characters = 16,386 UTF-8 bytes — over the 16 KiB
    // byte bound while under zod's character counting, proving the byte
    // semantics SYNTH-006 mandates.
    const oversize = "é".repeat(8 * 1024 + 1)
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Question?",
      answer: "Answer.",
      [field]: oversize,
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("INVALID_INPUT")
    expect(result.message).toContain(`${field}: `)
    // Rejected before any security-sensitive operation: nothing on disk.
    expect(await readdir(root)).toEqual([])
  })

  it("accepts question and answer of exactly 16 KiB (inclusive bound)", async () => {
    const root = await newRoot()
    const edge = "x".repeat(16 * 1024)
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: edge,
      answer: edge,
    })

    expect(result.status).toBe("ok")
    expect(readHumanDecisionRecord(root, "opsx/change-name")).toHaveLength(1)
  })

  it.each([
    ["an empty question", { question: "" }],
    ["an empty answer", { answer: "" }],
    ["an unknown source", { source: "carrier-pigeon" }],
  ])("rejects %s through the input schema", async (_label, override) => {
    const root = await newRoot()
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Question?",
      answer: "Answer.",
      ...override,
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("INVALID_INPUT")
    expect(await readdir(root)).toEqual([])
  })

  it("rejects caller-supplied identity and timestamp stamps (strict schema)", async () => {
    const root = await newRoot()
    // Callers supply only source, question, and answer (SYNTH-006): a
    // smuggled decision_id or recorded_at is an unknown key the strict
    // object rejects — identity is machine-generated, never accepted.
    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Question?",
      answer: "Answer.",
      decision_id: "00000000-0000-4000-8000-000000000000",
      recorded_at: "2026-01-01T00:00:00.000Z",
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("INVALID_INPUT")
    expect(await readdir(root)).toEqual([])
  })

  it("machine-generates distinct UUID v4 identity across recordings", async () => {
    const root = await newRoot()
    const first = await invokeRecordTool(root, {
      source: "question-tool",
      question: "First question?",
      answer: "First answer.",
    })
    const second = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Second question?",
      answer: "Second answer.",
    })

    expect(first.status).toBe("ok")
    expect(second.status).toBe("ok")
    if (first.status !== "ok" || second.status !== "ok") return
    const firstEntry = first.data as RecordHumanDecisionData
    const secondEntry = second.data as RecordHumanDecisionData
    expect(firstEntry.decision_id).not.toBe(secondEntry.decision_id)
    expect(HUMAN_DECISION_ID_PATTERN.test(firstEntry.decision_id)).toBe(true)
    expect(HUMAN_DECISION_ID_PATTERN.test(secondEntry.decision_id)).toBe(true)
  })

  it("fails closed with HUMAN_DECISION_PERSIST_FAILED when the record is corrupted", async () => {
    const root = await newRoot()
    const keyDir = join(root, ".uf/progress", "opsx-change-name")
    await mkdir(keyDir, { recursive: true })
    const corrupted = "{ not json"
    await writeFile(recordPath(root, "opsx-change-name"), corrupted)

    const result = await invokeRecordTool(root, {
      source: "question-tool",
      question: "Question?",
      answer: "Answer.",
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("HUMAN_DECISION_PERSIST_FAILED")
    expect(result.retryable).toBe(true)
    // The audit trail is never overwritten from an unreadable state.
    expect(await readFile(recordPath(root, "opsx-change-name"), "utf8")).toBe(
      corrupted,
    )
  })
})

// ---------------------------------------------------------------------------
// Checkpoint-side reference enforcement (SYNTH-001 with SYNTH-006 lookup
// semantics, tasks 2.4-2.5)
//
// The checkpoint tool's data-key tripwire: any `data` key at any nesting
// depth containing "override" (case-insensitive) marks a gate override and
// requires a `human_decision_ref` naming a recorded decision_id. Shape is
// validated first (UUID v4, at most 256 characters — free text never
// reaches the lookup); an absent record file is the empty-record case
// (missing_human_decision); fabricated refs reject identically; non-ENOENT
// read failures fail closed (human_decision_record_unreadable). The
// HARDEN-002 spec_review_decisions surface keeps its own override
// justification machinery and is not swept into the tripwire.
// ---------------------------------------------------------------------------

/** Invoke the checkpoint tool and parse its JSON envelope. */
async function invokeCheckpoint(
  root: string,
  args: Record<string, unknown>,
): Promise<ToolResult<ProgressState>> {
  const toolInstance = createCheckpointTool(branchExec, root)
  const result = await (
    toolInstance as {
      execute: (
        args: Record<string, unknown>,
        ctx: { sessionID: string },
      ) => Promise<{ output: string }>
    }
  ).execute(args, { sessionID: "test-session" })
  return JSON.parse(result.output) as ToolResult<ProgressState>
}

describe("checkpoint human-decision reference enforcement (SYNTH-001, tasks 2.4-2.5)", () => {
  /** Scratch root whose spec-review iteration 1 consolidated to REQUEST CHANGES. */
  async function gateRoot(): Promise<string> {
    const root = await newRoot()
    appendSpecReviewIterationToState(root, "opsx-change-name", {
      verdict: "REQUEST CHANGES",
      correlation_id: "5355acba-62f4-4859-8775-7294b3e8702c",
      finding_count: 2,
      finding_ids: ["1", "2"],
    })
    return root
  }

  it("rejects the incident's undeclared override step with undeclared_step (SYNTH-003 gates first)", async () => {
    const root = await newRoot()
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "checklist-override-phase8-9",
      data: JSON.stringify({
        checklist_override: `User explicitly directed continuation via '${SYNTHETIC_COMPACTION_PROMPT}'`,
      }),
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    // Task 4.1 layering: the declared-step predicate is the authoritative
    // gate — an undeclared step is rejected regardless of its data — so the
    // incident shape (undeclared step + override data + no ref) now rejects
    // with undeclared_step before the SYNTH-001 data-key tripwire
    // evaluates. The tripwire's own envelope (step + data key,
    // `missing_human_decision`) is pinned by the declared-step test below.
    expect(result.code).toBe("undeclared_step")
    expect(result.retryable).toBe(false)
    expect(result.message).toContain("checklist-override-phase8-9")
    // The step gate is step-scoped: the override-carrying data key is not
    // named in this envelope.
    expect(result.message).not.toContain("checklist_override")
    // Bounded identifiers only — the refusal never echoes the quoted prompt.
    expect(result.message).not.toContain("Continue if you have next steps")
    // Rejected before any state write: no progress persists for the step.
    expect(readState(root, "opsx-change-name", "uf.unleash")).toBeNull()
  })

  it("rejects data.checklist_override riding the declared step implement without a ref", async () => {
    const root = await newRoot()
    // Spec scenario "Override data under a declared step name is
    // rejected": a declared step name does not launder override data.
    // SYNTH-003's step allowlist is the step-name gate; this data-key
    // tripwire is the other layer of the defense-in-depth pair, so the
    // quoted synthetic prompt is rejected under `implement` exactly as
    // it is under an undeclared step name.
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({
        checklist_override: `User explicitly directed continuation via '${SYNTHETIC_COMPACTION_PROMPT}'`,
      }),
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("missing_human_decision")
    expect(result.retryable).toBe(false)
    // The rejection names the declared step and the override-carrying key.
    expect(result.message).toContain("implement")
    expect(result.message).toContain("checklist_override")
    // Bounded identifiers only — the quoted prompt is never echoed.
    expect(result.message).not.toContain("Continue if you have next steps")
    // MUST NOT persist progress state for that step.
    expect(readState(root, "opsx-change-name", "uf.unleash")).toBeNull()
  })

  it("rejects a prior-session free-text citation carried forward without its own ref", async () => {
    const root = await newRoot()
    // Spec scenario "Prior-session free-text override does not carry
    // forward": the prior session's override (pre-enforcement state on
    // disk) was free text with no human-decision reference, and no
    // decision was ever recorded for it.
    const priorText = "prior session override for Phase 1 remains in effect"
    writeState(root, "opsx-change-name", "uf.unleash", {
      command: "uf.unleash",
      branch: "opsx-change-name",
      key: "opsx-change-name",
      current_step: "spec-review",
      completed_steps: ["clarify", "plan", "tasks", "spec-review"],
      data: { checklist_override: priorText },
      created_at: "2026-11-14T00:00:00.000Z",
      updated_at: "2026-11-14T00:00:00.000Z",
    })
    expect(readHumanDecisionRecord(root, "opsx/change-name")).toEqual([])

    // The new session cites the prior session's free text as its
    // justification without its own human_decision_ref — a citation is
    // not authorization (SYNTH-001).
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({
        checklist_override: `carried forward from the prior session: "${priorText}"`,
      }),
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("missing_human_decision")
    expect(result.retryable).toBe(false)
    expect(result.message).toContain("implement")
    expect(result.message).toContain("checklist_override")
    // The prior session's free text is justification text, never echoed.
    expect(result.message).not.toContain("prior session override")
    // The rejected call persists nothing: the prior state stands exactly
    // as seeded — no step added, no reference stamped over the free text.
    const state = readState(root, "opsx-change-name", "uf.unleash")
    expect(state?.current_step).toBe("spec-review")
    expect(state?.completed_steps).toEqual([
      "clarify",
      "plan",
      "tasks",
      "spec-review",
    ])
    expect(state?.data.checklist_override).toBe(priorText)
    expect(state?.data.human_decision_ref).toBeUndefined()
  })

  it("detects an override key at any nesting depth (objects and arrays, case-insensitive)", async () => {
    const root = await newRoot()
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({
        phases: [
          { name: "phase-8", gates: { qualityOverride: "skip" } },
          { name: "phase-9" },
        ],
      }),
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("missing_human_decision")
    // The reported key path names the array index and the nested object walk.
    expect(result.message).toContain("phases[0].gates.qualityOverride")
    expect(readState(root, "opsx-change-name", "uf.unleash")).toBeNull()
  })

  it("rejects a free-text human_decision_ref quoting the synthetic prompt before any record lookup", async () => {
    const root = await newRoot()
    // A corrupted record would surface as human_decision_record_unreadable
    // if the lookup ran; the shape rejection must fire first.
    await mkdir(join(root, ".uf/progress", "opsx-change-name"), { recursive: true })
    await writeFile(recordPath(root, "opsx-change-name"), "{ not json")

    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({ checklist_override: "proceed" }),
      human_decision_ref: `User said: ${SYNTHETIC_COMPACTION_PROMPT}`,
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("missing_human_decision")
    expect(result.message).toContain("UUID v4")
    // The supplied free text is never echoed — bounded identifiers only.
    expect(result.message).not.toContain("Continue if you have next steps")
  })

  it("rejects an oversize human_decision_ref before any record lookup", async () => {
    const root = await newRoot()
    // A corrupted record would surface as human_decision_record_unreadable
    // if the lookup ran; the shape rejection must fire first — the same
    // ordering proof as the free-text variant above, applied to the
    // length bound.
    await mkdir(join(root, ".uf/progress", "opsx-change-name"), { recursive: true })
    await writeFile(recordPath(root, "opsx-change-name"), "{ not json")
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({ checklist_override: "proceed" }),
      human_decision_ref: "a".repeat(HUMAN_DECISION_REF_MAX_LENGTH + 1),
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("missing_human_decision")
    expect(result.message).toContain(`${HUMAN_DECISION_REF_MAX_LENGTH} characters`)
  })

  it("rejects a fabricated ref naming no recorded entry identically (missing_human_decision)", async () => {
    const root = await newRoot()
    const stamped = await appendHumanDecisions(root, "opsx/change-name", [
      relay("Proceed past the gate?", "Yes, proceed."),
    ])
    const fabricated = "ffffffff-ffff-4fff-8fff-ffffffffffff"
    expect(fabricated).not.toBe(stamped[0]?.decision_id)

    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({ checklist_override: "proceed" }),
      human_decision_ref: fabricated,
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("missing_human_decision")
    expect(result.retryable).toBe(false)
    // The shape-validated ref is a bounded decision_id-shaped identifier.
    expect(result.message).toContain(fabricated)
    expect(result.message).toContain("implement")
    expect(result.message).toContain("checklist_override")
    expect(readState(root, "opsx-change-name", "uf.unleash")).toBeNull()
  })

  it("treats an absent record file as the empty-record case (missing_human_decision, not unreadable)", async () => {
    const root = await newRoot()
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({ checklist_override: "proceed" }),
      human_decision_ref: "11111111-1111-4111-8111-111111111111",
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    // First use: no entry can resolve, so the rejection is a missing
    // reference — the record being absent is not a read failure.
    expect(result.code).toBe("missing_human_decision")
    expect(readState(root, "opsx-change-name", "uf.unleash")).toBeNull()
  })

  it("fails closed with human_decision_record_unreadable when the record is corrupted", async () => {
    const root = await newRoot()
    await mkdir(join(root, ".uf/progress", "opsx-change-name"), { recursive: true })
    const corrupted = "{ not json"
    await writeFile(recordPath(root, "opsx-change-name"), corrupted)

    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({ checklist_override: "proceed" }),
      human_decision_ref: "11111111-1111-4111-8111-111111111111",
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("human_decision_record_unreadable")
    expect(result.retryable).toBe(false)
    expect(result.message).toContain("implement")
    expect(result.message).toContain("checklist_override")
    // Nothing persisted and the corrupted audit trail is never touched.
    expect(readState(root, "opsx-change-name", "uf.unleash")).toBeNull()
    expect(await readFile(recordPath(root, "opsx-change-name"), "utf8")).toBe(corrupted)
  })

  it("fails closed with human_decision_record_unreadable on a non-ENOENT read failure (EISDIR)", async () => {
    const root = await newRoot()
    // A directory at the record path: lstat succeeds (not a symlink) and the
    // read fails with EISDIR — a non-ENOENT failure that must fail closed
    // rather than pass for a missing record.
    await mkdir(recordPath(root, "opsx-change-name"), { recursive: true })

    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({ checklist_override: "proceed" }),
      human_decision_ref: "11111111-1111-4111-8111-111111111111",
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("human_decision_record_unreadable")
  })

  it("accepts a gate override with a valid ref and persists the reference as provenance", async () => {
    const root = await newRoot()
    // Both write surfaces resolve identically at checkpoint time: the ref
    // names a decision_id, whichever surface recorded it.
    const inputs = [
      questionTool("Skip the checklist gate for phase 8?", "Yes, skip it."),
      relay("Relayed: proceed past the gate?", "Proceed."),
    ] as const
    for (const entry of inputs) {
      const stamped = await appendHumanDecisions(root, "opsx/change-name", [entry])
      const ref = stamped[0]?.decision_id
      expect(ref).toBeDefined()
      expect(HUMAN_DECISION_ID_PATTERN.test(ref ?? "")).toBe(true)

      const result = await invokeCheckpoint(root, {
        command: "uf.unleash",
        step: "checklist-override-phase8-9",
        data: JSON.stringify({ checklist_override: "human approved" }),
        human_decision_ref: ref,
      })

      expect(result.status, entry.source).toBe("ok")
      const state = readState(root, "opsx-change-name", "uf.unleash")
      expect(state?.completed_steps).toContain("checklist-override-phase8-9")
      expect(state?.data.checklist_override).toBe("human approved")
      expect(state?.data.human_decision_ref).toBe(ref)
    }
  })

  it("overwrites a caller-supplied data.human_decision_ref with the machinery-stamped reference", async () => {
    const root = await newRoot()
    const stamped = await appendHumanDecisions(root, "opsx/change-name", [
      relay("Proceed?", "Yes."),
    ])
    const ref = stamped[0]?.decision_id
    expect(ref).toBeDefined()

    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify({
        checklist_override: "proceed",
        // A smuggled same-named data key cannot spoof the provenance the
        // gate verified — the machinery stamp is applied over the data.
        human_decision_ref: "00000000-0000-4000-8000-000000000000",
      }),
      human_decision_ref: ref,
    })

    expect(result.status).toBe("ok")
    const state = readState(root, "opsx-change-name", "uf.unleash")
    expect(state?.data.human_decision_ref).toBe(ref)
  })

  it("leaves non-override data untouched (no reference required)", async () => {
    const root = await newRoot()
    const data = {
      checklist: "green",
      reason: "ordinary merge",
      bypass_note: "not an override key",
      nested: { phase8: [{ notes: "inner" }] },
    }
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify(data),
    })

    expect(result.status).toBe("ok")
    const state = readState(root, "opsx-change-name", "uf.unleash")
    expect(state?.data).toEqual(data)
    expect(state?.data.human_decision_ref).toBeUndefined()
  })

  it("treats non-object data as carrying no keys (behavior unchanged)", async () => {
    const root = await newRoot()
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: JSON.stringify("plain string"),
    })

    expect(result.status).toBe("ok")
  })

  it("does not require a ref for the HARDEN-002 spec_review_decisions override surface", async () => {
    const root = await gateRoot()
    // An all-override adjudication with NO data override key and NO
    // human_decision_ref: the HARDEN-002 surface carries its own override
    // justification (the per-finding reason) and must keep advancing.
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "spec-review",
      key: "opsx-change-name",
      spec_review_decisions: [
        { finding_id: "1", decision: "override", reason: "accepted risk" },
        { finding_id: "2", decision: "override", reason: "out of scope" },
      ],
    })

    expect(result.status).toBe("ok")
    const state = readState(root, "opsx-change-name", "uf.unleash")
    expect(state?.completed_steps).toContain("spec-review")
  })

  it("a rejected override call persists nothing, including spec_review_decisions", async () => {
    const root = await gateRoot()
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "spec-review",
      key: "opsx-change-name",
      data: JSON.stringify({ checklist_override: "proceed" }),
      spec_review_decisions: [
        { finding_id: "1", decision: "override", reason: "accepted risk" },
        { finding_id: "2", decision: "override", reason: "out of scope" },
      ],
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("missing_human_decision")
    // The SYNTH-001 rejection precedes the HARDEN-002 decision recording:
    // an unauthorized call leaves no adjudication behind.
    const state = readState(root, "opsx-change-name", "uf.unleash")
    const record = state?.data[SPEC_REVIEW_RECORD_KEY] as SpecReviewGateRecord | undefined
    expect(record?.decisions).toBeUndefined()
    expect(state?.completed_steps).not.toContain("spec-review")
  })

  it("composes with the HARDEN-002 gate: a valid ref does not bypass the gate hold", async () => {
    const root = await gateRoot()
    const stamped = await appendHumanDecisions(root, "opsx/change-name", [
      relay("Override the checklist gate?", "Yes."),
    ])
    const ref = stamped[0]?.decision_id
    expect(ref).toBeDefined()

    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      key: "opsx-change-name",
      data: JSON.stringify({ checklist_override: "proceed" }),
      human_decision_ref: ref,
    })

    // SYNTH-001 resolved (the ref is valid and recorded), then HARDEN-002
    // refuses the advancing step — a human_decision_ref does not bypass
    // that gate, whose own adjudication path is spec_review_decisions.
    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("missing_per_finding_decisions")
    expect(readState(root, "opsx-change-name", "uf.unleash")?.completed_steps).not.toContain(
      "implement",
    )
  })
})

// ---------------------------------------------------------------------------
// Declared checkpoint step enforcement (SYNTH-003, task 4.1)
// ---------------------------------------------------------------------------
//
// The checkpoint tool's command-keyed step predicate: under command
// uf.unleash a step name must be declared — the union of the pipeline's
// user-facing steps, the plugin's auto-checkpointed machinery steps, and the
// command prose's gate/budget steps, plus the implement-phase-N family — and
// an undeclared step rejects with undeclared_step unless a valid
// human_decision_ref per SYNTH-001 authorizes it. Other commands carry
// their own step vocabularies and are unaffected. The step-name length
// bound (CHECKPOINT_STEP_NAME_MAX_LENGTH) rejects before any
// security-sensitive operation. This suite is the task 4.1 wiring; the
// comprehensive SYNTH-003 matrix is task 4.3.
// ---------------------------------------------------------------------------
describe("checkpoint declared-step enforcement (SYNTH-003, task 4.1)", () => {
  it("accepts every declared step name, including the implement-phase-N family", () => {
    for (const step of UNLEASH_DECLARED_STEP_ALLOWLIST) {
      expect(isDeclaredUnleashCheckpointStep(step), step).toBe(true)
    }
    // The implement-phase-N family: N a positive integer written
    // canonically (the phase machinery records String(number), never a
    // zero-padded or suffixed form).
    for (const step of ["implement-phase-1", "implement-phase-2", "implement-phase-10"]) {
      expect(isDeclaredUnleashCheckpointStep(step), step).toBe(true)
    }
    // Negative pattern space: N must be >= 1 with no leading zero and no
    // trailing characters.
    for (const step of [
      "implement-phase-0",
      "implement-phase-",
      "implement-phase-01",
      "implement-phase-2x",
    ]) {
      expect(isDeclaredUnleashCheckpointStep(step), step).toBe(false)
    }
  })

  it("rejects an undeclared step under uf.unleash with undeclared_step naming the step", async () => {
    const root = await newRoot()
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "ad-hoc-undocumented-step",
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("undeclared_step")
    expect(result.retryable).toBe(false)
    expect(result.message).toContain("ad-hoc-undocumented-step")
    expect(result.message).toContain("uf.unleash")
    // Rejected before any state write: no progress persists for the step.
    expect(readState(root, "opsx-change-name", "uf.unleash")).toBeNull()
  })

  it("accepts an undeclared step carrying a valid human_decision_ref (the consented escape hatch)", async () => {
    const root = await newRoot()
    const stamped = await appendHumanDecisions(root, "opsx/change-name", [
      relay("Record the ad-hoc step under a new name?", "Yes, record it."),
    ])
    const ref = stamped[0]?.decision_id
    expect(ref).toBeDefined()

    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "ad-hoc-undocumented-step",
      data: JSON.stringify({ note: "ordinary data, no override key" }),
      human_decision_ref: ref,
    })

    expect(result.status).toBe("ok")
    const state = readState(root, "opsx-change-name", "uf.unleash")
    expect(state?.completed_steps).toContain("ad-hoc-undocumented-step")
    // The authorizing reference is stamped as provenance — the same
    // anti-spoofing discipline SYNTH-001 applies to gate overrides.
    expect(state?.data.human_decision_ref).toBe(ref)
    // Ordinary non-override data rides along untouched.
    expect(state?.data.note).toBe("ordinary data, no override key")
  })

  it("does not apply the uf.unleash predicate to another command's checkpoint", async () => {
    const root = await newRoot()
    // Spec scenario "Another command's checkpoint is unaffected":
    // uf.review-council carries its own step vocabulary.
    const result = await invokeCheckpoint(root, {
      command: "uf.review-council",
      step: "phase-1a-preflight",
    })

    expect(result.status).toBe("ok")
    const state = readState(root, "opsx-change-name", "uf.review-council")
    expect(state?.completed_steps).toContain("phase-1a-preflight")
  })

  it("rejects an oversize step name before any record lookup", async () => {
    const root = await newRoot()
    // A corrupted record would fail the shared resolver closed
    // (human_decision_record_unreadable) if any gate's lookup ran; the
    // oversize rejection must fire first — the same ordering proof the
    // SYNTH-001 shape tests use, applied to the step-name bound, which
    // precedes both the declared-step predicate and the override tripwire.
    await mkdir(join(root, ".uf/progress", "opsx-change-name"), { recursive: true })
    await writeFile(recordPath(root, "opsx-change-name"), "{ not json")
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "x".repeat(CHECKPOINT_STEP_NAME_MAX_LENGTH + 1),
      data: JSON.stringify({ checklist_override: "proceed" }),
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("oversize_step")
    expect(result.retryable).toBe(false)
    expect(result.message).toContain(`${CHECKPOINT_STEP_NAME_MAX_LENGTH} characters`)
    // The offending step is named, truncated to the bound.
    expect(result.message).toContain("...")
    // Neither the override tripwire nor the declared-step predicate ran —
    // their envelopes would have named other codes and other identifiers.
    expect(result.message).not.toContain("checklist_override")
    expect(readState(root, "opsx-change-name", "uf.unleash")).toBeNull()
  })
})
