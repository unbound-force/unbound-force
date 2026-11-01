import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import type { PluginInput } from "@opencode-ai/plugin"
import { afterEach, describe, expect, it } from "vitest"

import InvokeAgentPlugin from "../plugins/invoke-agent/index.js"
import ReviewDispatchPlugin from "../plugins/review-dispatch/index.js"
import UfFinalePlugin from "../plugins/uf-finale/index.js"
import UfWorkflowPlugin from "../plugins/uf-workflow/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"

const scratchDirectories: string[] = []

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

type BunRuntime = { readonly YAML: { readonly parse: (text: string) => unknown } }

function withBun<T>(run: () => T | Promise<T>): Promise<T> {
  const runtime = globalThis as typeof globalThis & { Bun?: BunRuntime }
  const original = runtime.Bun
  // The production plugin reads the narrow Bun.YAML.parse boundary. This stub
  // supplies an in-process parser so the smoke test never touches a provider,
  // package registry, GitHub, or the network.
  runtime.Bun = { YAML: { parse: (text: string): unknown => JSON.parse(text) as unknown } }
  const finish = (): void => {
    runtime.Bun = original
  }
  try {
    return Promise.resolve(run()).finally(finish)
  } catch (error) {
    finish()
    throw error
  }
}

/** Builds a target-project-shaped `.opencode/` tree with real policy files. */
async function scratchProject(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-scratch-smoke-"))
  scratchDirectories.push(directory)
  await mkdir(join(directory, ".opencode"), { recursive: true })
  await mkdir(join(directory, ".uf"), { recursive: true })
  await writeFile(join(directory, ".uf", "review-matrix.yaml"), JSON.stringify(matrixFixture))
  await writeFile(join(directory, ".uf", "reviewer-capabilities.yaml"), JSON.stringify(manifestFixture))
  await writeFile(join(directory, ".uf", "sibling-repos.yaml"), JSON.stringify({ version: 1, siblings: [] }))
  return directory
}

describe("provider-free scratch-repository smoke test", () => {
  it("activates both plugins in a target project and registers every tool with zero network", async () => {
    await withBun(async () => {
      const project = await scratchProject()

      const invokeHooks = await InvokeAgentPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)
      const dispatchHooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      expect(Object.keys(invokeHooks.tool)).toEqual(["invoke_agent"])
      expect(Object.keys(dispatchHooks.tool).sort()).toEqual([
        "acquire_sibling_evidence",
        "build_review_prompts",
        "capture_review_diff",
        "consolidate_dispatch",
        "discard_orphaned_submissions",
        "dispatch_agent_run",
        "dispatch_review_runs",
        "dispatch_status",
        "fetch_failed_run_logs",
        "fetch_pr_checks",
        "fetch_pr_review_context",
        "finalize_review_dispatch",
        "plan_review_dispatch",
        "prepare_lesson_learning",
        "read_assessment_fields",
        "read_diff_chunk",
        "redispatch_failed_runs",
        "submit_lesson_proposal",
        "submit_review_findings",
      ])
    })
  })

  it("produces a deterministic plan from scratch-repository policy files", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const dispatchHooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const output = await dispatchHooks.tool.plan_review_dispatch.execute(
        {
          mode: "code",
          discovered_agents: ["divisor-guard", "divisor-adversary"],
          changed_files: [{ path: ".opencode/agents/guard.md", additions: 3, deletions: 0 }],
        },
        {} as never,
      )
      const plan = JSON.parse(output) as { plan_version: number; status: string; entries: Array<{ agent: string; decision: string }> }
      expect(plan.plan_version).toBe(1)
      expect(plan.status).toBe("ready")
      const includedAgents = [...new Set(plan.entries.filter((entry) => entry.decision === "include").map((entry) => entry.agent))].sort()
      expect(includedAgents).toEqual(["divisor-adversary", "divisor-guard"])
    })
  })

  it("never emits the invoke_agent permission contract into child requests", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await InvokeAgentPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      expect(hooks.tool.invoke_agent.args).not.toHaveProperty("permission")
      expect(hooks.tool.invoke_agent.args).not.toHaveProperty("tools")
    })
  })

  it("activates uf-workflow plugin and registers all workflow tools", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await UfWorkflowPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      expect(Object.keys(hooks.tool).sort()).toEqual([
        "checkpoint",
        "clear_progress",
        "format_unleash_demo",
        "format_unleash_exit",
        "get_progress",
        "git_commit_no_edit",
        "git_conflict_files",
        "git_fetch",
        "git_merge",
        "git_merge_abort",
        "git_push",
        "git_push_force_with_lease",
        "git_rebase",
        "git_rebase_abort",
        "mark_task_complete",
        "normalize_model_name",
        "parse_tasks",
        "resolve_base_ref",
        "resolve_clarification",
        "scan_clarifications",
        "scan_conflict_markers",
        "uf_delegate_command",
        "uf_unleash_branch_gate",
        "uf_unleash_detect_resume",
        "uf_unleash_resolve_issue",
        "uf_unleash_route_decision",
        "write_unleash_summary",
      ])
    })
  })

  it("registers uf_delegate_command with the documented argument contract on a scratch project", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await UfWorkflowPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      // DE-007 Layer 1 (delegation-skill-enforcement task 4.2): the same
      // argument-set pin as uf-workflow.test.ts, verified on a real
      // scratch-project activation — the registration surface the OpenCode
      // host consumes. The three classification-contract parameters
      // (task_classification, skill, skill_absent_reason) must reach the
      // model-facing schema alongside the eight pre-change parameters,
      // plus the four model-matrix selection parameters (model, variant,
      // profile, pack — model-matrix-delegation task 5.1, UF-FR-002).
      expect(Object.keys(hooks.tool.uf_delegate_command.args).sort()).toEqual([
        "agent",
        "checkpoint_step",
        "command_path",
        "expected_artifact",
        "feature_dir",
        "model",
        "pack",
        "phase",
        "profile",
        "prompt",
        "skill",
        "skill_absent_reason",
        "task_classification",
        "timeout",
        "variant",
      ])
    })
  })

  it("activates uf-finale plugin and registers all 9 finale tools", async () => {
    const project = await scratchProject()
    const hooks = await UfFinalePlugin.server({
      client: { session: {} },
      directory: project,
      worktree: project,
    } as unknown as PluginInput)

    expect(Object.keys(hooks.tool).sort()).toEqual([
      "uf_finale_base_state",
      "uf_finale_ci_poll",
      "uf_finale_commit_context",
      "uf_finale_create_pr",
      "uf_finale_pr_context",
      "uf_finale_push_state",
      "uf_finale_return_to_main",
      "uf_finale_stage",
      "uf_finale_startup",
    ])
  })
})
