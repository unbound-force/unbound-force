---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:28:19Z
identity: test-quality-20261110T202819-jay-flowers
tier: draft
---

Locale-pinned, line-anchored signature matching for subprocess output classification: when a health check classifies a subprocess failure by matching a message signature (e.g. git's "fatal: not a git repository"), run the subprocess under a pinned locale (env LC_ALL=C ...) so gettext ignores LANGUAGE and the signature is stable across user locales, and anchor the signature to the start of an output line with (?m)^...\b so a target path that embeds the phrase inside another error cannot spoof the classification. Cover both properties with dedicated subtests: one asserting the command carries the locale prefix, one feeding a spoofed mid-line occurrence and asserting the failure path (Warn) is taken, and one feeding the signature after a preamble line and asserting the skip branch still fires. The mid-line spoof subtest fails under a naive strings.Contains implementation, making it a genuine regression test.
