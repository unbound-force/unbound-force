import { mkdtemp, rm } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import {
  createFinalizationDependencies,
  finalizeReviewDispatch,
  type FinalizationDependencies,
  type GenericVerdict,
  type ReviewDispatchPayload,
} from "../plugins/review-dispatch/index.js"
import samplePayload from "../../schemas/review-dispatch/samples/sample-review-dispatch.json"
import consolidationFixture from "./fixtures/review-dispatch/consolidation-cases.json"

type FixtureOperation =
  | { readonly op: "set"; readonly path: string; readonly value: unknown }
  | { readonly op: "copy"; readonly from: string; readonly path: string }

interface ExpectedModelProvenance {
  readonly run_id: string
  readonly requested_model: string | null
  readonly resolved_parent_model: string | null
  readonly reported_model: string | null
}

interface ExpectedContributorProvenance {
  readonly run_id: string
  readonly agent: string
  readonly source: string
  readonly requested_model: string | null
  readonly variant: string | null
  readonly sequence: number
}

interface FixtureExpectation {
  readonly finalize: "success" | "semantic_failure" | "structural_failure" | "persistence_failure"
  readonly operation_verdict: GenericVerdict
  readonly native_value?: string
  readonly canonical_decision?: string
  readonly persona_summary?: string
  readonly persona_count?: number
  readonly error_contains?: readonly string[]
  readonly non_success_has_no_votes_or_findings?: boolean
  readonly run_counts?: Readonly<Record<string, number>>
  readonly finding_run_ids?: readonly (readonly string[])[]
  readonly canonical_finding_count?: number
  readonly contributor_provenance?: readonly ExpectedContributorProvenance[]
  readonly retained_model?: ExpectedModelProvenance
}

interface ConsolidationCase {
  readonly id: string
  readonly templates: readonly string[]
  readonly operations: readonly FixtureOperation[]
  readonly schema_valid: boolean
  readonly persistence_failure?: boolean
  readonly expected: FixtureExpectation
}

interface ConsolidationFixtureSuite {
  readonly base_fixture: string
  readonly templates: Readonly<Record<string, readonly FixtureOperation[]>>
  readonly cases: readonly ConsolidationCase[]
}

const fixtureSuite = consolidationFixture as unknown as ConsolidationFixtureSuite
const scratchDirectories: string[] = []

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}

function pointerSegments(pointer: string): string[] {
  if (!pointer.startsWith("/")) {
    throw new Error(`fixture pointer must start with /: ${pointer}`)
  }
  return pointer
    .slice(1)
    .split("/")
    .map((segment) => segment.replaceAll("~1", "/").replaceAll("~0", "~"))
}

function arrayIndex(segment: string, length: number, allowAppend: boolean): number {
  if (!/^(0|[1-9][0-9]*)$/.test(segment)) {
    throw new Error(`fixture array index is invalid: ${segment}`)
  }
  const index = Number(segment)
  const maximum = allowAppend ? length : length - 1
  if (!Number.isSafeInteger(index) || index < 0 || index > maximum) {
    throw new Error(`fixture array index ${index} is outside 0..${maximum}`)
  }
  return index
}

function readPointer(root: unknown, pointer: string): unknown {
  let current = root
  for (const segment of pointerSegments(pointer)) {
    if (Array.isArray(current)) {
      current = current[arrayIndex(segment, current.length, false)]
    } else if (isRecord(current) && segment in current) {
      current = current[segment]
    } else {
      throw new Error(`fixture pointer does not exist: ${pointer}`)
    }
  }
  return current
}

function setPointer(root: unknown, pointer: string, value: unknown): void {
  const segments = pointerSegments(pointer)
  const property = segments.pop()
  if (property === undefined) {
    throw new Error("fixture pointer must identify a property")
  }
  let parent = root
  for (const segment of segments) {
    if (Array.isArray(parent)) {
      parent = parent[arrayIndex(segment, parent.length, false)]
    } else if (isRecord(parent) && segment in parent) {
      parent = parent[segment]
    } else {
      throw new Error(`fixture pointer parent does not exist: ${pointer}`)
    }
  }

  const clonedValue = structuredClone(value)
  if (Array.isArray(parent)) {
    const index = arrayIndex(property, parent.length, true)
    if (index === parent.length) {
      parent.push(clonedValue)
    } else {
      parent[index] = clonedValue
    }
    return
  }
  if (!isRecord(parent)) {
    throw new Error(`fixture pointer parent is not a container: ${pointer}`)
  }
  parent[property] = clonedValue
}

function applyOperations(root: unknown, operations: readonly FixtureOperation[]): void {
  for (const operation of operations) {
    const value = operation.op === "copy" ? readPointer(root, operation.from) : operation.value
    setPointer(root, operation.path, value)
  }
}

function materializeFixture(testCase: ConsolidationCase): ReviewDispatchPayload {
  const payload: unknown = structuredClone(samplePayload)
  for (const templateName of testCase.templates) {
    const template = fixtureSuite.templates[templateName]
    if (template === undefined) {
      throw new Error(`unknown consolidation fixture template: ${templateName}`)
    }
    applyOperations(payload, template)
  }
  applyOperations(payload, testCase.operations)
  return payload as ReviewDispatchPayload
}

async function scratchDirectory(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-consolidation-fixture-"))
  scratchDirectories.push(directory)
  return directory
}

function finalizationDependencies(projectRoot: string, persistenceFailure: boolean): FinalizationDependencies {
  const dependencies = createFinalizationDependencies(projectRoot)
  return {
    ...dependencies,
    now: (): Date => new Date("2026-10-01T12:30:00Z"),
    uuid: (): string => "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
    link: persistenceFailure
      ? async (): Promise<void> => {
          throw new Error("fixture persistence failure")
        }
      : dependencies.link,
  }
}

function findRun(payload: ReviewDispatchPayload, runID: string): ReviewDispatchPayload["runs"][number] {
  const run = payload.runs.find((candidate) => candidate.run_id === runID)
  if (run === undefined) {
    throw new Error(`expected fixture run ${runID}`)
  }
  return run
}

describe("review dispatch consolidation fixtures", () => {
  it("uses the registered review-dispatch schema sample as its reusable base", () => {
    expect(fixtureSuite.base_fixture).toBe("schemas/review-dispatch/samples/sample-review-dispatch.json")
  })

  it.each(fixtureSuite.cases)("finalizes $id at the semantic boundary", async (testCase) => {
    const directory = await scratchDirectory()
    const payload = materializeFixture(testCase)
    const result = await finalizeReviewDispatch(
      {
        payload,
        provenance: {
          branch: "feature/review-fanout",
          commit: "2222222222222222222222222222222222222222",
          workflow_id: `fixture-${testCase.id}`,
        },
      },
      finalizationDependencies(directory, testCase.persistence_failure === true),
    )

    expect(result.operation_verdict).toBe(testCase.expected.operation_verdict)
    expect(result.status).toBe(testCase.expected.finalize === "success" ? "success" : "failed")

    if (testCase.expected.finalize === "structural_failure") {
      expect(result.correlation_id).toBeNull()
      expect(result.calculated_assessment).toBeNull()
    } else if (testCase.expected.finalize === "semantic_failure") {
      expect(result.correlation_id).toBe(payload.correlation_id)
      expect(result.calculated_assessment).toBeNull()
      expect(result.artifact_path).toBeNull()
    } else if (testCase.expected.finalize === "persistence_failure") {
      expect(result.calculated_assessment?.authoritative).toBe(false)
      expect(result.artifact_path).toBeNull()
    } else {
      expect(result.calculated_assessment?.authoritative).toBe(true)
      expect(result.artifact_path).not.toBeNull()
    }

    for (const expectedError of testCase.expected.error_contains ?? []) {
      expect(result.errors.join(" ")).toContain(expectedError)
    }

    const projectedPayload = result.review_dispatch?.payload ?? result.calculated_assessment?.payload
    const projectedVerdict = result.review_verdict ?? result.calculated_assessment?.review_verdict
    if (testCase.expected.native_value !== undefined) {
      expect(projectedPayload?.workflow_result.value).toBe(testCase.expected.native_value)
    }
    if (testCase.expected.canonical_decision !== undefined) {
      expect(projectedVerdict?.schema_version).toBe("2.0.0")
      expect(projectedVerdict?.council_decision).toBe(testCase.expected.canonical_decision)
    }
    if (testCase.expected.persona_summary !== undefined) {
      expect(projectedVerdict?.persona_verdicts[0]?.summary).toBe(testCase.expected.persona_summary)
    }
    if (testCase.expected.persona_count !== undefined) {
      expect(projectedVerdict?.persona_verdicts).toHaveLength(testCase.expected.persona_count)
    }
    if (testCase.expected.run_counts !== undefined) {
      expect(projectedPayload?.run_counts).toEqual(testCase.expected.run_counts)
    }
    if (testCase.expected.finding_run_ids !== undefined) {
      expect(projectedPayload?.findings.map((finding) => finding.run_ids)).toEqual(testCase.expected.finding_run_ids)
    }
    if (testCase.expected.canonical_finding_count !== undefined) {
      expect(projectedVerdict?.unresolved_findings).toHaveLength(testCase.expected.canonical_finding_count)
    }
    if (testCase.expected.non_success_has_no_votes_or_findings === true) {
      for (const run of payload.runs.filter((candidate) => candidate.status !== "success")) {
        expect(run.workflow_verdict).toBeNull()
        expect(run.findings).toEqual([])
      }
    }
    for (const provenance of testCase.expected.contributor_provenance ?? []) {
      expect(findRun(projectedPayload ?? payload, provenance.run_id)).toMatchObject(provenance)
    }
    if (testCase.expected.retained_model !== undefined) {
      expect(findRun(payload, testCase.expected.retained_model.run_id)).toMatchObject(testCase.expected.retained_model)
    }
  })
})
