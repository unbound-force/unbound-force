---
tag: cli-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-08T20:00:30Z
identity: cli-tooling-20261108T200030-jay-flowers
tier: draft
---

When a shared CLI module is extracted to stop two same-PR CLIs from drifting (PR 959's cli-common.ts: findRoot, git, spawnDetail, runMain), the text-output sanitization boundary is part of that shared plumbing surface. instruction-lint.ts kept its escapeControlChars print-site sanitizer private while the new tool-inventory.ts shipped three text-mode print-site groups without one, re-opening the terminal-control-character drift the shared module was created to close — and neither test suite caught it, because the drifted edge (a control character in a file-derived field) was untested in the new CLI. Extraction scope for shared CLI modules must include the output-sanitization helper, landed with its own unit tests (ESC/OSC sequences, C1 bytes, newline/tab preserved), so both CLIs inherit the guarantee the same way they inherit runMain's exit-code contract; a sanitization helper left private to the first CLI guarantees the second CLI ships unsanitized.
