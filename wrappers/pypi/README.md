# agentsummons

Python wrapper around
[agentsummons](https://github.com/agent-ecosystem/agentsummons), a Go
library + CLI that invokes agent harnesses (Antigravity CLI, Claude Code,
Codex CLI) headlessly: it owns the per-harness flag knowledge so scripts
and experiments don't have to rediscover it.

The wheel bundles the real Go binary for your platform, plus a thin Python
API.

## CLI

```bash
pip install agentsummons
agentsummons run --harness claude-code -p "Summarize this repo" \
  --allowed-tools Read,Grep,Glob --auto-approve
```

Identical to the Go CLI — see the
[project README](https://github.com/agent-ecosystem/agentsummons#cli) for
the full surface (`run`, `build`, `info`, `doctor`).

## API

```python
import agentsummons

res = agentsummons.run(
    harness="claude-code",
    prompt="Summarize this repo",
    workdir=dir,
    allowed_tools=["Read", "Grep", "Glob"],
    auto_approve=True,
    timeout=300,
)
# res is the CLI's `run --json` envelope as a dict:
# res["exit_code"], res["stdout"], res["timed_out"], res["session_id"], ...
```

A nonzero harness exit is data (`res["exit_code"]`), not an exception; a
timeout returns `timed_out: True` with the partial output. Bad requests
raise `RequestError`, a missing harness binary raises `NotInstalledError`.
`build(...)` returns the assembled argv without executing;
`binary_path()` exposes the bundled binary for anything else.

Set `AGENTSUMMONS_BINARY` to override which binary the wrapper invokes.
