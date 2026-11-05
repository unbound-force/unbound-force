---
tag: delegation-contracts
author: jay-flowers
category: pattern
created_at: 2026-11-03T16:46:15Z
identity: delegation-contracts-20261103T164615-jay-flowers
tier: draft
---

Composition-site completeness for unstated-constraint fixes: when a fix documents a tool-enforced constraint (e.g., a 512-char maxLength that models violate) at the prose sites where agents compose the field, grep the command file for every occurrence of the field name and annotate each site — including templates for sibling tools that "enforce the same parameter contract" by reference, since they carry the identical limit and the identical failure mode. A CHANGELOG note that a template is "unchanged" documents that decision and keeps the sibling sites discoverable at review time.
