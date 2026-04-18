## MODIFIED Requirements

### Requirement: Container User Namespace Mapping

On Linux, `buildRunArgs()` MUST include
`--userns=keep-id` in the Podman run arguments. This
maps the host user's UID/GID into the container so
bind mount permissions work correctly.

On macOS, `--userns=keep-id` MUST NOT be included
(Podman's VM layer handles UID mapping).

#### Scenario: Linux bind mount write succeeds

- **GIVEN** the host user is UID 1000
- **AND** the platform is Linux
- **WHEN** the sandbox starts with `--mode direct`
- **THEN** `--userns=keep-id` is included in the
  Podman arguments
- **AND** the container process can write to the
  mounted project directory

#### Scenario: gcloud token refresh succeeds on Linux

- **GIVEN** `~/.config/gcloud/` is mounted read-write
- **AND** the platform is Linux
- **WHEN** the auth library refreshes an OAuth2 token
- **THEN** the refreshed token is written to
  `access_tokens.db` successfully

#### Scenario: macOS unaffected

- **GIVEN** the platform is macOS
- **WHEN** the sandbox starts
- **THEN** `--userns=keep-id` is NOT included in the
  Podman arguments

## REMOVED Requirements

None.
