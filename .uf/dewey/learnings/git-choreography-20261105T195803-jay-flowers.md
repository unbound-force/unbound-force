---
tag: git-choreography
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:58:03Z
identity: git-choreography-20261105T195803-jay-flowers
tier: draft
---

git fetch <url> <ref> (a refspec fetched from a URL rather than a named remote) updates only FETCH_HEAD — it never updates a remote-tracking ref, because opportunistic tracking-ref updates apply only to configured named remotes. Any divergence gate that follows a fetch must therefore compare explicitly against FETCH_HEAD (git rev-list --left-right --count HEAD...FETCH_HEAD; the right count is remote-only commits) rather than relying on git status or @{upstream}, both of which read remote-tracking refs that URL fetches leave stale or absent. Repo precedent: uf.rebase-pr Phase 3 and its OpenSpec requirement consume FETCH_HEAD for URL-fetched PR heads; /uf.address-feedback Section 4.4 (PR 955, feedback item 3) shipped a git status-keyed divergence gate that silently no-ops when HEAD_REMOTE is a URL — a value Section 1.2.1 explicitly permits.
