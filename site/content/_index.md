---
title: agentsummons
description: Go library + CLI for invoking agent harnesses (Antigravity CLI, Claude Code, Codex CLI) in headless mode.
---

agentsummons is a Go library and CLI for invoking agent harnesses (Antigravity
CLI, Claude Code, Codex CLI) in headless mode. Every headless-agent experiment
rediscovers the same lore: which binary, which permission-bypass flag, which
argument ordering. agentsummons owns that knowledge once, behind one Go API
and one CLI.

It is the companion to
[agentminutes](https://github.com/agent-ecosystem/agentminutes), which parses
the session transcripts harnesses write: agentsummons convenes the meeting,
agentminutes takes the minutes.
