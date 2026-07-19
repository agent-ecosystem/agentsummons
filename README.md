# agentsummons

Go library + CLI for invoking agent harnesses — Antigravity CLI, Claude
Code, Codex CLI — in headless mode.

Every headless-agent experiment rediscovers the same lore: which binary,
which permission-bypass flag, that `agy` ignores your working directory
without `--add-dir` and swallows everything after `-p`, that Codex wants a
bare positional prompt under an `exec` subcommand. agentsummons owns that
knowledge once, behind one Go API and one CLI, so your scripts and
experiments don't have to.

It is the companion to
[agentminutes](https://github.com/agent-ecosystem/agentminutes), which
parses the session transcripts harnesses write: **agentsummons convenes the
meeting, agentminutes takes the minutes.** agentsummons deliberately knows
nothing about transcripts — its contract ends at "the command ran; here is
everything observable about that" — and a `Result`'s timestamps and workdir
are exactly the inputs `agentminutes sessions` needs to find what was
written.

## Install

```bash
brew install agent-ecosystem/tap/agentsummons
# or
go install github.com/agent-ecosystem/agentsummons/cmd/agentsummons@latest
# or, for Node and Python projects (real binary + thin run/build API):
npm install agentsummons
pip install agentsummons
```

Prebuilt static binaries (darwin/linux/windows) are on the
[releases page](https://github.com/agent-ecosystem/agentsummons/releases).
As a Go library: `go get github.com/agent-ecosystem/agentsummons`. The
[npm](https://www.npmjs.com/package/agentsummons) and
[PyPI](https://pypi.org/project/agentsummons/) packages bundle the
platform binary and expose `run`/`build` over the JSON envelope contract;
see `wrappers/`.

## CLI

```bash
# Invoke a harness. Default is a transparent wrapper: harness stdout/stderr
# pass through and the harness exit code propagates.
agentsummons run --harness claude-code -p "Summarize this repo" \
  --allowed-tools Read,Grep,Glob --auto-approve

# --json swaps to a single result envelope on stdout (harness output
# embedded) — the shape scripts and containers consume. It stamps the
# installed harness version and a drift_hint when that version is newer
# than the flag surface this build was validated against.
agentsummons run --harness codex -p "hi" --model gpt-5.6-terra --json

# Print the exact argv/dir/env run would use, without executing. This is
# the correct-assembly path for callers in other languages: ordering
# quirks are implemented once, here, not reimplemented from the manifest.
agentsummons build --harness antigravity -p "hi" --auto-approve

# What does each harness support, via which syntax?
agentsummons info [--harness codex] [--json]

# Compare installed versions against the validated flag surface. Free —
# only version commands run. The cheap pre-experiment gate.
agentsummons doctor [--json]
```

Exit codes: `run` propagates the harness exit code; agentsummons's own
failures use 64 (bad request / unsupported capability), 69 (harness not
installed), 75 (timeout). `doctor` exits 1 on a drift candidate, 2 on a
failed version probe.

All `--json` output carries `"schema_version"` and is a cross-language
contract: shapes only change with a version review.

## Library

```go
res, err := agentsummons.Run(ctx, agentsummons.Request{
    Harness:      agentsummons.ClaudeCode,
    Prompt:       "Summarize this repo",
    Workdir:      dir,
    SessionID:    id, // preset identity; claude-code only
    AllowedTools: []string{"Read", "Grep", "Glob"},
    JSONOutput:   true,
    AutoApprove:  true,
})
```

A nonzero harness exit is data (`res.ExitCode`), not an error. On timeout
the partial `Result` is returned alongside the context error, so
experiments can archive failures. Setting a field the harness can't
express is a loud `*UnsupportedError` — never a silent drop. `Build`
returns the assembled command as pure data without executing;
`ExtraArgs`/`ExtraEnv` are the escape hatches for anything unmodeled.
agentsummons never applies policy by default: models, tool restrictions,
sandbox levels, and permission bypass are always the caller's explicit
choice.

## Capability matrix

| Harness | Binary | Model | Session ID | Resume | Allowed tools | JSON output | Auto-approve |
|---|---|---|---|---|---|---|---|
| antigravity | `agy` | `--model "<display name>"` | — | `--conversation <id>` | — | — | `--dangerously-skip-permissions` |
| claude-code | `claude` | `--model <id>` | `--session-id <uuid>` | `--resume <session-id>` | `--allowedTools a,b` | `--output-format json` (envelope) | `--dangerously-skip-permissions` |
| codex | `codex` | `-m <id>` | — | `exec resume <session-id>` | — | `--json` (JSONL events) | `--dangerously-bypass-approvals-and-sandbox` |

`agentsummons info` has the full manifests, including the per-harness quirk
notes (TLS trust-store behavior, headless tool registration, stdout
reliability).

## Multi-turn

Harness session state lives on disk in each harness's own store, so a
conversation is a sequence of discrete invocations naming the same session:

1. Turn 1: `run` (claude-code: preset `--session-id` to know the ref up
   front).
2. Get the session ref — in-band from stdout (claude-code envelope, codex
   event stream), or post-hoc with `agentminutes sessions` (antigravity,
   whose headless mode never emits the ref — there's an
   [open upstream request](https://github.com/google-antigravity/antigravity-cli/issues/7)
   to change that; if it lands, this caveat relaxes).
3. Turn N: `run --resume <ref>`.

An explicit ref is required; the harnesses' continue-most-recent forms are
deliberately not modeled (racy against any concurrent agent session), but
remain reachable via `--extra-arg`.

## Containers

For reproducible eval environments, bake the harness binaries and the
`agentsummons` binary into the image and make `agentsummons run --json` the
container command: stdout carries exactly one envelope, and transcript
roots volume-mount out for agentminutes to parse on the host. A
`RUN agentsummons doctor` step fails the image build when installed harness
versions drift past validated. Releases are static binaries
(`CGO_ENABLED=0`) so this is a single `COPY`.

## Validated flag surface

Harness flags drift. `LastValidated` (in `versions.go`) records the newest
release each harness's spec was validated against, and `agentsummons
doctor` compares it against what's installed. Newer installed releases
usually still work (removed flags fail loudly); `doctor` and `run --json`'s
`drift_hint` tell you when you're past validated coverage. See
`DEVELOPMENT.md` for the revalidation loop.

## License

MIT.
