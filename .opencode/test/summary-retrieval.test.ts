import { describe, expect, it } from "vitest"

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
  })

  it("caps observability arrays and reports truncation flags", () => {
    const compact = buildCompactFinalization(
      result,
      RUN_COUNTS,
      Array.from({ length: 12 }, (_, index) => makeProposal(index)),
      Array.from({ length: 7 }, () => makeWarning("orphaned_submissions_excluded")),
    )
    expect(compact.proposals).toHaveLength(10)
    expect(compact.warnings).toHaveLength(5)
    expect(compact.proposals_truncated).toBe(true)
    expect(compact.warnings_truncated).toBe(true)
    const keys = Object.keys(compact)
    expect(keys.indexOf("proposals_truncated")).toBeLessThan(
      keys.indexOf("calculated_assessment"),
    )
    expect(keys.indexOf("warnings_truncated")).toBeLessThan(
      keys.indexOf("calculated_assessment"),
    )
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
