---
tag: stdin-prompt
category: gotcha
created_at: 2026-05-07T23:20:29Z
identity: stdin-prompt-1
tier: draft
---

When writing interactive confirmation prompts in Go CLI tools, avoid fmt.Fscanln for reading user input from stdin. fmt.Fscanln uses whitespace-delimited token scanning semantics, not line-oriented reading. Under certain terminal configurations (particularly on macOS with iTerm2), pressing Enter can send a bare carriage return (\r, 0x0D) instead of a newline (\n, 0x0A). fmt.Fscanln does not recognize \r as a line terminator, causing it to block indefinitely while the terminal echoes ^M. The correct approach is bufio.NewScanner(stdin) with scanner.Scan() and scanner.Text(). bufio.Scanner's default ScanLines split function handles \n, \r\n, and bare \r line endings correctly. scanner.Text() returns the line with the terminator already stripped. Always combine with strings.TrimSpace for defense against trailing whitespace. This pattern preserves DI testability since bufio.Scanner accepts any io.Reader, allowing tests to inject strings.NewReader with various line endings for regression testing.
