// Shared CLI plumbing for the repo-local lib CLIs. Every runnable lib
// module needs the same root discovery, git access, error detail, and
// process wiring; one copy keeps the copies from drifting (CS-004, DE-001).
import { execFileSync } from "node:child_process"
import { existsSync, statSync } from "node:fs"
import { join, resolve } from "node:path"

/** Writers the CLI prints through; injected so tests can capture. */
export interface CliIO {
  cwd: string
  out: (s: string) => void
  err: (s: string) => void
}

/** Nearest ancestor directory holding a `.opencode` directory, or null.
 * A file named `.opencode` does not terminate the walk.
 * @param start Directory to start walking from.
 * @returns The root directory, or null at the filesystem root. */
export function findRoot(start: string): string | null {
  let dir = resolve(start)
  for (;;) {
    const candidate = join(dir, ".opencode")
    if (existsSync(candidate) && statSync(candidate).isDirectory()) return dir
    const parent = resolve(dir, "..")
    if (parent === dir) return null
    dir = parent
  }
}

/** Run git in `root`, capturing stdout; `input` feeds stdin.
 * @param root Directory to run git in.
 * @param args Git arguments, e.g. `["merge-base", ref, "HEAD"]`.
 * @param input Optional stdin payload.
 * @returns Captured stdout. */
export function git(root: string, args: string[], input?: string): Buffer {
  return execFileSync("git", args, {
    cwd: root,
    input,
    maxBuffer: 256 * 1024 * 1024,
    stdio: ["pipe", "pipe", "pipe"],
  })
}

/** Human-readable detail from a failed spawned process.
 * @param e The thrown value (execFileSync error or any throw).
 * @returns Trimmed stderr when present, else the error message; a
 *   non-Error throw stringifies instead of printing "undefined". */
export function spawnDetail(e: unknown): string {
  return (e as { stderr?: Buffer }).stderr?.toString("utf8").trim() || (e instanceof Error ? e.message : String(e))
}

/** Wire `main` as the module's runnable entry point: slice argv, inject
 * process-bound writers (one newline per write), record the exit code.
 * @param main The CLI entry function to wire.
 * @returns Nothing; the exit code lands on `process.exitCode`. */
export function runMain(main: (argv: string[], io: CliIO) => number): void {
  const code = main(process.argv.slice(2), {
    cwd: process.cwd(),
    out: (s: string): void => {
      process.stdout.write(s + "\n")
    },
    err: (s: string): void => {
      process.stderr.write(s + "\n")
    },
  })
  process.exitCode = code
}
