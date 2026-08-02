---
title: Multi-Turn Sessions
description: Resuming conversations across discrete invocations.
icon: forum
weight: 600
---

Harness session state lives on disk in each harness's own store, so a
conversation is a sequence of discrete invocations naming the same session:

1. Turn 1: `run` (claude-code: preset `--session-id` to know the ref up
   front).
2. Get the session ref in-band from stdout with JSON output: the claude-code
   and antigravity envelopes and the codex event stream all carry the
   session/conversation ID. Antigravity's envelope landed in agy 1.1.8 (the
   [upstream request](https://github.com/google-antigravity/antigravity-cli/issues/7)
   for an in-band ref); on older releases, or in text mode, the ref is only
   discoverable post-hoc with `agentminutes sessions`.
3. Turn N: `run --resume <ref>`.

Resume appends rather than forks on all three harnesses: same transcript
file (claude-code, codex) or conversation directory (antigravity), same
session identity. The ref from turn 1 stays valid for every later turn.

One antigravity trap when discovering refs post-hoc: a single `-p`
invocation writes **two** conversations (a warm-up plus the real one), so
discovery must match candidates on content (e.g. the recorded prompt),
never on recency alone.

agentsummons requires an explicit ref; the harnesses' continue-most-recent
forms are deliberately unmodeled, because they're racy against any concurrent
agent session. However, if you prefer to use the harnesses' continue-most-recent
forms, instead, they remain reachable with `--extra-arg`.

## Example: preset the ref (Claude Code)

Claude Code accepts a session ID up front, so the simplest multi-turn shape
generates the ref first and never has to parse anything:

```sh
SESSION=$(uuidgen | tr '[:upper:]' '[:lower:]')

# Turn 1: preset the session identity.
agentsummons run --harness claude-code --session-id "$SESSION" \
  -p "Read the failing test and propose a fix" --auto-approve

# Turn 2: resume the same session.
agentsummons run --harness claude-code --resume "$SESSION" \
  -p "Apply the fix and run the tests" --auto-approve
```

Because the ref is preset, the transcript is addressable before turn 1
even starts, which is useful when something else (archiving, a watchdog)
needs to find the session while it's still running.

## Example: capture the ref in-band (Antigravity)

Antigravity CLI doesn't expose the ability to preset an ID, so
turn 1 must capture the ref from the harness's own JSON envelope instead.
`--json-output` asks the harness for machine-readable stdout, which
passes through untouched, so `jq` can read the `conversation_id` field
directly:

```sh
# Turn 1: run with harness JSON output and capture the conversation ref.
REF=$(agentsummons run --harness antigravity \
  -p "Outline a test plan for the parser" \
  --auto-approve --json-output | jq -r '.conversation_id')

# Turn 2: resume that conversation.
agentsummons run --harness antigravity --resume "$REF" \
  -p "Expand the edge cases into concrete test names" --auto-approve
```

The behaviors these examples encode:

- The caller stores one string and passes it back as `--resume`. Which
  flag that becomes (`--resume`, `--conversation`, or codex's
  `exec resume` subcommand) is agentsummons's problem.
- The ref extraction differs per harness, and only where it has to: the
  codex variant of turn 1 reads the session ID from its JSONL event
  stream rather than a single envelope.
- `Result.SessionID` (and the `session_id` field in `run --json`'s
  envelope) echoes a preset identity. agentsummons never parses refs out
  of harness output for you; that stays in-band, in the harness's own
  format.
