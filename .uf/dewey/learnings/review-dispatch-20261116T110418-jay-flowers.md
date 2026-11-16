---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-16T11:04:18Z
identity: review-dispatch-20261116T110418-jay-flowers
tier: draft
as_of: "unbound-force@1d0e5cef2b10545a779afe8aa2c925fb2ab22f9e"
---

Manifest-delivered review dispatch can leave the working tree at the PR base SHA: instruction-lint --base and tool-inventory --seams then run clean (exit 0, valid JSON) but report empty delta.files and empty seams because tree == merge base. Verified on PR #1147. Review consequence: source head-state content from read_diff_chunk; unchanged files on disk are valid for both states (they did not change); DE-007 falls back to reading whether the change's own tests reference the tool; and the linter's whole-tree token map still corroborates a recorded DE-004 before-value — it reported 16,738 tokens / 23,306 load tokens for uf.address-feedback.md at base, exactly matching the CHANGELOG's recorded 16,738→16,841 (23,306→23,409) delta — so the before-half of a recorded instruction delta is verifiable even when the after-half is not.
