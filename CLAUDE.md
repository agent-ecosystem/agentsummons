# agentsummons

Go library + CLI for invoking agent harnesses (Antigravity CLI, Claude
Code, Codex CLI) headlessly. Companion to agentminutes: agentsummons
convenes the meeting, agentminutes takes the minutes. `plans/design.md`
holds the design rationale; `DEVELOPMENT.md` the flag-knowledge
maintenance loop.

## Commands

```bash
go test ./... -count=1        # hermetic: golden argv, fake-harness process, and CLI contract tests
golangci-lint run             # lint + gofumpt (CI-enforced)
GOOS=windows go build ./...   # CI also tests on windows-latest

# Live validation against real installed harnesses (spends tokens):
AGENTSUMMONS_LIVE=all go test -run TestLive -v
```

## Non-negotiables

- agentsummons never learns about transcripts: no transcript paths, no
  session discovery, no output parsing beyond capturing bytes. That
  knowledge lives in agentminutes, which imports this module — never the
  reverse. The core library stays stdlib-only (cobra is cmd/-only).
- Setting a `Request` field a harness cannot express is a loud
  `*UnsupportedError`, never a silent drop; a Request field is accepted
  exactly when its `Capabilities` field is non-empty (test-enforced).
- No policy by default: permission bypass, models, sandbox levels, and TLS
  env workarounds are caller decisions; `ExtraArgs`/`ExtraEnv` must always
  suffice as escape hatches. Manifest notes inform, never auto-apply.
- The prompt is always the final Argv element (test-enforced); `ExtraArgs`
  splice in before it.
- Harness lists stay alphabetical (constants, spec table, flag help, docs).
- `harnesses.go` is the single adjustment point for flag knowledge; it
  moves together with `LastValidated` in `versions.go` (test-enforced),
  and both are revalidated via the live tests per `DEVELOPMENT.md`.
- JSON output shapes (`run --json`, `build`, `info`, `doctor`) are a
  cross-language contract; changes need a `schema_version` review.

## Layout

- Root package: `agentsummons.go` facade (`Build`, `Run`, `Version`,
  `Harnesses`, `InfoFor`); `harnesses.go` spec table (THE knowledge);
  `capabilities.go` manifest types; `versions.go` flag-surface
  `LastValidated` + `VersionNewer` (copied from agentminutes, not
  imported); `request.go` `Request`/`Built`/`Result`; `errors.go` typed
  errors.
- `cmd/agentsummons/` cobra CLI: `build`, `doctor`, `info`, `run`.
- `internal/fakeharness/` test-only scaffolding: the fake-harness pattern
  shared by the library and CLI test suites. It imports `testing` but is
  only ever imported from `_test.go` files; it is not part of the library's
  API surface.
- `plans/` design/status docs that are not user documentation.
- `wrappers/` npm and PyPI packages that deliver the Go binary plus a thin
  `run`/`build` API over the `--json` envelopes. They are schema_version
  consumers: any envelope or flag-surface change must be mirrored here (both
  have fake-binary test suites, run in CI). Platform matrices in
  `wrappers/npm/scripts/build-packages.mjs` and `wrappers/pypi/build_wheels.py`
  move together with `.goreleaser.yaml`'s build matrix. Published by the
  release workflow; versions always match the Go tag.

## Docs site (site/)

Hugo + Lotus Docs site for agentsummons.dev, instantiated from
af-site-scaffold's template. Deploy with `site/build_and_sync` (rsync to
Dreamhost); llms.txt and per-page markdown are Hugo output formats,
regenerated on every build. Verify locally with
`hugo server -p 1717` + `afdocs check http://localhost:1717` (never port
1719/1720: Node's fetch blocks WHATWG bad-list ports and every check
reports "fetch failed"; `content-negotiation` passes only on the live
Apache site). The repo pre-commit hook runs `site/check_prose_style`
(Vale, DC style: em dashes and "not X, but Y" constructions are errors).
README.md is outside the hook's scope; lint it with
`vale --config site/.vale.ini README.md`.

Docs that move with the code — update when `harnesses.go`/`LastValidated`
move (the capability matrix now lives in THREE places: `harnesses.go`,
README, and the site):

- `content/docs/harnesses.md`: capability matrix mirrors the spec table.
- `content/docs/flag-drift.md`: doctor example bakes in real
  `LastValidated` values (currently agy 1.1.10 / claude-code 2.1.212 /
  codex 0.146.0) plus a fabricated newer codex as the drift candidate;
  the `drift_hint` string mirrors `driftHint()` in `cmd/.../run.go`.
- `content/docs/cli.md`: the `run --json` envelope example
  (claude-code 2.1.205) is deliberately older than validated so no
  `drift_hint` appears; keep that invariant when versions move. The
  argv in it mirrors the claude-code assemble function.
- `content/docs/multi-turn.md`: the agy 1.1.8 `conversation_id` claim
  mirrors the antigravity notes in `harnesses.go`.
- `assets/images/terminal-hero.svg` (hand-edited) and
  `sharing-image.html` both bake in doctor output with versions; after
  editing sharing-image.html, run `site/generate_sharing_image`
  (headless Chrome) to re-render `static/sharing.png`.
- `data/landing.yaml`: hero badge pins the release version.

Conventions: CLI examples show captured real output; when shapes change,
re-run the command rather than hand-editing numbers. Landing copy in
`data/landing.yaml` is mirrored into the homepage markdown template for
content parity; never use `--` in landing prose (the typographer's en
dash breaks parity matching) and keep unused theme landing sections
(imageCompare) explicitly disabled. `example-runner.md` quotes
agentminutes' Locator API and a skillxp-derived runTurn; revisit if
either API moves.
