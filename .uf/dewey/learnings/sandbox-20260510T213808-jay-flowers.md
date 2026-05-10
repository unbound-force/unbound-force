---
tag: sandbox
author: jay-flowers
category: pattern
created_at: 2026-05-10T21:38:08Z
identity: sandbox-20260510T213808-jay-flowers
tier: draft
---

DevPod proxies git credentials into containers via its own credential helper, but the gh CLI maintains a separate token store at ~/.config/gh/hosts.yml. To enable gh commands (PR reviews, issue management) inside DevPod sessions, the devcontainer postStartCommand can extract the GitHub token from DevPod's git credential proxy using 'printf protocol=https\nhost=github.com\n\n | git credential fill' and pipe it to 'gh auth login --with-token'. Using postStartCommand (not postCreateCommand) ensures the token is refreshed on every container start. The || true suffix provides graceful degradation when running outside DevPod or when gh is not installed. The token flows exclusively through stdin pipes (never as a command-line argument), preventing /proc/cmdline exposure.
