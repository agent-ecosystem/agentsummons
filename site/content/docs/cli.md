---
title: CLI
description: The run, build, info, and doctor commands, exit codes, and the JSON contract.
icon: terminal
weight: 300
---

## agentsummons run

Invoke a harness with:

```sh
agentsummons run --harness claude-code -p "Summarize this repo" \
  --allowed-tools Read,Grep,Glob --auto-approve
```

agentsummons defaults to a transparent wrapper: harness stdout/stderr pass
through, and the harness exit code propagates.

When you invoke agentsummons with `--json`, stdout carries a single result
envelope with the harness output embedded. The envelope stamps the installed
harness version and a `drift_hint` when that version is newer than the flag
surface this build was validated against.

For example, this invocation presets a session ID and asks for the envelope:

```sh
agentsummons run --harness claude-code -p "Summarize this repo" \
  --allowed-tools Read,Grep,Glob --auto-approve \
  --session-id 3f7d0c2e-9a41-4e5b-8f26-71d3f0b6c9e4 --json
```

It produces one envelope on stdout:

```json
{
  "schema_version": 1,
  "harness": "claude-code",
  "harness_version": "2.1.205",
  "argv": [
    "claude",
    "--dangerously-skip-permissions",
    "--session-id", "3f7d0c2e-9a41-4e5b-8f26-71d3f0b6c9e4",
    "--allowedTools", "Read,Grep,Glob",
    "-p", "Summarize this repo"
  ],
  "prompt_index": 7,
  "workdir": "/Users/me/project",
  "start": "2026-08-02T17:31:04.128Z",
  "end": "2026-08-02T17:31:19.804Z",
  "exit_code": 0,
  "session_id": "3f7d0c2e-9a41-4e5b-8f26-71d3f0b6c9e4",
  "stdout": "This repo contains a Go library that ...",
  "stderr": ""
}
```

A few things the envelope encodes:

- `argv` is the exact command agentsummons assembled, so you can see
  (and replay) precisely what ran. Note the translation: `--auto-approve`
  became Claude's `--dangerously-skip-permissions`, and `--allowed-tools`
  became `--allowedTools`.
- `prompt_index` points at the prompt's position in `argv`, so log
  scrubbers and dashboards can find or redact it without parsing flags.
- `stdout` and `stderr` carry the harness output verbatim. Nothing is
  interpreted or dropped.
- `session_id` echoes the identity the session ran under, which is what
  you pass to `--resume` for a follow-up turn.
- `drift_hint` and `timed_out` appear only when relevant: a harness
  version past validated coverage, or a run killed by `--timeout`.

## agentsummons build

To check harness commands without executing, use the `build` command:

```sh
agentsummons build --harness antigravity -p "hi" --auto-approve
```

This prints the exact argv, directory, and environment `run` would use.
This is the correct-assembly path for callers in other languages:
ordering quirks are implemented once, here, and don't need to be
reimplemented from the manifest.

## agentsummons info

To find out what each harness supports, use the `info` command:

```sh
agentsummons info [--harness codex] [--json]
```

Info shows syntax and options for each harness, including
per-harness quirk notes, such as TLS trust-store behavior,
headless tool registration, stdout reliability.

## agentsummons doctor

To check whether your currently installed harness has been validated with
the installed version of agentsummons, use the `doctor` command:

```sh
agentsummons doctor [--json]
```

This command is free to run: only version commands execute.

Because harness flags may change between versions, we consider a harness
version that hasn't been tested with agentsummons a "drift candidate." This
means that the version of the harness that you have installed may have
different flags or invocation behavior. It may still work, but behavior
and flags haven't been validated, so you should smoke test before
performing a token-heavy task.

`doctor` exits 1 on a drift candidate and 2 on a failed version probe.

## Exit codes

`run` propagates the harness exit code. agentsummons's own failures use:

| Code | Meaning |
| --- | --- |
| 64 | Bad request / unsupported capability |
| 69 | Harness not installed |
| 75 | Timeout |

## JSON contract

All `--json` output carries `"schema_version"` and is a cross-language
contract. Shapes only change with a version review.
