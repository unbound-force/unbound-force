import { describe, expect, it } from "vitest"

import { classifyGhFailure, type GhFailureSignals } from "../lib/gh-error.js"
import { classifyGhError } from "../lib/feedback-replies.js"

/**
 * One row of the shared decision table: the signals and the expected
 * classification. FRR-FR-001 (issue #1023 MEDIUM 3) requires every gh
 * consumer to derive its code and retryability from
 * `classifyGhFailure`; these rows are the table of record.
 */
interface Row {
  readonly name: string
  readonly signals: GhFailureSignals
  readonly code: string
  readonly retryable: boolean
}

const ROWS: readonly Row[] = [
  { name: "missing gh binary", signals: { spawnCode: "ENOENT" }, code: "gh_not_found", retryable: false },
  { name: "spawn timeout", signals: { spawnCode: "ETIMEDOUT" }, code: "timeout", retryable: true },
  { name: "killed by SIGTERM", signals: { signal: "SIGTERM" }, code: "timeout", retryable: true },
  { name: "stderr says timed out", signals: { stderr: "gh command timed out" }, code: "timeout", retryable: true },
  {
    name: "stderr says aborted due to timeout",
    signals: { stderr: "The operation was aborted due to timeout" },
    code: "timeout",
    retryable: true,
  },
  { name: "stderr says rate limit", signals: { stderr: "HTTP 403: API rate limit exceeded" }, code: "rate_limited", retryable: true },
  { name: "stderr says gh auth login", signals: { stderr: "To get started with GitHub CLI, please run:  gh auth login" }, code: "gh_auth_failed", retryable: false },
  { name: "stderr says not logged in", signals: { stderr: "not logged in" }, code: "gh_auth_failed", retryable: false },
  { name: "exit 4 with no matching text", signals: { stderr: "", exitCode: 4 }, code: "gh_auth_failed", retryable: false },
  { name: "exit 1 with no matching text", signals: { stderr: "HTTP 422: Validation Failed", exitCode: 1 }, code: "api_error", retryable: false },
  { name: "no signals at all", signals: {}, code: "api_error", retryable: false },
]

describe("classifyGhFailure", () => {
  it.each(ROWS)("classifies $name", ({ signals, code, retryable }) => {
    const classified = classifyGhFailure(signals)
    expect(classified.code).toBe(code)
    expect(classified.retryable).toBe(retryable)
  })

  it("lets stderr text outrank exit code 4: a rate limit gh also reports as exit 4 stays rate_limited", () => {
    const classified = classifyGhFailure({ stderr: "HTTP 429: API rate limit exceeded", exitCode: 4 })
    expect(classified).toEqual({ code: "rate_limited", retryable: true })
  })

  it("lets authentication text outrank exit code 4 (same classification, text is the stronger signal)", () => {
    const classified = classifyGhFailure({ stderr: "gh auth login required", exitCode: 4 })
    expect(classified).toEqual({ code: "gh_auth_failed", retryable: false })
  })

  it("classifies on the thrown message when stderr is absent", () => {
    const classified = classifyGhFailure({ message: "To get started with GitHub CLI, please run:  gh auth login" })
    expect(classified).toEqual({ code: "gh_auth_failed", retryable: false })
  })

  it("never throws on hostile inputs", () => {
    expect(classifyGhFailure({ stderr: "\u0000\u0001", exitCode: Number.NaN })).toEqual({
      code: "api_error",
      retryable: false,
    })
  })
})

/**
 * Adapter parity (FRR-FR-001): the feedback-replies adapter must map
 * every execFileSync-shaped failure onto the shared table without
 * drift — same thrown record, same code and retryability as
 * `classifyGhFailure` sees from the extracted signals. The
 * review-dispatch adapter (`ghFailureError`) is pinned at its public
 * surface in fetch-pr-checks.test.ts.
 */
describe("classifyGhError adapter parity", () => {
  it.each(ROWS)("maps $name through classifyGhError onto the shared table", ({ signals, code, retryable }) => {
    const thrown = Object.assign(new Error("Command failed: gh"), {
      stderr: signals.stderr ?? "",
      status: signals.exitCode ?? 1,
      ...(signals.spawnCode === undefined ? {} : { code: signals.spawnCode }),
      ...(signals.signal === undefined ? {} : { signal: signals.signal }),
    })
    const classified = classifyGhError(thrown)
    expect(classified.code).toBe(code)
    expect(classified.retryable).toBe(retryable)
  })
})
