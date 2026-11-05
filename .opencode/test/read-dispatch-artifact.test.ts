import { Buffer } from "node:buffer"
import { execFile } from "node:child_process"
import { mkdir, mkdtemp, rm, symlink, truncate, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { promisify } from "node:util"

import { afterEach, describe, expect, it } from "vitest"

import { DISPATCH_ARTIFACT_DIRECTORY } from "../lib/review-dispatch-artifact.js"
import { readDispatchArtifact } from "../plugins/review-dispatch/index.js"
import samplePayload from "../../schemas/review-dispatch/samples/sample-review-dispatch.json"

// RD-FR-057: consumers page through a finalized artifact with
// read_dispatch_artifact instead of the file-read tool, whose 2000-character
// line limit cuts long JSON strings such as full-length lesson proposals.

const execFileAsync = promisify(execFile)
const CORRELATION = samplePayload.correlation_id
const roots: string[] = []

afterEach(async () => {
  await Promise.all(roots.splice(0).map((root) => rm(root, { recursive: true, force: true })))
})

async function projectWith(payload: Record<string, unknown>, artifactType = "review-dispatch"): Promise<string> {
  const root = await mkdtemp(join(tmpdir(), "read-dispatch-artifact-"))
  roots.push(root)
  const directory = join(root, DISPATCH_ARTIFACT_DIRECTORY)
  await mkdir(directory, { recursive: true })
  const envelope = { hero: "the-divisor", artifact_type: artifactType, payload }
  await writeFile(join(directory, `${CORRELATION}.json`), `${JSON.stringify(envelope, null, 2)}\n`, "utf8")
  return root
}

function proposals(count: number, length: number) {
  return Array.from({ length: count }, (_, index) => ({
    agent: "divisor-guard",
    information: `${index}:`.padEnd(length, "x"),
    tag: "review-dispatch",
    category: "gotcha",
  }))
}

describe("read_dispatch_artifact (RD-FR-057)", () => {
  it("returns scalars and entry counts for the summary section", async () => {
    const root = await projectWith({ ...samplePayload, proposals: proposals(2, 40) })
    const summary = await readDispatchArtifact({ correlation_id: CORRELATION, section: "summary" }, root)
    expect(summary).toMatchObject({
      status: "ok",
      correlation_id: CORRELATION,
      section: "summary",
      command: samplePayload.command,
      mode: samplePayload.mode,
      verdict: samplePayload.verdict,
      verdict_reason: samplePayload.verdict_reason,
      run_counts: samplePayload.run_counts,
      coverage: samplePayload.coverage,
      workflow_result: samplePayload.workflow_result,
      counts: { findings: 1, advisories: 0, proposals: 2, runs: 1 },
    })
    // Pin the full summary shape: a regression dropping any documented
    // field (mode, verdict_reason, run_counts, coverage, workflow_result)
    // must fail, not just the subset above (typescript.md TC-003).
    expect(Object.keys(summary).sort()).toEqual(
      [
        "status",
        "correlation_id",
        "section",
        "command",
        "mode",
        "verdict",
        "verdict_reason",
        "run_counts",
        "coverage",
        "workflow_result",
        "counts",
      ].sort(),
    )
    expect(summary).not.toHaveProperty("entries")
  })

  it("pages full-length proposals longer than the file tool's 2000-character line limit", async () => {
    const all = proposals(7, 9_000)
    const root = await projectWith({ ...samplePayload, proposals: all })
    const seen: unknown[] = []
    let offset: number | null = 0
    let pages = 0
    while (offset !== null) {
      const page = await readDispatchArtifact({ correlation_id: CORRELATION, section: "proposals", offset }, root)
      expect(page.status).toBe("ok")
      expect(page.total).toBe(7)
      expect(Buffer.byteLength(JSON.stringify(page), "utf8")).toBeLessThan(40 * 1024)
      seen.push(...(page.entries as unknown[]))
      offset = page.next_offset as number | null
      pages += 1
    }
    expect(pages).toBeGreaterThan(1)
    expect(seen).toEqual(all)
    expect((seen[0] as { information: string }).information).toHaveLength(9_000)
  })

  it("serves one oversized entry alone so paging always advances", async () => {
    const root = await projectWith({ ...samplePayload, proposals: proposals(2, 16_384).map((p) => ({ ...p, information: "\u0001".repeat(16_384) })) })
    const first = await readDispatchArtifact({ correlation_id: CORRELATION, section: "proposals" }, root)
    expect(first.entries).toHaveLength(1)
    expect(first.next_offset).toBe(1)
  })

  it("omits run findings and reports finding_count", async () => {
    const root = await projectWith({ ...samplePayload })
    const page = await readDispatchArtifact({ correlation_id: CORRELATION, section: "runs" }, root)
    const run = (page.entries as Record<string, unknown>[])[0]
    expect(run).not.toHaveProperty("findings")
    expect(run.finding_count).toBe(samplePayload.runs[0].findings.length)
    expect(run.workflow_verdict).toBe(samplePayload.runs[0].workflow_verdict)
    expect(page.next_offset).toBeNull()
  })

  it("returns an empty page past the end and an empty proposals section for legacy artifacts", async () => {
    const root = await projectWith({ ...samplePayload })
    const legacy = await readDispatchArtifact({ correlation_id: CORRELATION, section: "proposals" }, root)
    expect(legacy).toMatchObject({ status: "ok", total: 0, entries: [], next_offset: null })
    const past = await readDispatchArtifact({ correlation_id: CORRELATION, section: "findings", offset: 5 }, root)
    expect(past).toMatchObject({ total: 1, offset: 5, entries: [], next_offset: null })
  })

  it("serves complete advisories entries and the full finding at an in-bounds offset", async () => {
    // The sample payload carries zero advisories, so a request against
    // it would only assert another empty page. Inject a valid entry
    // first, then require the served entry to be complete.
    const advisory = {
      severity: "LOW",
      description: "Advisories are served complete through the reader.",
      run_ids: [samplePayload.runs[0].run_id],
    }
    const root = await projectWith({ ...samplePayload, advisories: [advisory] })
    const advisories = await readDispatchArtifact({ correlation_id: CORRELATION, section: "advisories" }, root)
    expect(advisories).toMatchObject({ status: "ok", total: 1, offset: 0, next_offset: null })
    expect(advisories.entries).toEqual([advisory])

    // RD-FR-057: an in-bounds page serves the complete, untruncated
    // finding — every field, at its full length.
    const findings = await readDispatchArtifact({ correlation_id: CORRELATION, section: "findings", offset: 0 }, root)
    expect(findings).toMatchObject({ status: "ok", total: 1, offset: 0, next_offset: null })
    expect(findings.entries).toEqual([
      {
        severity: "HIGH",
        category: "correctness",
        description: "A required validation branch is missing.",
        root_cause: "Unvalidated external input",
        file: "internal/review/dispatch.go",
        line: 42,
        run_ids: ["11111111-1111-4111-8111-111111111111"],
      },
    ])
  })

  it("fails closed on invalid input, a missing artifact, and an invalid artifact", async () => {
    const root = await projectWith({ ...samplePayload })
    expect(await readDispatchArtifact({ correlation_id: "../etc", section: "summary" }, root)).toMatchObject({
      status: "error",
      code: "invalid_input",
    })
    expect(await readDispatchArtifact({ correlation_id: CORRELATION, section: "plan" }, root)).toMatchObject({
      code: "invalid_input",
    })
    expect(
      await readDispatchArtifact({ correlation_id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", section: "summary" }, root),
    ).toMatchObject({ code: "not_found" })

    const wrongType = await projectWith({ ...samplePayload }, "quality-report")
    expect(await readDispatchArtifact({ correlation_id: CORRELATION, section: "summary" }, wrongType)).toMatchObject({
      code: "invalid_artifact",
    })
    const badPayload = await projectWith({ ...samplePayload, proposals: [{ agent: "divisor-guard", information: "x", tag: "" }] })
    expect(await readDispatchArtifact({ correlation_id: CORRELATION, section: "summary" }, badPayload)).toMatchObject({
      code: "invalid_artifact",
    })
  })

  it("rejects unparseable JSON and a symlinked artifact", async () => {
    const root = await projectWith({ ...samplePayload })
    const path = join(root, DISPATCH_ARTIFACT_DIRECTORY, `${CORRELATION}.json`)
    await writeFile(path, "{ not json", "utf8")
    expect(await readDispatchArtifact({ correlation_id: CORRELATION, section: "summary" }, root)).toMatchObject({
      code: "invalid_artifact",
    })

    const linked = await projectWith({ ...samplePayload })
    const linkedPath = join(linked, DISPATCH_ARTIFACT_DIRECTORY, `${CORRELATION}.json`)
    const target = join(linked, "elsewhere.json")
    await writeFile(target, "{}", "utf8")
    await rm(linkedPath)
    await symlink(target, linkedPath)
    expect(await readDispatchArtifact({ correlation_id: CORRELATION, section: "summary" }, linked)).toMatchObject({
      code: "invalid_artifact",
    })
  })

  it("rejects an artifact larger than the 64 MiB bound", async () => {
    const root = await projectWith({ ...samplePayload })
    const path = join(root, DISPATCH_ARTIFACT_DIRECTORY, `${CORRELATION}.json`)
    // A sparse truncate past the bound keeps the test cheap: the stat
    // size exceeds 64 MiB without 64 MiB ever being written.
    await truncate(path, 64 * 1024 * 1024 + 1)
    expect(await readDispatchArtifact({ correlation_id: CORRELATION, section: "summary" }, root)).toMatchObject({
      status: "error",
      code: "invalid_artifact",
    })
  })

  it("rejects a FIFO artifact without blocking the open (RD-FR-057)", async () => {
    const root = await projectWith({ ...samplePayload })
    const path = join(root, DISPATCH_ARTIFACT_DIRECTORY, `${CORRELATION}.json`)
    // A writer-less FIFO would block a plain O_RDONLY open forever;
    // O_NONBLOCK plus the regular-file stat check must reject it.
    await rm(path)
    await execFileAsync("mkfifo", [path])
    expect(await readDispatchArtifact({ correlation_id: CORRELATION, section: "summary" }, root)).toMatchObject({
      status: "error",
      code: "invalid_artifact",
    })
  })
})
