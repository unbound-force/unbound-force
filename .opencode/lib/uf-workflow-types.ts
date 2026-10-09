/**
 * Result envelope types for uf-workflow plugin tools.
 *
 * Every tool response is wrapped in a discriminated union
 * so callers can branch on `status` without guessing.
 *
 * Error messages SHOULD NOT include file paths or stack
 * traces — keep them user-facing and sanitized.
 */

/** Successful tool result with optional typed payload. */
export interface ToolSuccess<T = void> {
  readonly status: "ok"
  readonly data?: T
}

/** Failed tool result with structured error metadata. */
export interface ToolFailure {
  readonly status: "error"
  readonly message: string
  readonly retryable: boolean
  readonly code?: string
}

/** Discriminated union returned by every tool handler. */
export type ToolResult<T = void> = ToolSuccess<T> | ToolFailure

/** Build a success envelope, optionally carrying data. */
export function success<T = void>(data?: T): ToolSuccess<T> {
  if (data === undefined) {
    return { status: "ok" } as ToolSuccess<T>
  }
  return { status: "ok", data }
}

/** Build a failure envelope. `retryable` defaults to false. */
export function failure(
  message: string,
  options?: { retryable?: boolean; code?: string },
): ToolFailure {
  const result: ToolFailure = {
    status: "error",
    message,
    retryable: options?.retryable ?? false,
    ...(options?.code !== undefined && { code: options.code }),
  }
  return result
}
