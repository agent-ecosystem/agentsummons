---
title: Quickstart
description: Install agentsummons and invoke your first harness.
icon: rocket_launch
weight: 100
---

## Install

macOS users can `brew` install:

```sh
brew install agent-ecosystem/tap/agentsummons
```

For other install options, including go install, npm, PyPI, and
prebuilt binaries, refer to [Installation](/docs/installation/).

## Invoke a harness

To start an agent session with your preferred harness, use the
`agentsummons run` command:

```sh
agentsummons run --harness claude-code -p "Summarize this repo" \
  --allowed-tools Read,Grep,Glob --auto-approve
```

By default, agentsummons acts as a transparent wrapper.
Harness stdout/stderr pass through and the harness exit code propagates.

## Get a machine-readable result

To get output in a machine-readable format, use the `--json` flag:

```sh
agentsummons run --harness codex -p "hi" --model gpt-5.6-terra --json
```

`--json` swaps to a single result envelope on stdout with the harness
output embedded. Use this format for scripts and containers. The
envelope stamps the installed harness version and a `drift_hint` when
that version is newer than the flag surface this build was validated
against, letting you know to check the output as invocation behavior
may have changed.

## Check your harnesses

To check installed harness versions against agentsummons before running
a session, use the `doctor` command:

```sh
agentsummons doctor
```

`doctor` compares installed harness versions against the validated flag
surface without running anything except version commands. It's a cheap
pre-experiment gate. For more details, refer to
[Flag drift](/docs/flag-drift/).
