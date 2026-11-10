---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-11-10T16:54:58Z
identity: scaffold-mirrors-20261110T165458-jay-flowers
tier: draft
---

Mirror-parity verification at review time can be cryptographic instead of visual: in a unified diff, each file header carries `index <old-blob>..<new-blob>`, so a canonical scaffold asset and its embedded mirror are byte-identical if and only if their `index` lines match. Verified on PR #1007 (feedback-reply-tools): the command pair (bd6092719..be8075a69), lib pair (000000000..489332d42), and plugin pair (65e2a498b..9e585e327) each appeared at both the canonical and mirror diff locations, proving parity without re-reading a 41KB mirror chunk. Grep the saved diff file for the canonical index line and count occurrences (2 = parity, 1 = drift); for new-file mirrors, read the mirror's index line directly from the diff file. Combine with the repo's drift-test coverage (task42CanonicalAssets / assertCommandCopies) plus green CI for pairs the tests pin — the blob-hash check covers pairs whose drift enforcement is uncertain.
