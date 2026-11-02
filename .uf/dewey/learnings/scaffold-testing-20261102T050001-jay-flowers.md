---
tag: scaffold-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:00:01Z
identity: scaffold-testing-20261102T050001-jay-flowers
tier: draft
---

Stealth-mode Run() tests that assert a file was NOT deployed must stub LookPath — a real opencode binary found on PATH runs real `opencode debug config` probes during plugin activation (defaulted ExecCmdInDir) and opencode itself writes .opencode/.gitignore into the target, a foreign actor that masquerades as a scaffold deployment and makes the test machine-dependent (discovered in init-override-stubs-gitignore-sync task 5.3, TestRun_OpencodeGitignore_StealthSkipsIt; npm ci only creates node_modules, verified empirically). The inverse trap: stubbing LookPath empty makes activation fail with a non-nil err, and Run()'s stealth arm treats any subResult.err as fatal ("stealth init sub-tool %q failed") — the run aborts and rolls back. The deterministic wiring sits between the two: LookPath finds only node and npm (opencode absent, the CI-machine shape), ExecCmdInDir answers the version probes and npm ci with canned output, so activation lands on its err-free "skipped" disposition and stealth proceeds. Also note runScaffoldWithoutExternalTools preserves a pre-set ExecCmd, so a stealth test can pass the stealthGitFake as ExecCmd and still get the LookPath/ExecCmdInDir guards defaulted. Pre-existing stealth tests wired with a real LookPath (TestRun_OverrideStubs_StealthWritesNeitherFile) pass only accidentally — their asserted paths happen not to collide with what opencode/npm write.
