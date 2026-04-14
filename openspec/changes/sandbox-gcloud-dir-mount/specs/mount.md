## MODIFIED Requirements

### Requirement: gcloud ADC Credential Mount

When `GOOGLE_APPLICATION_CREDENTIALS` is not set and
`~/.config/gcloud/` exists on the host, the sandbox
MUST mount the entire `~/.config/gcloud/` directory
into the container as a read-write volume.

Previously: mounted only
`application_default_credentials.json` as read-only.

#### Scenario: Authorized user ADC authenticates

- **GIVEN** the engineer has run
  `gcloud auth application-default login`
- **AND** `GOOGLE_APPLICATION_CREDENTIALS` is not set
- **WHEN** the sandbox starts
- **THEN** `~/.config/gcloud/` is mounted read-write
  at `/home/dev/.config/gcloud/` inside the container
- **AND** OpenCode can authenticate to Vertex AI using
  the refresh token flow

#### Scenario: Service account key still works

- **GIVEN** `GOOGLE_APPLICATION_CREDENTIALS` points to
  a service account key file
- **WHEN** the sandbox starts
- **THEN** only the key file is mounted (Strategy 1
  unchanged)

## REMOVED Requirements

None.
