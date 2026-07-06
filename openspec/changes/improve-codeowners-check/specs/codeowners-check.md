## ADDED Requirements

### Requirement: docs-codeowners-path

The CODEOWNERS check MUST search `docs/CODEOWNERS` in
addition to `CODEOWNERS` and `.github/CODEOWNERS`.

#### Scenario: CODEOWNERS exists in docs directory

- **GIVEN** a repository with a `docs/CODEOWNERS` file and
  no `CODEOWNERS` or `.github/CODEOWNERS` file
- **WHEN** the agent performs the CODEOWNERS existence check
- **THEN** the agent MUST detect the file and display the
  CODEOWNER review warning if `require_code_owner_reviews`
  is true

### Requirement: non-404-error-warning

When a `gh api` call returns a non-404 error during the
CODEOWNERS check, the agent MUST display an informational
warning indicating the check was inconclusive.

#### Scenario: API returns 500 server error

- **GIVEN** a repository with `require_codeowners` enabled
- **WHEN** the agent calls `gh api` to check for CODEOWNERS
  and receives a 500 error
- **THEN** the agent MUST display:
  "Note: CODEOWNERS check was inconclusive (API error).
  Could not determine if this repo uses CODEOWNERS."
- **AND** the agent MUST NOT display the CODEOWNER review
  warning (since existence is unknown)

#### Scenario: API returns 429 rate limit

- **GIVEN** a repository with `require_codeowners` enabled
- **WHEN** the agent calls `gh api` to check for CODEOWNERS
  and receives a 429 rate-limit error
- **THEN** the agent MUST display the inconclusive warning
- **AND** the agent MUST NOT skip the check silently

### Requirement: short-circuit-on-success

The CODEOWNERS check SHOULD short-circuit on the first
successful path. If `.github/CODEOWNERS` is found, the
agent SHOULD NOT check `CODEOWNERS` or `docs/CODEOWNERS`.

#### Scenario: CODEOWNERS found at first path

- **GIVEN** a repository with `.github/CODEOWNERS`
- **WHEN** the agent performs the CODEOWNERS existence check
- **THEN** the agent SHOULD issue only one `gh api` call
- **AND** the agent MUST treat CODEOWNERS as found

## MODIFIED Requirements

### Requirement: codeowners-error-handling

The CODEOWNERS check MUST distinguish between 404 responses
(file not found) and other error responses. 404 errors MUST
be handled silently. Non-404 errors MUST produce a visible
warning.

Previously: "If any API call fails: skip silently."

#### Scenario: 404 response handled silently

- **GIVEN** a repository with no CODEOWNERS file at any of
  the three valid paths
- **WHEN** each `gh api` call returns a 404 response
- **THEN** the agent MUST treat this as "no CODEOWNERS file"
- **AND** the agent MUST NOT display any error or warning
  about the check itself

### Requirement: scaffold-sync

Active command copies MUST match their scaffold source copies.
Both `/review-pr` and `/review-council` MUST use identical
CODEOWNERS check logic.

Previously: The commands used identical logic but this was
not an explicit requirement.

#### Scenario: scaffold drift detection

- **GIVEN** updated active copies of `review-pr.md` and
  `review-council.md`
- **WHEN** scaffold drift detection tests run
- **THEN** the active copies MUST match the scaffold source
  copies under `internal/scaffold/assets/opencode/commands/`

## REMOVED Requirements

None.
