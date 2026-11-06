// Shared CLI plumbing for the repo-local lib CLIs. Every runnable lib
// module needs the same root discovery, git access, error detail, and
// process wiring; one copy keeps the copies from drifting (DE-003).
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
 * A file named `.opencode` does not terminate the walk. */
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

/** Run git in `root`, capturing stdout; `input` feeds stdin. */
export function git(root: string, args: string[], input?: string): Buffer {
  return execFileSync("git", args, {
    cwd: root,
    input,
    maxBuffer: 256 * 1024 * 1024,
    stdio: ["pipe", "pipe", "pipe"],
  })
}

/** Human-readable detail from a failed spawned process. */
export function spawnDetail(e: unknown): string {
  return (e as { stderr?: Buffer }).stderr?.toString("utf8").trim() || (e as Error).message
}

/** Wire `main` as the module's runnable entry point: slice argv, inject
 * process-bound writers (one newline per write), record the exit code. */
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
