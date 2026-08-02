---
title: Harness Capabilities
description: The capability matrix for Antigravity CLI, Claude Code, and Codex CLI.
icon: hub
weight: 500
---

## Capability matrix

Each harness exposes different flags that may be relevant when your project
needs a cross-harness runner, such as:

| Harness | Binary | Model | Session ID | Resume | Allowed tools | JSON output | Auto-approve |
|---|---|---|---|---|---|---|---|
| antigravity | `agy` | `--model "<display name>"` | (none) | `--conversation <id>` | (none) | `--output-format json` (envelope) | `--dangerously-skip-permissions` |
| claude-code | `claude` | `--model <id>` | `--session-id <uuid>` | `--resume <session-id>` | `--allowedTools a,b` | `--output-format json` (envelope) | `--dangerously-skip-permissions` |
| codex | `codex` | `-m <id>` | (none) | `exec resume <session-id>` | (none) | `--json` (JSONL events) | `--dangerously-bypass-approvals-and-sandbox` |

Use the `agentsummons info` command to view the full manifests, including
per-harness quirk notes, such as TLS trust-store behavior, headless tool
registration, and stdout reliability.

## Companion tooling

agentsummons knows nothing about transcripts by design. Its contract ends
at "the command ran; here is everything observable about that". A `Result`'s
timestamps and workdir are the inputs that companion tool
[agentminutes](https://github.com/agent-ecosystem/agentminutes) `sessions`
needs to find what the harness wrote.
