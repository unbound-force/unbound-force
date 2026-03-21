# Hero Manifest Schema

## Overview

The hero manifest defines metadata for an Unbound Force hero
repository, including its name, version, constitution version,
capabilities, and integration points.

## Producer

- Any hero repository (Gaze, Website, future heroes)

## Consumers

- `unbound` CLI (contract validation)
- CI pipelines (hero compliance checking)

## Required Fields

- `name` (string): Hero identifier
- `version` (semver): Hero version
- `constitution_version` (semver): Org constitution version
- `capabilities` (array): Hero capabilities

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-02-24 | Initial schema (Spec 002) |
