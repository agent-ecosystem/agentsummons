# agentsummons

Go library + CLI for invoking agent harnesses (Antigravity CLI, Claude
Code, Codex CLI) in headless mode.

Every headless-agent experiment rediscovers the same lore: which binary,
which permission-bypass flag, that `agy` ignores your working directory
without `--add-dir` and swallows everything after `-p`, that Codex wants a
bare positional prompt under an `exec` subcommand. agentsummons owns that
knowledge once, behind one Go API and one CLI, so your scripts and
experiments don't have to.

It is the companion to
[agentminutes](https://github.com/agent-ecosystem/agentminutes), which
parses the session transcripts harnesses write: **agentsummons convenes
the meeting, agentminutes takes the minutes.**

## Install

```bash
brew install agent-ecosystem/tap/agentsummons
# or
go install github.com/agent-ecosystem/agentsummons/cmd/agentsummons@latest
```

Also on [npm](https://www.npmjs.com/package/agentsummons) and
[PyPI](https://pypi.org/project/agentsummons/) for Node and Python
projects, with prebuilt static binaries on the
[releases page](https://github.com/agent-ecosystem/agentsummons/releases).
See [Installation](https://agentsummons.dev/docs/installation/) for all
options.

## Quick start

```bash
# Invoke a harness. Default is a transparent wrapper: harness
# stdout/stderr pass through and the harness exit code propagates.
agentsummons run --harness claude-code -p "Summarize this repo" \
  --allowed-tools Read,Grep,Glob --auto-approve

# --json swaps to a single result envelope on stdout, the shape
# scripts and containers consume.
agentsummons run --harness codex -p "hi" --json

# Compare installed harness versions against the validated flag
# surface before an experiment burns tokens. Free; only version
# commands run.
agentsummons doctor
```

As a Go library:

```go
res, err := agentsummons.Run(ctx, agentsummons.Request{
    Harness:     agentsummons.ClaudeCode,
    Prompt:      "Summarize this repo",
    Workdir:     dir,
    AutoApprove: true,
})
```

## Capability matrix

One request shape, translated per harness:

| Harness | Binary | Model | Session ID | Resume | Allowed tools | JSON output | Auto-approve |
|---|---|---|---|---|---|---|---|
| antigravity | `agy` | `--model "<display name>"` | (none) | `--conversation <id>` | (none) | `--output-format json` (envelope) | `--dangerously-skip-permissions` |
| claude-code | `claude` | `--model <id>` | `--session-id <uuid>` | `--resume <session-id>` | `--allowedTools a,b` | `--output-format json` (envelope) | `--dangerously-skip-permissions` |
| codex | `codex` | `-m <id>` | (none) | `exec resume <session-id>` | (none) | `--json` (JSONL events) | `--dangerously-bypass-approvals-and-sandbox` |

`agentsummons info` has the full manifests, including per-harness quirk
notes. Setting a field a harness can't express is a loud
`*UnsupportedError`, never a silent drop, and agentsummons never applies
policy by default: permission bypass is always the caller's explicit
choice.

## Documentation

Full documentation is available at
**[agentsummons.dev](https://agentsummons.dev)**:

- [Quickstart](https://agentsummons.dev/docs/quickstart/): install
  agentsummons and invoke your first harness
- [CLI](https://agentsummons.dev/docs/cli/): the run, build, info, and
  doctor commands, exit codes, and the JSON contract
- [Go Library](https://agentsummons.dev/docs/library/): the Run and Build
  API, error semantics, and escape hatches
- [Harness Capabilities](https://agentsummons.dev/docs/harnesses/): the
  capability matrix and per-harness quirk notes
- [Multi-Turn Sessions](https://agentsummons.dev/docs/multi-turn/):
  resuming conversations across invocations, with worked examples
- [Example: An Experiment Runner](https://agentsummons.dev/docs/example-runner/):
  agentsummons driving a real multi-harness benchmark
- [Containers](https://agentsummons.dev/docs/containers/): baking
  reproducible eval images
- [Flag Drift](https://agentsummons.dev/docs/flag-drift/): the validated
  flag surface and what a drift candidate means

## License

MIT.
