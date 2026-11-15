import { describe, expect, it } from "vitest"
import { z } from "zod"

import { formatValidationError } from "../lib/reviewer-manifest.js"
import {
  buildCompactFinalization,
  buildFullRetrieval,
  buildSummaryRetrieval,
  SUMMARY_LIMITS,
} from "../plugins/review-dispatch/index.js"

const RUN_COUNTS = {
  total: 4,
  success: 3,
  failed: 1,
  skipped: 0,
  budget_skipped: 0,
  limit_skipped: 0,
  cancelled: 0,
}

const COVERAGE = {
  preflight_verdict: "PASS",
  checks_total: 3,
  checks_passed: 3,
}

function makeFinding(
  index: number,
  overrides: Record<string, unknown> = {},
): Record<string, unknown> {
  return {
    severity: "MEDIUM",
    category: "test-quality",
    description: `finding ${index}`,
    root_cause: "root cause",
    file: `src/file-${index}.ts`,
    line: index + 1,
    run_ids: ["11111111-1111-4111-8111-111111111111"],
    ...overrides,
  }
}

function makeAdvisory(
  index: number,
  overrides: Record<string, unknown> = {},
): Record<string, unknown> {
  return {
    severity: "LOW",
    description: `advisory ${index}`,
    run_ids: ["11111111-1111-4111-8111-111111111111"],
    ...overrides,
  }
}

function makeProposal(
  index: number,
  overrides: Record<string, unknown> = {},
): Record<string, unknown> {
  return {
    agent: "divisor-guard",
    information: `proposal ${index}`,
    tag: "review-dispatch",
    category: "pattern",
    ...overrides,
  }
}

function makeWarning(type: string): Record<string, unknown> {
  switch (type) {
    case "orphaned_submissions_excluded":
      return {
        type,
        agent: "divisor-guard",
        run_id: "run-guard",
        discarded_findings: 2,
        discarded_proposals: 1,
      }
    case "invalid_submissions_excluded":
      return {
        type,
        agent: "divisor-guard",
        run_id: "run-guard",
        invalid_findings: 2,
        invalid_proposals: 1,
        errors: ["findings[1].severity: invalid"],
      }
    case "run_findings_overflow":
      return {
        type,
        agent: "divisor-guard",
        run_id: "run-guard",
        kept_findings: 512,
        overflow_findings: 4,
      }
    case "consolidated_findings_overflow":
      return { type, kept_findings: 512, dropped_findings: 5 }
    case "child_empty_response":
      return {
        type,
        agent: "divisor-guard",
        run_id: "run-guard",
        finish_reason: "length",
        tokens_input: 4096,
        tokens_output: 0,
      }
    default:
      return { type, agent: "divisor-guard", run_id: "run-guard" }
  }
}

function summaryInput(overrides: Record<string, unknown> = {}) {
  return {
    correlation_id: "11111111-1111-4111-8111-111111111111",
    proposals: [] as unknown[],
    warnings: [] as unknown[],
    verdict: "APPROVE",
    verdict_reason: "No findings requiring changes",
    run_counts: RUN_COUNTS,
    coverage: COVERAGE,
    findings: [] as unknown[],
    advisories: [] as unknown[],
    ...overrides,
  }
}

// Spec literals (RD-FR-049), asserted directly so raising SUMMARY_LIMITS
// cannot silently loosen the guard.
const BYTE_BUDGET = 40_960

function byteSize(value: unknown): number {
  return new TextEncoder().encode(JSON.stringify(value)).length
}

function fullPayload(overrides: Record<string, unknown> = {}) {
  return {
    command: "review-council",
    mode: "code",
    full: true,
    input_context: {
      kind: "local",
      pr_number: null,
      base_ref: "main",
      base_sha: "a".repeat(40),
      head_ref: "opsx/change",
      head_sha: "b".repeat(40),
      uncommitted: true,
    },
    change_profile: {
      kind: "diff",
      lines: 10,
      files: 1,
      components: 1,
      security_sensitive: false,
      user_facing: false,
      categories: ["standard"],
      tier: "lightweight",
    },
    plan_version: 1,
    correlation_id: "11111111-1111-4111-8111-111111111111",
    verdict: "APPROVE",
    verdict_reason: "ok",
    run_counts: RUN_COUNTS,
    coverage: COVERAGE,
    runs: [{ agent: "divisor-guard" }],
    plan: [{ agent: "divisor-guard" }],
    findings: [makeFinding(1)],
    advisories: [makeAdvisory(1)],
    workflow_result: { kind: "council", value: "APPROVE" },
    ...overrides,
  }
}

describe("summary retrieval", () => {
  it("emits the observability-first key order and never artifact_path", () => {
    const summary = buildSummaryRetrieval(
      summaryInput() as Parameters<typeof buildSummaryRetrieval>[0],
    )
    expect(Object.keys(summary)).toEqual([
      "correlation_id",
      "verdict",
      "verdict_reason",
      "run_counts",
      "coverage",
      "findings",
      "advisories",
    ])
    expect(summary).not.toHaveProperty("artifact_path")
    expect(summary).not.toHaveProperty("plan")
    expect(summary).not.toHaveProperty("runs")
  })

  it("places proposals and warnings before findings", () => {
    const summary = buildSummaryRetrieval(
      summaryInput({
        proposals: [makeProposal(1)],
        warnings: [makeWarning("empty_coverage_attestation")],
        findings: [makeFinding(1)],
        advisories: [makeAdvisory(1)],
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    const keys = Object.keys(summary)
    expect(keys[0]).toBe("correlation_id")
    expect(keys.indexOf("proposals")).toBeLessThan(keys.indexOf("warnings"))
    expect(keys.indexOf("warnings")).toBeLessThan(keys.indexOf("findings"))
    expect(keys.indexOf("proposals")).toBeLessThan(keys.indexOf("findings"))
  })

  it("keeps sixteen findings under the 40 KiB budget with the 15-entry cap", () => {
    const summary = buildSummaryRetrieval(
      summaryInput({
        findings: Array.from({ length: 16 }, (_, index) => makeFinding(index)),
        proposals: Array.from({ length: 7 }, (_, index) => makeProposal(index)),
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    expect(summary.findings).toHaveLength(15)
    expect(summary.proposals).toHaveLength(7)
    expect(summary.findings_truncated).toBe(true)
    expect(summary.proposals_truncated).toBeUndefined()
    expect(byteSize(summary)).toBeLessThan(BYTE_BUDGET)
  })

  it("caps pathological dispatches at 15/15/10/5 and stays under budget", () => {
    const summary = buildSummaryRetrieval(
      summaryInput({
        findings: Array.from({ length: 512 }, (_, index) => makeFinding(index)),
        advisories: Array.from({ length: 512 }, (_, index) => makeAdvisory(index)),
        proposals: Array.from({ length: 512 }, (_, index) => makeProposal(index)),
        warnings: Array.from({ length: 512 }, () =>
          makeWarning("orphaned_submissions_excluded"),
        ),
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    expect(summary.findings).toHaveLength(15)
    expect(summary.advisories).toHaveLength(15)
    expect(summary.proposals).toHaveLength(10)
    expect(summary.warnings).toHaveLength(5)
    expect(summary.findings_truncated).toBe(true)
    expect(summary.advisories_truncated).toBe(true)
    expect(summary.proposals_truncated).toBe(true)
    expect(summary.warnings_truncated).toBe(true)
    expect(byteSize(summary)).toBeLessThan(BYTE_BUDGET)
  })

  it("drops findings before advisories when multibyte text blows the budget", () => {
    const emoji = "\u{1F600}".repeat(512)
    const summary = buildSummaryRetrieval(
      summaryInput({
        findings: Array.from({ length: 15 }, (_, index) =>
          makeFinding(index, {
            description: emoji,
            category: emoji,
            file: emoji,
          }),
        ),
        advisories: Array.from({ length: 15 }, (_, index) =>
          makeAdvisory(index, { description: emoji }),
        ),
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    expect(byteSize(summary)).toBeLessThan(BYTE_BUDGET)
    const kept = summary.findings as Array<Record<string, unknown>>
    expect(kept.length).toBeLessThan(15)
    expect(kept.length).toBeGreaterThan(0)
    // Trailing entries are dropped: the survivors are the leading findings.
    expect(kept.map((finding) => finding.line)).toEqual(
      Array.from({ length: kept.length }, (_, index) => index + 1),
    )
    expect(summary.findings_truncated).toBe(true)
    expect(summary.advisories).toHaveLength(15)
    expect(summary.advisories_truncated).toBeUndefined()
  })

  it("pins the limits to the RD-FR-049 literals", () => {
    expect(SUMMARY_LIMITS).toEqual({
      findings: 15,
      advisories: 15,
      proposals: 10,
      warnings: 5,
      errors: 5,
      severityChars: 16,
      categoryChars: 128,
      descriptionChars: 512,
      fileChars: 256,
      verdictReasonChars: 512,
      warningKindChars: 64,
      warningDetailChars: 256,
      maxBytes: BYTE_BUDGET,
    })
  })

  /**
   * Every array at its cap and every bounded field at (or cut to) its
   * maximum. Warning kinds are always one of the five documented names, so
   * the 64-character kind cap is unreachable; the composed detail is pinned
   * at 256 through an oversized agent name. Proposals carry no per-field
   * cap (RD-FR-058); they use the schema-maximum agent and tag (128) with a
   * 512-character body, kept verbatim.
   */
  function worstCaseInput(text: (length: number) => string) {
    return summaryInput({
      verdict_reason: text(600),
      findings: Array.from({ length: 15 }, (_, index) =>
        makeFinding(index, {
          severity: text(20),
          category: text(200),
          description: text(600),
          file: text(300),
          line: index + 1,
        }),
      ),
      advisories: Array.from({ length: 15 }, (_, index) =>
        makeAdvisory(index, { severity: text(20), description: text(600) }),
      ),
      proposals: Array.from({ length: 10 }, (_, index) =>
        makeProposal(index, { agent: text(128), information: text(512), tag: text(128) }),
      ),
      warnings: Array.from({ length: 5 }, () => ({
        ...makeWarning("orphaned_submissions_excluded"),
        agent: text(300),
      })),
    }) as Parameters<typeof buildSummaryRetrieval>[0]
  }

  it("keeps every-field-at-maximum entries under budget without dropping", () => {
    const summary = buildSummaryRetrieval(worstCaseInput((length) => "x".repeat(length)))
    expect(byteSize(summary)).toBeLessThan(BYTE_BUDGET)
    expect(summary.findings).toHaveLength(15)
    expect(summary.advisories).toHaveLength(15)
    expect(summary.proposals).toHaveLength(10)
    expect(summary.warnings).toHaveLength(5)
    for (const flag of ["findings_truncated", "advisories_truncated", "proposals_truncated", "warnings_truncated"]) {
      expect(summary[flag]).toBeUndefined()
    }
    expect((summary.verdict_reason as string).length).toBe(512)
    for (const finding of summary.findings as Array<Record<string, string>>) {
      expect(finding.severity.length).toBe(16)
      expect(finding.category.length).toBe(128)
      expect(finding.description.length).toBe(512)
      expect(finding.file.length).toBe(256)
      expect(finding.truncated).toBe(true)
    }
    for (const advisory of summary.advisories as Array<Record<string, string>>) {
      expect(advisory.description.length).toBe(512)
    }
    for (const proposal of summary.proposals as Array<Record<string, string>>) {
      expect(proposal.agent.length).toBe(128)
      expect(proposal.information.length).toBe(512)
      expect(proposal.tag.length).toBe(128)
      expect(proposal).not.toHaveProperty("truncated")
    }
    for (const warning of summary.warnings as Array<Record<string, string>>) {
      expect(warning.detail.length).toBe(256)
      expect(warning.truncated).toBe(true)
    }
  })

  it("holds the budget by construction for escape-heavy multibyte content", () => {
    // Quotes and backslashes serialize to two bytes, control characters to
    // six (\u0001), CJK to three, and astral emoji to four.
    const unit = '"\\\u0001\n中文\u{1F600}'
    const text = (length: number) => unit.repeat(Math.ceil(length / unit.length) + 1)
    const summary = buildSummaryRetrieval(worstCaseInput(text))

    expect(byteSize(summary)).toBeLessThan(BYTE_BUDGET)
    // The drop loop engaged, consuming findings first.
    expect(summary.findings_truncated).toBe(true)
    const findings = summary.findings as Array<Record<string, unknown>>
    const advisories = summary.advisories as Array<Record<string, unknown>>
    if (summary.advisories_truncated === true) expect(findings).toHaveLength(0)
    if (summary.proposals_truncated === true) expect(advisories).toHaveLength(0)
    expect(findings.map((finding) => finding.line)).toEqual(
      Array.from({ length: findings.length }, (_, index) => index + 1),
    )
    // Truncation never splits a surrogate pair.
    expect(JSON.stringify(summary)).not.toMatch(/\\ud83d(?!\\ude00)/i)
    for (const finding of findings) {
      expect(finding.description as string).not.toMatch(/[\ud800-\udbff]$/)
    }
  })

  it("never splits a surrogate pair at a truncation boundary", () => {
    // "a" plus emoji puts a high surrogate at index 511, the last kept unit.
    const summary = buildSummaryRetrieval(
      summaryInput({
        findings: [makeFinding(0, { description: "a" + "\u{1F600}".repeat(600) })],
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    const finding = (summary.findings as Array<Record<string, unknown>>)[0]
    expect(finding.description).toBe("a" + "\u{1F600}".repeat(255))
    expect(finding.truncated).toBe(true)
  })

  it("composes warning detail per kind and discards unknown kinds", () => {
    const summary = buildSummaryRetrieval(
      summaryInput({
        warnings: [
          makeWarning("orphaned_submissions_excluded"),
          makeWarning("invalid_submissions_excluded"),
          makeWarning("run_findings_overflow"),
          makeWarning("consolidated_findings_overflow"),
          makeWarning("empty_coverage_attestation"),
        ],
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    expect(summary.warnings).toEqual([
      {
        kind: "orphaned_submissions_excluded",
        detail: "agent=divisor-guard run=run-guard findings=2 proposals=1",
      },
      {
        kind: "invalid_submissions_excluded",
        detail: "agent=divisor-guard run=run-guard invalid=3",
      },
      {
        kind: "run_findings_overflow",
        detail: "agent=divisor-guard kept=512 dropped=4",
      },
      {
        kind: "consolidated_findings_overflow",
        detail: "kept=512 dropped=5",
      },
      {
        kind: "empty_coverage_attestation",
        detail: "agent=divisor-guard",
      },
    ])

    const withUnknown = buildSummaryRetrieval(
      summaryInput({
        warnings: [{ type: "unknown_warning_kind" }],
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    expect(withUnknown).not.toHaveProperty("warnings")
  })

  it("projects child_empty_response detail with and without reported provenance (RD-FR-061)", () => {
    // Full provenance: finish reason and token counts ride along — a
    // reported zero output count is the issue #1052 signature and must
    // survive as a count, not read as absent.
    const full = buildSummaryRetrieval(
      summaryInput({
        warnings: [makeWarning("child_empty_response")],
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    expect(full.warnings).toEqual([
      {
        kind: "child_empty_response",
        detail:
          "agent=divisor-guard run=run-guard finish=length tokens.input=4096 tokens.output=0",
      },
    ])

    // Unreported provenance reads as absent segments, never as zero counts.
    const unreported = buildSummaryRetrieval(
      summaryInput({
        warnings: [
          {
            type: "child_empty_response",
            agent: "divisor-guard",
            run_id: "run-guard",
            finish_reason: null,
            tokens_input: null,
            tokens_output: null,
          },
        ],
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    expect(unreported.warnings).toEqual([
      {
        kind: "child_empty_response",
        detail: "agent=divisor-guard run=run-guard",
      },
    ])
  })

  it("projects slim entries and marks field truncation per entry", () => {
    const summary = buildSummaryRetrieval(
      summaryInput({
        findings: [
          makeFinding(1, {
            category: "c".repeat(200),
            description: "d".repeat(600),
            file: null,
            line: null,
          }),
        ],
        advisories: [makeAdvisory(1, { description: "a".repeat(600) })],
        proposals: [makeProposal(1, { information: "i".repeat(600), category: "invalid" })],
      }) as Parameters<typeof buildSummaryRetrieval>[0],
    )
    const finding = (summary.findings as Array<Record<string, unknown>>)[0]
    expect(finding).toMatchObject({ file: null, line: null, truncated: true })
    expect((finding.description as string).length).toBe(512)
    expect(finding).not.toHaveProperty("root_cause")
    expect(finding).not.toHaveProperty("run_ids")
    const advisory = (summary.advisories as Array<Record<string, unknown>>)[0]
    expect(advisory.truncated).toBe(true)
    const proposal = (summary.proposals as Array<Record<string, unknown>>)[0]
    // Proposals are never shortened (RD-FR-058, issue #969).
    expect(proposal.information).toBe("i".repeat(600))
    expect(proposal).not.toHaveProperty("truncated")
    expect(proposal).not.toHaveProperty("category")
  })

  describe("complete lesson proposals (RD-FR-058, issue #969)", () => {
    // A lesson whose conclusion sits past the old 512-character cap.
    const LONG_LESSON =
      "When a review dispatch harvests lesson proposals, the parent stores them verbatim. ".repeat(30) +
      "Conclusion: never shorten a proposal body in any retrieval mode."

    it("returns a proposal past 512 characters verbatim and unflagged", () => {
      expect(LONG_LESSON.length).toBeGreaterThan(2048)
      const summary = buildSummaryRetrieval(
        summaryInput({ proposals: [makeProposal(1, { information: LONG_LESSON })] }) as Parameters<
          typeof buildSummaryRetrieval
        >[0],
      )
      const proposals = summary.proposals as Array<Record<string, unknown>>
      expect(proposals[0]).toEqual({
        agent: "divisor-guard",
        information: LONG_LESSON,
        tag: "review-dispatch",
        category: "pattern",
      })
      expect(summary.proposals_truncated).toBeUndefined()
    })

    it("drops over-budget proposals whole and flags the drop", () => {
      // Ten schema-maximum (16,384-character) proposals cannot fit 40 KiB.
      const maximal = (index: number) => `${index}`.padEnd(16_384, "x")
      const summary = buildSummaryRetrieval(
        summaryInput({
          findings: [makeFinding(1)],
          proposals: Array.from({ length: 10 }, (_, index) =>
            makeProposal(index, { information: maximal(index) }),
          ),
        }) as Parameters<typeof buildSummaryRetrieval>[0],
      )
      expect(byteSize(summary)).toBeLessThan(BYTE_BUDGET)
      expect(summary.proposals_truncated).toBe(true)
      // Findings drop before proposals.
      expect(summary.findings).toEqual([])
      // The displacement damage design.md names as the mitigation.
      expect(summary.findings_truncated).toBe(true)
      const kept = summary.proposals as Array<Record<string, unknown>>
      expect(kept.length).toBeGreaterThan(0)
      expect(kept.length).toBeLessThan(10)
      kept.forEach((proposal, index) => {
        expect(proposal.information).toBe(maximal(index))
        expect(proposal).not.toHaveProperty("truncated")
      })
    })

    it("keeps the compact finalization proposals verbatim and under budget", () => {
      const result = {
        status: "success",
        operation_verdict: "APPROVE",
        correlation_id: "11111111-1111-4111-8111-111111111111",
        artifact_path: "/tmp/artifact.json",
        failure_record: null,
        review_verdict: { council_decision: "APPROVED" },
        calculated_assessment: { authoritative: true },
        errors: [],
      } as unknown as Parameters<typeof buildCompactFinalization>[0]
      const single = buildCompactFinalization(result, RUN_COUNTS, [makeProposal(1, { information: LONG_LESSON })], [])
      expect((single.proposals as Array<Record<string, unknown>>)[0].information).toBe(LONG_LESSON)
      expect(single.proposals_truncated).toBeUndefined()

      const maximal = (index: number) => `${index}`.padEnd(16_384, "y")
      const crowded = buildCompactFinalization(
        result,
        RUN_COUNTS,
        Array.from({ length: 10 }, (_, index) => makeProposal(index, { information: maximal(index) })),
        [],
      )
      expect(byteSize(crowded)).toBeLessThan(BYTE_BUDGET)
      expect(crowded.proposals_truncated).toBe(true)
      const kept = crowded.proposals as Array<Record<string, unknown>>
      expect(kept.length).toBeGreaterThan(0)
      kept.forEach((proposal, index) => expect(proposal.information).toBe(maximal(index)))
      expect(Object.keys(crowded).slice(-2)).toEqual(["calculated_assessment", "errors"])
    })
  })
})

describe("full retrieval", () => {
  it("orders observability before the large arrays and keeps runs first", () => {
    const full = buildFullRetrieval(
      fullPayload() as Parameters<typeof buildFullRetrieval>[0],
      [makeProposal(1)],
      [makeWarning("empty_coverage_attestation")],
    )
    const keys = Object.keys(full)
    expect(keys.indexOf("correlation_id")).toBeLessThan(keys.indexOf("verdict"))
    expect(keys.indexOf("proposals")).toBeLessThan(keys.indexOf("verdict"))
    expect(keys.indexOf("warnings")).toBeLessThan(keys.indexOf("verdict"))
    expect(keys.indexOf("runs")).toBeLessThan(keys.indexOf("plan"))
    expect(keys.indexOf("runs")).toBeLessThan(keys.indexOf("findings"))
    expect(keys.indexOf("runs")).toBeLessThan(keys.indexOf("advisories"))
  })

  it("omits empty proposals and warnings", () => {
    const full = buildFullRetrieval(
      fullPayload() as Parameters<typeof buildFullRetrieval>[0],
      [],
      [],
    )
    expect(full).not.toHaveProperty("proposals")
    expect(full).not.toHaveProperty("warnings")
    expect(Object.keys(full)[0]).toBe("command")
  })
})

describe("compact finalization", () => {
  const result = {
    status: "success",
    operation_verdict: "APPROVE",
    correlation_id: "11111111-1111-4111-8111-111111111111",
    artifact_path: "/tmp/artifact.json",
    failure_record: null,
    review_dispatch: { payload: { runs: [] } },
    review_verdict: { council_decision: "APPROVED" },
    calculated_assessment: { authoritative: true, payload: { runs: [] } },
    errors: [],
  } as unknown as Parameters<typeof buildCompactFinalization>[0]

  it("emits the compact ordered shape with a scalar review_verdict", () => {
    const compact = buildCompactFinalization(
      result,
      RUN_COUNTS,
      [makeProposal(1)],
      [makeWarning("empty_coverage_attestation")],
    )
    expect(Object.keys(compact)).toEqual([
      "status",
      "operation_verdict",
      "correlation_id",
      "artifact_path",
      "review_verdict",
      "run_counts",
      "proposals",
      "warnings",
      "calculated_assessment",
      "errors",
    ])
    expect(compact.review_verdict).toBe("APPROVED")
    expect(compact.calculated_assessment).toEqual({ authoritative: true })
    expect(compact).not.toHaveProperty("review_dispatch")
    expect(compact).not.toHaveProperty("payload")
    // RD-FR-066: the failure-record locator is omitted when null.
    expect(compact).not.toHaveProperty("failure_record")
  })

  it("places failure_record immediately after artifact_path (RD-FR-066 D3)", () => {
    // A persisted failure record surfaces as a locator immediately after
    // artifact_path — early keys survive host tail-truncation — while errors
    // stays the final key. The record writes that produce non-null values
    // land with the phase-2 persistence tasks; this pins the placement the
    // compact builder must hold for them.
    const failureRecord = {
      correlation_id: "22222222-2222-4222-8222-222222222222",
      path: "/project/.uf/artifacts/dispatch/22222222-2222-4222-8222-222222222222.failure.json",
    }
    const compact = buildCompactFinalization(
      {
        ...result,
        status: "failed",
        artifact_path: null,
        failure_record: failureRecord,
      } as unknown as Parameters<typeof buildCompactFinalization>[0],
      RUN_COUNTS,
      [],
      [],
    )
    expect(compact.failure_record).toEqual(failureRecord)
    const keys = Object.keys(compact)
    expect(keys.indexOf("failure_record")).toBe(keys.indexOf("artifact_path") + 1)
    expect(keys[keys.length - 1]).toBe("errors")
  })

  it("caps observability arrays and reports truncation flags", () => {
    const compact = buildCompactFinalization(
      {
        ...result,
        errors: Array.from({ length: 7 }, () => "e".repeat(300)),
      } as unknown as Parameters<typeof buildCompactFinalization>[0],
      RUN_COUNTS,
      Array.from({ length: 12 }, (_, index) => makeProposal(index)),
      Array.from({ length: 7 }, () => makeWarning("orphaned_submissions_excluded")),
    )
    expect(compact.proposals).toHaveLength(10)
    expect(compact.warnings).toHaveLength(5)
    expect(compact.errors).toHaveLength(5)
    // RD-FR-066: the count cap drops whole entries; every survivor carries
    // its complete 300-character string, never a shortened one.
    expect(compact.errors).toEqual(Array.from({ length: 5 }, () => "e".repeat(300)))
    expect(compact.proposals_truncated).toBe(true)
    expect(compact.warnings_truncated).toBe(true)
    expect(compact.errors_truncated).toBe(true)
    const keys = Object.keys(compact)
    expect(keys.indexOf("proposals_truncated")).toBeLessThan(
      keys.indexOf("calculated_assessment"),
    )
    expect(keys.indexOf("warnings_truncated")).toBeLessThan(
      keys.indexOf("calculated_assessment"),
    )
    // RD-FR-066 amends RD-FR-051's order: errors_truncated rides with the
    // other observability flags (it previously followed errors), so errors
    // stays the final key.
    expect(keys.indexOf("errors_truncated")).toBeLessThan(
      keys.indexOf("calculated_assessment"),
    )
    expect(keys[keys.length - 1]).toBe("errors")
  })

  it("carries a long error verbatim without raising errors_truncated (RD-FR-066)", () => {
    // Longer than the removed 256-character per-string cap: shortening no
    // longer exists, so a long entry that is not dropped must neither be cut
    // nor flagged. Issue #1039 regression — the old projection shortened the
    // string and set errors_truncated for the shortening alone.
    const long = "e".repeat(300)
    const compact = buildCompactFinalization(
      {
        ...result,
        errors: [long],
      } as unknown as Parameters<typeof buildCompactFinalization>[0],
      RUN_COUNTS,
      [],
      [],
    )
    expect(compact.errors).toEqual([long])
    expect(compact.errors_truncated).toBeUndefined()
  })

  it("carries a composed validation error whole with no truncated key (RD-FR-066)", () => {
    // The structural-failure path's real error: formatValidationError joins
    // every Zod issue as `path: message`, sorted, with "; " — the spec's
    // three-or-more-issue threshold pushes the composition past the removed
    // 256-character cap. Issue #1039 regression: the old projection cut
    // this string mid-composition, silently discarding the tail issues on
    // exactly the path where no artifact existed to recover them from. The
    // composition uses the production function over a representative
    // schema, so the fixture is the string shape the finalizer produces.
    const outcome = z
      .object({
        command: z.enum(["review-council", "triage-issue", "address-feedback"]),
        mode: z.enum(["code", "specs", "triage", "feedback"]),
        correlation_id: z.string().uuid(),
        full: z.boolean(),
        run_counts: z.object({
          total: z.number().int(),
          success: z.number().int(),
          failed: z.number().int(),
        }),
      })
      .safeParse({
        command: "review-council",
        mode: "definitely-not-a-mode",
        correlation_id: "not-a-uuid-at-all",
        full: "yes",
        run_counts: { total: "four", success: -1.5, failed: null },
      })
    if (outcome.success) {
      throw new Error("test fixture bug: the invalid input must fail validation")
    }
    expect(outcome.error.issues.length).toBeGreaterThanOrEqual(3)
    // Composed exactly as structuralFailure composes it: the wrapper plus
    // the joined issues, with nothing between them and the projection.
    const composed = `validate review dispatch input: ${formatValidationError(outcome.error)}`
    expect(composed.length).toBeGreaterThan(256)

    const compact = buildCompactFinalization(
      {
        ...result,
        status: "failed",
        operation_verdict: "INCONCLUSIVE",
        artifact_path: null,
        review_verdict: null,
        calculated_assessment: null,
        errors: [composed],
      } as unknown as Parameters<typeof buildCompactFinalization>[0],
      RUN_COUNTS,
      [],
      [],
    )
    // The complete composition survives — every issue, verbatim. The entry
    // is the bare string itself, so no truncated key can exist on it, and a
    // kept-but-long entry raises no errors_truncated flag.
    expect(compact.errors).toEqual([composed])
    expect(compact.errors_truncated).toBeUndefined()
  })

  it("drops trailing errors whole once proposals are exhausted (RD-FR-066 D2)", () => {
    // Five whole errors blow the 40 KiB budget: uncapped since RD-FR-066
    // removed the 256-character per-string cap, the budget loop must drop
    // trailing entries whole — never shorten — and flag the drop. Regression
    // for issue #1039: before the loop learned to drop errors, an
    // error-heavy failure result sailed past the budget intact.
    const errors = Array.from({ length: 5 }, (_, index) => `error ${index}: `.padEnd(12_000, "e"))
    const compact = buildCompactFinalization(
      {
        ...result,
        status: "failed",
        operation_verdict: "INCONCLUSIVE",
        artifact_path: null,
        review_verdict: null,
        calculated_assessment: null,
        errors,
      } as unknown as Parameters<typeof buildCompactFinalization>[0],
      RUN_COUNTS,
      [],
      [],
    )
    expect(byteSize(compact)).toBeLessThan(BYTE_BUDGET)
    expect(compact.errors_truncated).toBe(true)
    const kept = compact.errors as string[]
    expect(kept.length).toBeGreaterThan(0)
    expect(kept.length).toBeLessThan(errors.length)
    // Trailing entries drop; every survivor is complete and verbatim.
    expect(compact.errors).toEqual(errors.slice(0, kept.length))
  })

  it("drops proposals before errors when both compete for the budget (RD-FR-066 D2)", () => {
    // Both arrays blow the budget together: proposals — recoverable from
    // the success artifact — drop first and completely, then trailing
    // errors drop whole. Errors drop last because they are the diagnostic
    // payload of exactly the failure paths where nothing else is persisted.
    const maximal = (index: number) => `${index}`.padEnd(16_384, "y")
    const errors = Array.from({ length: 5 }, (_, index) => `error ${index}: `.padEnd(12_000, "e"))
    const compact = buildCompactFinalization(
      {
        ...result,
        status: "failed",
        operation_verdict: "INCONCLUSIVE",
        artifact_path: null,
        review_verdict: null,
        calculated_assessment: null,
        errors,
      } as unknown as Parameters<typeof buildCompactFinalization>[0],
      RUN_COUNTS,
      Array.from({ length: 3 }, (_, index) => makeProposal(index, { information: maximal(index) })),
      [],
    )
    expect(byteSize(compact)).toBeLessThan(BYTE_BUDGET)
    // Every proposal dropped: the key is omitted and the drop is flagged.
    expect(compact).not.toHaveProperty("proposals")
    expect(compact.proposals_truncated).toBe(true)
    expect(compact.errors_truncated).toBe(true)
    const kept = compact.errors as string[]
    expect(kept.length).toBeGreaterThan(0)
    expect(kept.length).toBeLessThan(errors.length)
    expect(compact.errors).toEqual(errors.slice(0, kept.length))
  })

  it("stops once proposals and errors are both exhausted instead of spinning", () => {
    // run_counts is caller-supplied projection input; a pathological
    // oversized value leaves the loop nothing droppable, so it must break
    // and return the complete result unflagged rather than loop forever.
    // The real pipeline cannot build this shape (the payload schema bounds
    // run_counts to seven numeric fields) — the branch is defensive.
    const giantRunCounts = { filler: "r".repeat(50_000) }
    const compact = buildCompactFinalization(
      { ...result, errors: [] } as unknown as Parameters<typeof buildCompactFinalization>[0],
      giantRunCounts,
      [],
      [],
    )
    expect(byteSize(compact)).toBeGreaterThanOrEqual(BYTE_BUDGET)
    expect(compact.errors_truncated).toBeUndefined()
    expect(compact.proposals_truncated).toBeUndefined()
  })

  it("reports a null calculated assessment and preserves errors", () => {
    const compact = buildCompactFinalization(
      {
        ...result,
        calculated_assessment: null,
        errors: ["persistence failed"],
      } as unknown as Parameters<typeof buildCompactFinalization>[0],
      RUN_COUNTS,
      [],
      [],
    )
    expect(compact.calculated_assessment).toBeNull()
    expect(compact.errors).toEqual(["persistence failed"])
    expect(compact).not.toHaveProperty("proposals")
    expect(compact).not.toHaveProperty("warnings")
  })
})
