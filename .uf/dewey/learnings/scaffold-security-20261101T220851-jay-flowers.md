---
tag: scaffold-security
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:08:51Z
identity: scaffold-security-20261101T220851-jay-flowers
tier: draft
---

When a change embeds and deploys an agent definition to downstream repositories, verify the security-relevant frontmatter travels with the byte-identical mirror: the permission block (edit/bash/webfetch deny), mode, and hidden flags are the deployed access-control contract, and a mirror that diverges silently weakens least privilege in every scaffolded repo. Review technique that needs no file reads: compare the diff's git index blob hashes — a modified canonical agent file's post-image hash in its `index <old>..<new>` line against the new mirror's `index 0000000..<new>` line — and confirm the drift-test table registers the pair so future divergence fails CI. In the model-matrix-scaffold-deployment review (#889) the classifier mirror's deny-all permission set was verified this way (hash match plus task42CanonicalAssets registration); when the canonical agent file is unchanged and absent from the diff (as with model-matrix-classifier.md), the drift-test registration plus green CI is the enforcement evidence, and the reviewer should confirm the registration line explicitly names the pair.
