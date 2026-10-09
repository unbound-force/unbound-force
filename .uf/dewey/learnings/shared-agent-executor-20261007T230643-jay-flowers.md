---
tag: shared-agent-executor
author: jay-flowers
category: pattern
created_at: 2026-10-07T23:06:43Z
identity: shared-agent-executor-20261007T230643-jay-flowers
tier: draft
---

When extracting a shared library from an existing OpenCode plugin (like agent-executor.ts from invoke-agent), the re-export pattern is critical for backward compatibility. Every type and function that was previously importable from the original module must be re-exported from it so downstream consumers do not break. The invoke-agent refactoring moved types like ModelIdentity, InvocationProvenance, InvokeAgentResult and functions like sanitizeInvocationError to agent-executor.ts but preserved re-exports in invoke-agent/index.ts. This pattern was validated by 302 tests continuing to pass without any import changes in existing test files. The key design insight is that the shared executor (executeAgentSession) encapsulates the child session lifecycle (abort controller, create, prompt, response parsing, timeout, cleanup) while the caller retains input validation, model resolution, and prompt resolution — these are caller-specific concerns that differ between invoke_agent (host model introspection) and dispatch_agent_run (review matrix tier lookup).
