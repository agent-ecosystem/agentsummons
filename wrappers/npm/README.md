# agentsummons

Node wrapper around
[agentsummons](https://github.com/agent-ecosystem/agentsummons), a Go
library + CLI that invokes agent harnesses (Antigravity CLI, Claude Code,
Codex CLI) headlessly: it owns the per-harness flag knowledge so scripts
and experiments don't have to rediscover it.

Installing this package delivers the real Go binary for your platform via
an optional dependency (no install scripts), plus a thin JS API.

## CLI

```bash
npx agentsummons run --harness claude-code -p "Summarize this repo" \
  --allowed-tools Read,Grep,Glob --auto-approve
```

Identical to the Go CLI — see the
[project README](https://github.com/agent-ecosystem/agentsummons#cli) for
the full surface (`run`, `build`, `info`, `doctor`).

## API

```js
const agentsummons = require("agentsummons");

const res = await agentsummons.run({
  harness: "claude-code",
  prompt: "Summarize this repo",
  workdir: dir,
  allowedTools: ["Read", "Grep", "Glob"],
  autoApprove: true,
  timeoutMs: 300_000,
});
// res is the CLI's `run --json` envelope, snake_case keys and all:
// res.exit_code, res.stdout, res.timed_out, res.session_id, ...
```

A nonzero harness exit is data (`res.exit_code`), not a rejection; a
timeout resolves with `timed_out: true` and the partial output. Bad
requests reject with `RequestError`, a missing harness binary with
`NotInstalledError`. `build(options)` returns the assembled argv without
executing; `binaryPath()` exposes the bundled binary for anything else.

Set `AGENTSUMMONS_BINARY` to override which binary the wrapper invokes.
