# Feature Specification: Sub-tool Error Reporting

**Feature Branch**: `036-subtool-error-reporting`
**Created**: 2026-06-15
**Status**: Draft
**Input**: User description: "based on the issue 215 of the project's upstream"
**Upstream Issue**: unbound-force/unbound-force#215

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See Why a Sub-tool Failed During Init (Priority: P1)

A developer runs `uf init` to scaffold a new project or
reinitialize an existing one. One or more sub-tool steps
(dewey init, specify init, openspec init, etc.) fail. The
developer needs to see the actual error message from the
failing sub-tool so they can fix the problem without
having to re-run the sub-tool manually.

**Why this priority**: This is the core problem. Every
sub-tool failure currently requires manual re-execution
to discover the root cause. Fixing this eliminates the
most common friction point.

**Independent Test**: Can be fully tested by triggering
a sub-tool failure during `uf init` (e.g., running
`specify init` without the required configuration) and
verifying the error output includes the sub-tool's
actual error message.

**Acceptance Scenarios**:

1. **Given** a project directory where `specify` is not
   installed, **When** the user runs `uf init`, **Then**
   the failure message for the `.specify/` step includes
   the underlying error (e.g., "specify: executable file
   not found in $PATH") rather than just "specify init
   failed".
2. **Given** a project where `dewey index` fails because
   the workspace was not initialized, **When** the user
   runs `uf init`, **Then** the failure message for the
   `dewey index` step includes the sub-tool's error
   output describing the missing workspace.
3. **Given** any sub-tool step that fails during
   `uf init`, **When** the user reads the failure line,
   **Then** the message is sufficient to diagnose the
   problem without re-running the sub-tool manually.

---

### User Story 2 - See Why a Tool Installation Failed During Setup (Priority: P2)

A developer runs `uf setup` to install required tools.
A tool installation step fails or is skipped. The
developer needs to understand why so they can resolve
the issue -- whether that means installing a dependency,
using an alternative package manager, or taking manual
action.

**Why this priority**: Setup failures block onboarding
entirely. The `setup.go` code already captures error
objects in the `stepResult.err` field and prints them,
but the actual command output (stdout/stderr) from
failed installations is still discarded. This story
closes that remaining gap.

**Independent Test**: Can be fully tested by triggering
a setup installation failure (e.g., `brew install` for
a non-existent formula) and verifying the output
includes the package manager's error message.

**Acceptance Scenarios**:

1. **Given** a system where Homebrew is not available but
   `dnf` is, **When** the user runs `uf setup` and a
   tool's Homebrew installation is skipped, **Then** the
   skip message indicates which alternative package
   managers are available on the system.
2. **Given** a tool installation that fails via
   `go install`, **When** the user reads the failure
   output, **Then** the message includes the Go
   compiler's actual error output.
3. **Given** any tool installation step that fails
   during `uf setup`, **When** the user reads the
   failure line, **Then** the message is sufficient
   to understand the root cause.

---

### User Story 3 - Get Detailed Error Output on Demand (Priority: P3)

A developer encounters a failure during `uf init` or
`uf setup` where the default error summary is not
enough to diagnose the problem. The developer wants
to see the full output from the failing sub-tool
command, including both stdout and stderr.

**Why this priority**: While the default error messages
(from P1/P2) will cover most cases, some failures
produce multi-line diagnostic output that is too
verbose for the default summary. A mechanism to access
the full output is needed for complex debugging
scenarios.

**Independent Test**: Can be fully tested by triggering
a sub-tool failure that produces multi-line output and
verifying the full output is accessible through the
detailed output mechanism.

**Acceptance Scenarios**:

1. **Given** a sub-tool failure that produces multi-line
   stderr, **When** the user requests detailed output,
   **Then** the full stderr and stdout from the failed
   command are displayed.
2. **Given** a successful `uf init` run, **When** the
   user requests detailed output, **Then** no additional
   error output is shown (no noise for passing steps).

---

### Edge Cases

- What happens when a sub-tool produces output on both
  stdout and stderr during failure? Both streams MUST
  be captured and included in the error report.
- How does the system handle a sub-tool that hangs or
  times out? The timeout behavior SHOULD be reported
  with a clear message indicating the tool did not
  respond within the expected time. Note: the existing
  `ExecCmd` implementation has no timeout mechanism;
  timeout handling is out of scope for this feature
  and may be addressed in a follow-up.
- What happens when a sub-tool's error output is
  extremely long (hundreds of lines)? The default
  display SHOULD show a truncated summary with the
  last N lines, while the full output is available
  through the detailed output mechanism.
- What happens when a sub-tool fails with a non-zero
  exit code but produces no stderr output? The error
  report MUST include the exit code as a fallback
  diagnostic.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When a sub-tool invocation fails during
  `uf init`, the failure message MUST include the
  sub-tool's actual error output rather than a
  hardcoded summary string.
- **FR-002**: When a sub-tool invocation fails during
  `uf setup`, the failure message MUST include the
  package manager's or tool's actual error output.
- **FR-003**: Error messages from sub-tools MUST
  preserve the original error text without
  reformatting or summarizing it, so that users can
  search for the error message online.
- **FR-004**: When a sub-tool fails with a non-zero
  exit code but produces no stderr, the system MUST
  report the exit code in the failure message.
- **FR-005**: When a sub-tool is skipped due to a
  missing dependency, the skip message SHOULD indicate
  which alternative methods are available on the
  current system (e.g., "Homebrew not available; dnf
  and go install are available").
- **FR-006**: The system SHOULD provide a mechanism for
  users to access the full command output (stdout and
  stderr) from failed sub-tool invocations. Note: for
  this feature, the inline error output with truncation
  (FR-008) serves as the primary mechanism. A dedicated
  `--verbose` flag or `UF_VERBOSE` env var may be added
  in a follow-up.
- **FR-007**: Successful sub-tool steps MUST NOT
  produce additional output beyond the current summary
  line (no regression in default verbosity).
- **FR-008**: When a sub-tool's error output exceeds
  a reasonable length, the default display SHOULD show
  a truncated version with the most diagnostic portion
  (typically the last lines), with the full output
  available through the detailed output mechanism.

### Key Entities

- **Sub-tool Result**: Represents the outcome of a
  sub-tool invocation, including the tool name, action
  status, a human-readable detail message, and the
  captured error output from the sub-process.
- **Step Result**: Represents the outcome of a setup
  installation step, including the tool name, action
  status, detail message, and the captured error from
  the installation command.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of sub-tool failures during
  `uf init` display the sub-tool's actual error
  message, eliminating the need to re-run sub-tools
  manually for diagnosis.
- **SC-002**: 100% of tool installation failures during
  `uf setup` display the package manager's actual error
  output.
- **SC-003**: 100% of sub-tool failure paths in
  `uf init` and `uf setup` include either the
  sub-tool's stderr output or the exit code in the
  failure message.
- **SC-004**: The default output for successful runs
  remains unchanged -- no additional verbosity is
  introduced for passing steps.

## Dependencies & Assumptions

### Dependencies

- None. This feature modifies existing error handling
  within `uf init` and `uf setup` and does not
  introduce new external dependencies.

### Assumptions

- Sub-tool error output is written to stderr and/or
  returned in the error object from command execution.
  The existing `CombinedOutput()` call already captures
  both stdout and stderr; the issue is that this output
  is currently discarded.
- The detailed output mechanism will follow existing
  CLI patterns in the project (e.g., a flag or
  environment variable). The specific mechanism is an
  implementation decision.
- Error output truncation threshold (for FR-008) will
  use a reasonable default. The specific line count is
  an implementation decision.
