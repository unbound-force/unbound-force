/**
 * Tests for the shared review-dispatch artifact parser (item J, round-5
 * review: the parser's type-guard branches had no direct coverage — they
 * were only exercised incidentally through the resume tools' happy paths).
 *
 * The parser is a pure string function (constitution Principle IV): every
 * malformed-envelope shape must return null so resume detection skips it
 * (fail-open to "not done"), and a missing/unparseable timestamp must sort
 * the artifact oldest rather than discarding the decision fields.
 */
import { describe, expect, it } from "vitest"

import {
  REVIEW_DISPATCH_ARTIFACT_TYPE,
  parseReviewDispatchArtifact,
} from "../lib/review-dispatch-artifact.js"
import { buildFailureRecordEnvelope } from "../plugins/review-dispatch/index.js"

/** Minimal valid envelope in the producer's shape (review-dispatch plugin). */
function envelope(overrides: Record<string, unknown> = {}): string {
  return JSON.stringify({
    artifact_type: REVIEW_DISPATCH_ARTIFACT_TYPE,
    context: { branch: "opsx/unleash-deterministic-tools" },
    payload: {
      mode: "code",
      verdict: "APPROVE WITH ADVISORIES",
      input_context: { head_sha: "f0327a9fb17c9da59f5b4ebcf7d24d5d37ad503a" },
    },
    timestamp: "2026-10-17T15:23:29Z",
    ...overrides,
  })
}

describe("parseReviewDispatchArtifact", () => {
  it("parses a valid envelope into the decision summary", () => {
    const summary = parseReviewDispatchArtifact(envelope(), "a.json")

    expect(summary).toEqual({
      source: "a.json",
      mode: "code",
      branch: "opsx/unleash-deterministic-tools",
      headSha: "f0327a9fb17c9da59f5b4ebcf7d24d5d37ad503a",
      verdict: "APPROVE WITH ADVISORIES",
      writtenAt: Date.parse("2026-10-17T15:23:29Z"),
    })
  })

  it("returns null for malformed JSON", () => {
    expect(parseReviewDispatchArtifact("{not json", "a.json")).toBeNull()
  })

  it("returns null for a non-record root", () => {
    expect(parseReviewDispatchArtifact("[1, 2, 3]", "a.json")).toBeNull()
    expect(parseReviewDispatchArtifact('"a string"', "a.json")).toBeNull()
    expect(parseReviewDispatchArtifact("null", "a.json")).toBeNull()
  })

  it("returns null for a foreign artifact_type", () => {
    expect(parseReviewDispatchArtifact(envelope({ artifact_type: "issue-triage" }), "a.json")).toBeNull()
  })

  it("returns null for a review-dispatch-failure envelope (resume detection ignores failure records)", () => {
    // Built with the producer's own envelope builder so the fixture cannot
    // drift from the real record shape: failure records share the dispatch
    // store directory with success artifacts, so resume detection will
    // actually meet one — the artifact_type discriminator alone keeps it
    // ignored (RD-FR-066, design D3).
    const failureRecord = buildFailureRecordEnvelope(
      {
        dispatch_correlation_id: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
        command: "review-council",
        mode: "code",
        status: "failed",
        operation_verdict: "INCONCLUSIVE",
        errors: ["validate review dispatch input: garbage"],
      },
      {
        branch: "opsx/review-dispatch-errors-cap-loses-unrecov",
        commit: "2222222222222222222222222222222222222222",
        workflow_id: "uf.review-council",
      },
      "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
      "2026-11-15T12:00:00.000Z",
    )

    expect(parseReviewDispatchArtifact(JSON.stringify(failureRecord), "a.failure.json")).toBeNull()
  })

  it("returns null when context is not a record", () => {
    expect(parseReviewDispatchArtifact(envelope({ context: "main" }), "a.json")).toBeNull()
    expect(parseReviewDispatchArtifact(envelope({ context: ["main"] }), "a.json")).toBeNull()
  })

  it("returns null when payload is not a record", () => {
    expect(parseReviewDispatchArtifact(envelope({ payload: "summary" }), "a.json")).toBeNull()
    expect(parseReviewDispatchArtifact(envelope({ payload: null }), "a.json")).toBeNull()
  })

  it("returns null when input_context is not a record", () => {
    const raw = envelope({
      payload: { mode: "code", verdict: "APPROVE", input_context: "f0327a9f" },
    })

    expect(parseReviewDispatchArtifact(raw, "a.json")).toBeNull()
  })

  it.each([
    ["mode", { mode: 42 }],
    ["verdict", { verdict: null }],
    ["head_sha", { input_context: { head_sha: 4096 } }],
  ])("returns null when payload %s is not a string", (_field, payloadOverrides) => {
    const raw = envelope({
      payload: {
        mode: "code",
        verdict: "APPROVE WITH ADVISORIES",
        input_context: { head_sha: "f0327a9fb17c9da59f5b4ebcf7d24d5d37ad503a" },
        ...payloadOverrides,
      },
    })

    expect(parseReviewDispatchArtifact(raw, "a.json")).toBeNull()
  })

  it("returns null when context.branch is not a string", () => {
    const raw = envelope({ context: { branch: ["opsx/unleash-deterministic-tools"] } })

    expect(parseReviewDispatchArtifact(raw, "a.json")).toBeNull()
  })

  it("sorts a missing timestamp oldest instead of skipping the artifact", () => {
    const withoutTimestamp = JSON.parse(envelope()) as Record<string, unknown>
    delete withoutTimestamp.timestamp

    const summary = parseReviewDispatchArtifact(JSON.stringify(withoutTimestamp), "a.json")

    expect(summary?.writtenAt).toBe(0)
    // The decision fields remain intact — the artifact is still valid.
    expect(summary?.verdict).toBe("APPROVE WITH ADVISORIES")
    expect(summary?.branch).toBe("opsx/unleash-deterministic-tools")
  })

  it("sorts an unparseable timestamp oldest instead of skipping the artifact", () => {
    const summary = parseReviewDispatchArtifact(envelope({ timestamp: "not-a-date" }), "a.json")

    expect(summary?.writtenAt).toBe(0)
    expect(summary?.mode).toBe("code")
    expect(summary?.headSha).toBe("f0327a9fb17c9da59f5b4ebcf7d24d5d37ad503a")
  })
})
