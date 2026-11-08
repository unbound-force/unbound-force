// Shared CLI plumbing for the repo-local lib CLIs. Every runnable lib
// module needs the same root discovery, git access, error detail,
// entry-point detection, and process wiring; one copy keeps the
// copies from drifting (CS-004, DE-001).
import { execFileSync } from "node:child_process"
import { existsSync, realpathSync, statSync } from "node:fs"
import { join, resolve } from "node:path"
import { fileURLToPath } from "node:url"

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

/** Escape C0 and C1 control characters (tab and newline excluded) as
 * backslash-x hex escapes so untrusted file content cannot inject
 * terminal control sequences into printed output (Constitution V).
 * @param text Text to escape.
 * @returns Text with control characters escaped. */
export function escapeControlChars(text: string): string {
  return text.replace(/[\u0000-\u0008\u000b-\u001f\u007f-\u009f]/g, (c) => `\\x${c.charCodeAt(0).toString(16).padStart(2, "0")}`)
}

/** True when this module is the process entry point. Compared by path
 * rather than `import.meta.main`, which is undefined on runtimes older
 * than Node 24.2: a CLI would otherwise load, do nothing, and exit 0
 * there. Callers pass their own `import.meta.url`.
 * @param moduleUrl The module's `import.meta.url`.
 * @returns Whether the module is the process entry point. */
export function isMainModule(moduleUrl: string): boolean {
  const entry = process.argv[1]
  if (entry === undefined || !existsSync(entry)) return false
  try {
    return realpathSync(entry) === fileURLToPath(moduleUrl)
  } catch {
    return false
  }
}

/** Human-readable detail from a failed spawned process.
 * @param e The thrown value (execFileSync error or any throw).
 * @returns Trimmed stderr when present, else the error message; null
 *   and undefined throws stringify instead of raising a TypeError. */
export function spawnDetail(e: unknown): string {
  const stderr = e === null || e === undefined ? undefined : (e as { stderr?: unknown }).stderr
  const text = Buffer.isBuffer(stderr) ? stderr.toString("utf8").trim() : typeof stderr === "string" ? stderr.trim() : ""
  return text || (e instanceof Error ? e.message : String(e))
}

/** Wire `main` as the module's runnable entry point: slice argv, inject
 * process-bound writers (one newline per write), record the exit code.
 * A crash converts to exit code 2 — the could-not-run convention the
 * linter documents — instead of exiting 1 and masquerading as findings.
 * @param main The CLI entry function to wire.
 * @param name Tool name prefixing the crash message.
 * @returns Nothing; the exit code lands on `process.exitCode`. */
export function runMain(main: (argv: string[], io: CliIO) => number, name: string): void {
  try {
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
  } catch (e) {
    // Crash paths (EACCES, EISDIR, ENOTDIR, ...) must not masquerade
    // as findings: exit 2 is the could-not-run code.
    process.stderr.write(`${name}: could not run: ${escapeControlChars(e instanceof Error ? e.message : String(e))}\n`)
    process.exitCode = 2
  }
}
