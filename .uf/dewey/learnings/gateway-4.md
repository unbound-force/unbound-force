---
tag: gateway
category: gotcha
created_at: 2026-04-28T00:44:51Z
identity: gateway-4
tier: draft
---

When the uf gateway runs in detached mode (via uf gateway start --detach or auto-started by uf sandbox start), the child process re-execs with Setsid: true for session isolation. Previously cmd.Stdout and cmd.Stderr were set to nil, discarding all charmbracelet/log output from the detached process. This meant token refresh failures, provider errors, and upstream diagnostics were silently lost. The fix redirects child output to .uf/gateway.log (O_CREATE|O_WRONLY|O_TRUNC, 0600 permissions for security since the log contains auth diagnostics). The parent process closes its file handle after ExecStart since the child inherits the fd independently. The uf gateway status command now shows the log file path when it exists, guiding users to diagnostics during troubleshooting.
