<!-- capability: scaffold
     change: heredoc-ban-escaped-and-ansic
     originating_issue: 1048 -->

## 1. Contract test hardening

- [x] 1.1 Update the heredoc regex in `internal/scaffold/command_contract_test.go:858` from `<<-?\s*['"]?[A-Za-z_][A-Za-z0-9_]*['"]?` to the canonical SC-FR-016 spelling `<<-?\s*\$?\\?\s*['"]?[A-Za-z_][A-Za-z0-9_]*['"]?` (covers `<<\EOF` escaped-delimiter, `<<\ 'EOF'` spaced escaped-delimiter, and `<<$'EOF'` ANSI-C-quoted forms; preserves `<<-` dashed and quoted/unquoted existing coverage)
- [x] 1.2 Add table-driven subtest `TestCommandContracts_HeredocPattern` (or equivalent name) adjoining the scan loop that pins all 15 positive spellings -- `<<EOF`, `<<'EOF'`, `<<"EOF"`, `<<-EOF`, `<<-'EOF'`, `<<-"EOF"`, `<<\EOF`, `<<\ 'EOF'`, `<<\'EOF'`, `<<-\EOF`, `<<-\ 'EOF'`, `<<$'EOF'`, `<<-$'EOF'`, `<<< "foo"`, `<<< 'bar'` -- each case MUST match the compiled pattern
- [x] 1.3 Extend the table with all 4 negative spellings -- `>>> MANDATORY GATE <<<`, `>>> MANDATORY GATE: HUMAN CONFIRMATION REQUIRED <<<`, `<<<<<<<`, `>>>>` -- each case MUST NOT match the pattern; the table MUST assert all 15 positive and 4 negative cases (19 total) so silent case loss fails visibly
- [x] 1.4 Run `go test -race -count=1 ./internal/scaffold -run TestCommandContracts_RejectShellHeredoc` and `go test -race -count=1 ./internal/scaffold -run TestCommandContracts_HeredocPattern` and verify both pass; confirm `make coverage-gate` scope-based gate still passes
- [x] 1.5 Add a `t.TempDir()` synthetic-violation fixture test (e.g., `TestCommandContracts_HeredocScanFixture`) that plants a heredoc-bearing markdown file on a known line and asserts the scan reports exactly one violation whose message names the file, line number, and matched text and directs the author to the host's file-write tool; the extracted scan helper MUST fail closed on walk errors (missing directory, unreadable file, symlink loop, permission denied) per SC-FR-016

## 2. Verification

- [x] 2.1 Verify constitution alignment: re-read proposal's Constitution Alignment against the updated Constitution Alignment six-principle assessment and confirm no implementation drift (test-only change, no new artifact coupling, security-structural PASS per Principle V)
- [x] 2.2 Run `make check` (lint + vet + `go test -race -count=1 ./...`) before PR to ensure no vet/golint regression from the test addition
