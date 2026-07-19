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
