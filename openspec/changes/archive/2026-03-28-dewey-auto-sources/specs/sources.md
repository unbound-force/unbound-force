## ADDED Requirements

### Requirement: auto-detect-sibling-repos

`uf init` MUST detect sibling repositories in the
parent directory and generate a multi-repo Dewey
sources configuration.

#### Scenario: siblings detected

- **GIVEN** the parent directory contains 3 sibling
  repos with `.git/` directories
- **WHEN** `uf init` runs and `dewey init` succeeds
- **THEN** `.dewey/sources.yaml` contains per-repo
  disk sources for each sibling, plus a disk-org
  source for the parent directory

#### Scenario: no siblings

- **GIVEN** the parent directory contains no other
  repos (only the current project)
- **WHEN** `uf init` runs
- **THEN** `.dewey/sources.yaml` contains only the
  disk-local source and disk-org source

#### Scenario: sources.yaml already customized

- **GIVEN** `.dewey/sources.yaml` has been edited by
  the user (more than 1 source entry)
- **WHEN** `uf init` runs again
- **THEN** the file is NOT overwritten

### Requirement: github-org-detection

`uf init` SHOULD extract the GitHub org name from the
git remote URL and include a GitHub API source in the
generated Dewey sources config.

#### Scenario: SSH remote URL

- **GIVEN** `git remote get-url origin` returns
  `git@github.com:unbound-force/repo.git`
- **WHEN** the GitHub org is extracted
- **THEN** the org name is `unbound-force`

#### Scenario: HTTPS remote URL

- **GIVEN** `git remote get-url origin` returns
  `https://github.com/unbound-force/repo.git`
- **WHEN** the GitHub org is extracted
- **THEN** the org name is `unbound-force`

#### Scenario: non-GitHub remote

- **GIVEN** `git remote get-url origin` returns a
  non-GitHub URL
- **WHEN** the GitHub org extraction runs
- **THEN** the GitHub API source is omitted (no error)

#### Scenario: no remote

- **GIVEN** `git remote get-url origin` fails
- **WHEN** the GitHub org extraction runs
- **THEN** the GitHub API source is omitted (no error)

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
