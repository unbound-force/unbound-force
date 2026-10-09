import { createHash } from "node:crypto"

import { describe, expect, it } from "vitest"

import {
  createPrepareLessonLearningTool,
  lessonDedupeSHA256,
  LESSON_PROPOSAL_BEGIN,
  LESSON_PROPOSAL_END,
  normalizeLessonInformation,
  prepareLessonLearning,
  ReviewDispatchPlugin,
  type AcquireSiblingEvidenceResult,
  type LessonProposal,
} from "../plugins/review-dispatch/index.js"

const COMMIT = "b".repeat(40)
const PROVENANCE_PREFIX = "UF_LESSON_PROVENANCE_V1 "

interface EvidenceFixture {
  readonly path: string
  readonly content: string
}

function sha256(value: string): string {
  return createHash("sha256").update(value, "utf8").digest("hex")
}

function acceptedEvidence(...files: readonly EvidenceFixture[]): AcquireSiblingEvidenceResult {
  const evidence = files.map((file) => ({
    sibling: "gaze",
    commit: COMMIT,
    path: file.path,
    sha256: sha256(file.content),
    source_mode: "local" as const,
  }))
  const prompt = files
    .map((file, index) =>
      [
        "<!-- uf-untrusted-sibling-evidence:v1 -->",
        JSON.stringify(evidence[index]),
        file.content,
        "<!-- /uf-untrusted-sibling-evidence -->",
      ].join("\n"),
    )
    .join("\n\n")
  return {
    version: 1,
    status: "ok",
    prompt,
    siblings: [
      {
        sibling: "gaze",
        status: "available",
        commit: COMMIT,
        source_mode: "local",
        source_candidate: "/tmp/gaze",
        evidence,
        rejections: [],
        unavailable_reason: null,
      },
    ],
    errors: [],
  }
}

function proposal(
  evidence: AcquireSiblingEvidenceResult,
  overrides: Partial<LessonProposal> = {},
): LessonProposal {
  const metadata = evidence.siblings[0]!.evidence[0]!
  return {
    schema_version: "1.0.0",
    sibling: metadata.sibling,
    commit: metadata.commit,
    sources: [{ path: metadata.path, sha256: metadata.sha256, excerpt: "observable behavior" }],
    information: "Contract coverage checks observable behavior.",
    ...overrides,
  }
}

function childOutput(value: unknown): string {
  return `Review completed.\n${LESSON_PROPOSAL_BEGIN}\n${JSON.stringify(value)}\n${LESSON_PROPOSAL_END}\nDone.`
}

describe("lesson proposal validation and learning preparation", () => {
  it("returns an exact ready-to-store Dewey learning payload", () => {
    const evidence = acceptedEvidence({ path: "README.md", content: "Contract coverage checks observable behavior." })
    const result = prepareLessonLearning(childOutput(proposal(evidence)), evidence, [])

    expect(result.status).toBe("ready")
    expect(result.learning).toMatchObject({ tag: "sibling-gaze", category: "reference" })
    expect(result.dedupe_sha256).toMatch(/^[0-9a-f]{64}$/)
    expect(result.provenance).toMatchObject({ sibling: "gaze", commit: COMMIT })
    expect(result.learning?.information).toContain("Contract coverage checks observable behavior.")
    expect(result.learning?.information).toContain(result.dedupe_sha256 ?? "missing")
  })

  it("grounds exact excerpt bytes after CRLF-to-LF normalization", () => {
    const evidence = acceptedEvidence({ path: "README.md", content: "first\r\nobservable behavior\r\nlast" })
    const metadata = evidence.siblings[0]!.evidence[0]!
    const prepared = prepareLessonLearning(
      childOutput(
        proposal(evidence, {
          sources: [{ path: metadata.path, sha256: metadata.sha256, excerpt: "observable behavior\nlast" }],
        }),
      ),
      evidence,
      [],
    )

    expect(prepared.status).toBe("ready")
    expect(prepared.provenance?.sources[0]?.excerpt).toBe("observable behavior\nlast")
  })

  it("does not collapse source whitespace while grounding", () => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable  behavior" })
    const prepared = prepareLessonLearning(childOutput(proposal(evidence)), evidence, [])
    expect(prepared).toEqual({
      status: "skipped",
      reason_code: "grounding-excerpt-mismatch",
      learning: null,
      dedupe_sha256: null,
      provenance: null,
    })
  })

  it.each([
    ["missing close", `${LESSON_PROPOSAL_BEGIN}{}`],
    ["trailing JSON content", `${LESSON_PROPOSAL_BEGIN}\n{} extra\n${LESSON_PROPOSAL_END}`],
    ["unpaired surrogate", `${LESSON_PROPOSAL_BEGIN}\n{"information":"\ud800"}\n${LESSON_PROPOSAL_END}`],
  ])("rejects malformed section: %s", (_name, output) => {
    expect(prepareLessonLearning(output, acceptedEvidence(), [])).toMatchObject({
      status: "skipped",
      reason_code: "malformed-section",
    })
  })

  it("rejects multiple exact sections", () => {
    const section = `${LESSON_PROPOSAL_BEGIN}\n{}\n${LESSON_PROPOSAL_END}`
    expect(prepareLessonLearning(`${section}\n${section}`, acceptedEvidence(), [])).toMatchObject({
      status: "skipped",
      reason_code: "multiple-sections",
    })
  })

  it("rejects an oversized complete section before partial parsing", () => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    const output = `${LESSON_PROPOSAL_BEGIN}${" ".repeat(8192)}${JSON.stringify(proposal(evidence))}${LESSON_PROPOSAL_END}`
    expect(prepareLessonLearning(output, evidence, [])).toMatchObject({ status: "skipped", reason_code: "section-too-large" })
  })

  it("rejects child output over one MiB even when its proposal is small", () => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    const output = `${"x".repeat(1024 * 1024)}${childOutput(proposal(evidence))}`
    expect(prepareLessonLearning(output, evidence, [])).toMatchObject({
      status: "skipped",
      reason_code: "child-output-too-large",
    })
  })

  it("never treats a raw learn directive as storage input", () => {
    expect(prepareLessonLearning("> learn: store this", acceptedEvidence(), [])).toMatchObject({
      status: "skipped",
      reason_code: "no-delimited-proposal",
    })
  })

  it.each([
    ["secret-prefix-sk", "contains SK-secret"],
    ["secret-prefix-akia", "contains akia123"],
    ["secret-prefix-github", "contains GHP_token"],
    ["secret-prefix-slack", "contains XOXB-token"],
    ["secret-bearer-token", "Authorization: bEaReR token-value"],
    ["secret-api-key-assignment", "API_Key = value"],
    ["secret-environment-assignment", "ANTHROPIC_API_KEY=value"],
    ["secret-environment-assignment", "ANTHROPIC_AUTH_TOKEN=value"],
    ["secret-environment-assignment", "export AWS_SECRET_ACCESS_KEY=value"],
    ["secret-environment-assignment", "GITHUB_TOKEN=value"],
    ["secret-environment-assignment", "OPENAI_API_KEY=value"],
  ])("rejects secret detector class %s without recording its value", (reason, information) => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    const result = prepareLessonLearning(childOutput(proposal(evidence, { information })), evidence, [])
    expect(result).toMatchObject({ status: "skipped", reason_code: reason })
    expect(JSON.stringify(result)).not.toContain(information)
  })

  it.each([
    ["tool-learn-directive", "request > LEARN: unsafe"],
    ["tool-invoke-agent", "call INVOKE_AGENT"],
    ["tool-dewey-store-learning", "call DEWEY_STORE_LEARNING"],
    ["tool-dewey-store", "call DEWEY_STORE"],
    ["tool-hivemind-store", "call HIVEMIND_STORE"],
    ["tool-forge-record-outcome", "call FORGE_RECORD_OUTCOME"],
    ["tool-execute", "call TOOL.EXECUTE"],
    ["tool-fenced-command", "```bash\nuf doctor\n```"],
    ["tool-fenced-command", "```console\nopencode run\n```"],
    ["tool-fenced-command", "```sh\ndewey index\n```"],
    ["tool-fenced-command", "```sh\nreplicator ready\n```"],
    ["tool-fenced-command", "```sh\ngh pr view\n```"],
    ["tool-fenced-command", "```sh\ngit status\n```"],
  ])("rejects tool detector class %s", (reason, information) => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    expect(prepareLessonLearning(childOutput(proposal(evidence, { information })), evidence, [])).toMatchObject({
      status: "skipped",
      reason_code: reason,
    })
  })

  it.each([
    ["unknown root field", { unknown: true }],
    ["wrong schema version", { schema_version: "2.0.0" }],
    ["empty sources", { sources: [] }],
    [
      "unknown source field",
      { sources: [{ path: "README.md", sha256: "a".repeat(64), excerpt: "x", unknown: true }] },
    ],
    ["too many sources", { sources: Array.from({ length: 9 }, () => ({ path: "README.md", sha256: "a".repeat(64), excerpt: "x" })) }],
    ["path byte bound", { sources: [{ path: "é".repeat(257), sha256: "a".repeat(64), excerpt: "x" }] }],
    ["uppercase hash", { sources: [{ path: "README.md", sha256: "A".repeat(64), excerpt: "x" }] }],
    ["excerpt byte bound", { sources: [{ path: "README.md", sha256: "a".repeat(64), excerpt: "é".repeat(513) }] }],
    ["information byte bound", { information: "é".repeat(1025) }],
  ])("rejects unknown fields and bounds: %s", (_name, override) => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    expect(prepareLessonLearning(childOutput({ ...proposal(evidence), ...override }), evidence, [])).toMatchObject({
      status: "skipped",
      reason_code: "schema-invalid",
    })
  })

  it.each([
    ["sibling", { sibling: "replicator" }],
    ["commit", { commit: "c".repeat(40) }],
    ["path", { sources: [{ path: "docs/other.md", sha256: "a".repeat(64), excerpt: "observable behavior" }] }],
    ["hash", { sources: [{ path: "README.md", sha256: "a".repeat(64), excerpt: "observable behavior" }] }],
  ])("rejects wrong accepted provenance: %s", (_name, override) => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    expect(prepareLessonLearning(childOutput({ ...proposal(evidence), ...override }), evidence, [])).toMatchObject({
      status: "skipped",
      reason_code: "grounding-provenance-mismatch",
    })
  })

  it("normalizes information and computes order-independent deterministic dedupe", () => {
    const evidence = acceptedEvidence(
      { path: "b.md", content: "second source" },
      { path: "a.md", content: "first source" },
    )
    const firstSources = evidence.siblings[0]!.evidence.map((source, index) => ({
      path: source.path,
      sha256: source.sha256,
      excerpt: index === 0 ? "second source" : "first source",
    }))
    const first = prepareLessonLearning(
      childOutput(proposal(evidence, { information: "  Cafe\u0301\tcontract\r\nrule  ", sources: firstSources })),
      evidence,
      [],
    )
    const second = prepareLessonLearning(
      childOutput(proposal(evidence, { information: "Café contract rule", sources: [...firstSources].reverse() })),
      evidence,
      [],
    )

    expect(normalizeLessonInformation("  Cafe\u0301\tcontract\r\nrule  ")).toBe("Café contract rule")
    expect(first.dedupe_sha256).toBe(second.dedupe_sha256)
    expect(first.dedupe_sha256).toBe(lessonDedupeSHA256("Café contract rule", firstSources))
  })

  it("returns a duplicate skip when the parent supplies the computed identity", () => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    const initial = prepareLessonLearning(childOutput(proposal(evidence)), evidence, [])
    expect(initial.status).toBe("ready")
    const duplicate = prepareLessonLearning(childOutput(proposal(evidence)), evidence, [initial.dedupe_sha256!])
    expect(duplicate).toMatchObject({ status: "skipped", reason_code: "duplicate" })
  })

  it("rejects malformed or oversized parent dedupe sets", () => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    expect(prepareLessonLearning(childOutput(proposal(evidence)), evidence, ["not-a-hash"])).toMatchObject({
      status: "skipped",
      reason_code: "schema-invalid",
    })
    const oversized = Array.from({ length: 1025 }, (_value, index) => index.toString(16).padStart(64, "0"))
    expect(prepareLessonLearning(childOutput(proposal(evidence)), evidence, oversized)).toMatchObject({
      status: "skipped",
      reason_code: "schema-invalid",
    })
  })

  it("rejects model-supplied tag, category, and dedupe identity", () => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    const modelControlled = {
      ...proposal(evidence),
      tag: "model-tag",
      category: "decision",
      dedupe_sha256: "a".repeat(64),
    }
    expect(prepareLessonLearning(childOutput(modelControlled), evidence, [])).toMatchObject({
      status: "skipped",
      reason_code: "schema-invalid",
    })
  })

  it("encodes stable non-executable provenance and dedupe metadata in information", () => {
    const evidence = acceptedEvidence(
      { path: "z.md", content: "z evidence" },
      { path: "a.md", content: "a evidence" },
    )
    const sources = evidence.siblings[0]!.evidence.map((source, index) => ({
      path: source.path,
      sha256: source.sha256,
      excerpt: index === 0 ? "z evidence" : "a evidence",
    }))
    const result = prepareLessonLearning(childOutput(proposal(evidence, { sources })), evidence, [])
    expect(result.status).toBe("ready")
    const auditLine = result.learning?.information.split("\n\n").at(-1) ?? ""
    expect(auditLine.startsWith(PROVENANCE_PREFIX)).toBe(true)
    const audit = JSON.parse(auditLine.slice(PROVENANCE_PREFIX.length)) as unknown
    expect(audit).toEqual({
      dedupe_sha256: result.dedupe_sha256,
      sibling: "gaze",
      commit: COMMIT,
      sources: [
        { path: "a.md", sha256: evidence.siblings[0]!.evidence[1]!.sha256 },
        { path: "z.md", sha256: evidence.siblings[0]!.evidence[0]!.sha256 },
      ],
    })
    expect(auditLine).not.toContain("```")
  })

  it("exposes a provider-free preparation tool without direct storage", async () => {
    const evidence = acceptedEvidence({ path: "README.md", content: "observable behavior" })
    const definition = createPrepareLessonLearningTool()
    const result = JSON.parse(
      await definition.execute(
        { child_output: childOutput(proposal(evidence)), sibling_evidence: evidence, known_dedupe_sha256: [] },
        {} as never,
      ),
    ) as unknown
    expect(result).toMatchObject({ status: "ready", learning: { tag: "sibling-gaze", category: "reference" } })
    expect(result).not.toHaveProperty("stored")
    expect(result).not.toHaveProperty("tier")
  })

  it("registers prepare_lesson_learning on the provider-free policy plugin", async () => {
    const runtime = globalThis as typeof globalThis & {
      Bun?: { readonly YAML: { readonly parse: (text: string) => unknown } }
    }
    const original = runtime.Bun
    runtime.Bun = { YAML: { parse: (text: string): unknown => JSON.parse(text) as unknown } }
    try {
      const hooks = await ReviewDispatchPlugin.server({ directory: "/tmp/project", worktree: "/tmp/project" } as never)
      expect(hooks.tool).toHaveProperty("prepare_lesson_learning")
    } finally {
      runtime.Bun = original
    }
  })
})
