import { execSync } from "node:child_process"
import { tool, type PluginModule } from "@opencode-ai/plugin"
import { failure, success } from "../../lib/uf-workflow-types.js"

// Re-export envelope helpers for consumer convenience.
export { failure, success } from "../../lib/uf-workflow-types.js"
export type { ToolFailure, ToolResult, ToolSuccess } from "../../lib/uf-workflow-types.js"

/** Git execution function signature for testability. */
export type GitExec = (cmd: string, cwd: string) => string

/** Valid base-ref source labels. */
export type BaseRefSource = "upstream/main" | "origin/main" | "main"

/** Structured output from resolve_base_ref. */
export interface ResolveBaseRefData {
  ref: BaseRefSource
  sha: string
  source: BaseRefSource
}

/** Ordered candidate refs to try (no network access needed). */
const BASE_REF_CANDIDATES: readonly BaseRefSource[] = [
  "upstream/main",
  "origin/main",
  "main",
] as const

/** Default git executor using execSync. */
const defaultGitExec: GitExec = (cmd: string, cwd: string): string =>
  execSync(cmd, { cwd, encoding: "utf-8", stdio: ["pipe", "pipe", "pipe"] }).trim()

/**
 * Factory that creates the resolve_base_ref tool with injectable
 * git executor and working directory for testability.
 */
export function createResolveBaseRefTool(exec: GitExec, cwd: string) {
  return tool({
    description:
      "Resolve the best available base ref for local review diffs. " +
      "Tries upstream/main, origin/main, then main (no network access). " +
      "Returns the ref name, its resolved SHA, and a source label.",
    execute: async () => {
      for (const candidate of BASE_REF_CANDIDATES) {
        try {
          const sha = exec(`git rev-parse --verify ${candidate}`, cwd)
          if (sha) {
            const data: ResolveBaseRefData = {
              ref: candidate,
              sha,
              source: candidate,
            }
            return { output: JSON.stringify(success(data)) }
          }
        } catch {
          // Candidate not available, try next.
        }
      }
      return {
        output: JSON.stringify(
          failure(
            `No resolvable base ref found. Tried: ${BASE_REF_CANDIDATES.join(", ")}`,
            { retryable: false, code: "NO_BASE_REF" },
          ),
        ),
      }
    },
  })
}

/** UF Workflow plugin — provides workflow tools for review and release. */
const UfWorkflowPlugin = {
  id: "uf-workflow",
  server: async (input) => {
    const cwd = input.worktree || input.directory
    return {
      tool: {
        resolve_base_ref: createResolveBaseRefTool(defaultGitExec, cwd),
      },
    }
  },
} satisfies PluginModule

export default UfWorkflowPlugin
