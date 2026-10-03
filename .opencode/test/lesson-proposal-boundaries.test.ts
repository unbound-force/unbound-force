import { createHash } from "node:crypto"

import { describe, expect, it } from "vitest"

import {
  LESSON_PROPOSAL_BEGIN,
  LESSON_PROPOSAL_END,
  prepareLessonLearning,
  type AcquireSiblingEvidenceResult,
} from "../plugins/review-dispatch/index.js"

const COMMIT = "b".repeat(40)

function sha256(value: string): string {
  return createHash("sha256").update(value, "utf8").digest("hex")
}

function childOutput(value: unknown): string {
  return `${LESSON_PROPOSAL_BEGIN}\n${JSON.stringify(value)}\n${LESSON_PROPOSAL_END}`
}

function evidence(
  content: string,
  overrides: Partial<AcquireSiblingEvidenceResult["siblings"][number]> = {},
): AcquireSiblingEvidenceResult {
  const file = {
    sibling: "gaze",
    commit: COMMIT,
    path: "README.md",
    sha256: sha256(content),
    source_mode: "local" as const,
  }
  return {
    version: 1,
    status: "ok",
    prompt: [
      "<!-- uf-untrusted-sibling-evidence:v1 -->",
      JSON.stringify(file),
      content,
      "<!-- /uf-untrusted-sibling-evidence -->",
    ].join("\n"),
    siblings: [
      {
        sibling: "gaze",
        status: "available",
        commit: COMMIT,
        source_mode: "local",
        source_candidate: "/tmp/gaze",
        evidence: [file],
        rejections: [],
        unavailable_reason: null,
        ...overrides,
      },
    ],
    errors: [],
  }
}

function proposal(evidence: AcquireSiblingEvidenceResult, information = "Grounded reference.") {
  const metadata = evidence.siblings[0]!.evidence[0]!
  return {
    schema_version: "1.0.0",
    sibling: metadata.sibling,
    commit: metadata.commit,
    sources: [{ path: metadata.path, sha256: metadata.sha256, excerpt: contentOf(evidence) }],
    information,
  }
}

function contentOf(value: AcquireSiblingEvidenceResult): string {
  return value.prompt.split("\n")[2] ?? ""
}

describe("lesson proposal acceptance error boundaries", () => {
  it("fails grounding when acquisition was not accepted", () => {
    const result = prepareLessonLearning(
      childOutput(proposal(evidence("observable behavior"))),
      { ...evidence("observable behavior"), status: "invalid-config" },
      [],
    )
    expect(result).toMatchObject({ status: "skipped", reason_code: "grounding-unavailable" })
  })

  it("fails grounding when evidence metadata disagrees with its repository result", () => {
    const base = evidence("observable behavior")
    const sibling = { ...base.siblings[0]!, commit: "a".repeat(40) }
    expect(
      prepareLessonLearning(childOutput(proposal(base)), { ...base, siblings: [sibling] }, []),
    ).toMatchObject({ status: "skipped", reason_code: "grounding-unavailable" })
  })

  it("fails grounding on duplicated evidence metadata", () => {
    const base = evidence("observable behavior")
    const sibling = { ...base.siblings[0]! }
    sibling.evidence = [...sibling.evidence, { ...sibling.evidence[0]! }]
    expect(
      prepareLessonLearning(childOutput(proposal(base)), { ...base, siblings: [sibling] }, []),
    ).toMatchObject({ status: "skipped", reason_code: "grounding-unavailable" })
  })

  it("fails grounding on unexpected prompt content outside delimiters", () => {
    const base = evidence("observable behavior")
    expect(
      prepareLessonLearning(childOutput(proposal(base)), { ...base, prompt: `stray\n${base.prompt}` }, []),
    ).toMatchObject({ status: "skipped", reason_code: "grounding-unavailable" })
  })

  it("fails grounding on malformed evidence metadata", () => {
    const base = evidence("observable behavior")
    const malformed = "<!-- uf-untrusted-sibling-evidence:v1 -->\nnot-json\nobservable\n<!-- /uf-untrusted-sibling-evidence -->"
    expect(
      prepareLessonLearning(childOutput(proposal(base)), { ...base, prompt: malformed }, []),
    ).toMatchObject({ status: "skipped", reason_code: "grounding-unavailable" })
  })

  it("fails grounding when prompt metadata is not in the accepted set", () => {
    const base = evidence("observable behavior")
    const foreign = JSON.stringify({ sibling: "other", commit: "a".repeat(40), path: "x.md", sha256: sha256("x"), source_mode: "local" })
    const prompt = ["<!-- uf-untrusted-sibling-evidence:v1 -->", foreign, "x", "<!-- /uf-untrusted-sibling-evidence -->"].join("\n")
    expect(
      prepareLessonLearning(childOutput(proposal(base)), { ...base, prompt }, []),
    ).toMatchObject({ status: "skipped", reason_code: "grounding-unavailable" })
  })

  it("fails grounding when accepted evidence content is incomplete", () => {
    const base = evidence("observable behavior")
    expect(prepareLessonLearning(childOutput(proposal(base)), { ...base, prompt: "" }, [])).toMatchObject({
      status: "skipped",
      reason_code: "grounding-unavailable",
    })
  })

  it("skips empty information after normalization", () => {
    const base = evidence("observable behavior")
    expect(
      prepareLessonLearning(childOutput(proposal(base, "   \t\r\n  ")), base, []),
    ).toMatchObject({ status: "skipped", reason_code: "information-empty-after-normalization" })
  })

  it("rejects duplicate parent dedupe identities", () => {
    const base = evidence("observable behavior")
    const hash = "a".repeat(64)
    expect(
      prepareLessonLearning(childOutput(proposal(base)), base, [hash, hash]),
    ).toMatchObject({ status: "skipped", reason_code: "schema-invalid" })
  })

  it("rejects an unpaired trailing surrogate as a malformed section", () => {
    const base = evidence("observable behavior")
    const output = `${LESSON_PROPOSAL_BEGIN}\n{"information":"\udc00"}\n${LESSON_PROPOSAL_END}`
    expect(prepareLessonLearning(output, base, [])).toMatchObject({
      status: "skipped",
      reason_code: "malformed-section",
    })
  })
})
