# Artifact Envelope Schema

The artifact envelope is the standard JSON wrapper for all inter-hero
communication in the Unbound Force swarm. Every artifact — regardless
of type — is wrapped in this envelope to provide metadata for routing,
versioning, and provenance tracking.

## Producer

All heroes produce envelopes when writing artifacts.

## Consumers

All heroes consume envelopes when reading artifacts.

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `hero` | string | Producing hero identifier (e.g., `gaze`, `mx-f`) |
| `version` | string | Hero version (semver) |
| `timestamp` | string | ISO 8601 timestamp of artifact creation |
| `artifact_type` | string | Artifact type identifier (e.g., `quality-report`) |
| `schema_version` | string | Schema version (semver) |
| `context` | object | Workflow context metadata |
| `payload` | object | Type-specific payload (validated by type schema) |

## Context Fields (all optional)

| Field | Type | Description |
|-------|------|-------------|
| `branch` | string | Git branch name |
| `commit` | string | Git commit SHA |
| `backlog_item_id` | string | Originating backlog item ID |
| `correlation_id` | string | UUID linking related artifacts |
| `workflow_id` | string | Workflow instance ID |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-21 | Initial release |
