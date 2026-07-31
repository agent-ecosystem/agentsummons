# Development

## Commands

```bash
go test ./... -count=1        # hermetic tests: golden argv, fake-harness process, and CLI contract tests
golangci-lint run             # lint + gofumpt (CI-enforced)
GOOS=windows go build ./...   # CI also tests on windows-latest
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./...  # release artifacts are static

# Live validation against the real installed harnesses (SPENDS TOKENS —
# a couple of trivial prompts per enabled harness):
AGENTSUMMONS_LIVE=all go test -run TestLive -v
AGENTSUMMONS_LIVE=claude-code go test -run TestLive -v   # or a comma list
```

## What this repo is (and is not)

agentsummons owns headless *invocation* knowledge: binaries, flags,
assembly ordering, capability syntax. It must never learn about
transcripts — no roots, no discovery, no parsing. That knowledge lives in
[agentminutes](https://github.com/agent-ecosystem/agentminutes), which
imports this module; an import in the other direction would create a
cycle. See `plans/design.md` for the full rationale and `CLAUDE.md` for
the non-negotiables.

## Maintaining the flag knowledge

The spec table in `harnesses.go` and the `LastValidated` map in
`versions.go` move together: the table encodes what was checked, the map
records against which release. When a harness releases a new version
(`agentsummons doctor` reports a drift candidate):

1. Review the release's `--help` output (and changelog, if any) for every
   flag the spec table uses, including the resume forms.
2. Run the live tests against the new release:
   `AGENTSUMMONS_LIVE=<id> go test -run TestLive -v`. They exercise
   auto-approve, JSON output, session presetting, and all three resume
   loops end-to-end. Antigravity's resume test needs agy 1.1.8 or newer,
   where `--output-format json` landed and turn 1's envelope carries the
   conversation ID in-band. On older releases the ref only exists in the
   transcript store, and depending on agentminutes for discovery (even
   test-only) would create the module cycle the boundary forbids, so
   validate those manually by composing the two tools: `agentsummons run`
   turn 1, `agentminutes sessions --harness antigravity --since <start>
   --until <end>` for the conversation ID, then `agentsummons run
   --resume <id>` and confirm recall.
3. Fix the spec table if anything moved, update the manifest notes if a
   quirk appeared or disappeared, and bump the harness's `LastValidated`
   entry in the same change. Record anything user-visible (a capability
   gained or lost, a caveat that relaxed) in `CHANGELOG.md` under
   Unreleased.

This loop is also the pre-release gate, and ordering matters across the
two repos: agentminutes depends on this module, so revalidate the flag
surface and bump `LastValidated` here first, promote the Unreleased
section of `CHANGELOG.md` to the new version heading, and tag the
agentsummons release; then move to agentminutes and run its drift-check
process (the `drift probe` maintainer verb) against the same harness
releases. The release workflow extracts the tag's changelog section for
the GitHub release notes and fails the release if the section is
missing, so the promote step cannot be skipped.

The hermetic suite enforces the invariants that don't need a binary:
prompt-is-always-last, the manifest and the validator can never disagree
about what's supported, spec/LastValidated coverage stays paired, and
harness lists stay alphabetical.

### Adding a capability

A new `Request` field fans out across a fixed set of hand-maintained
spots: `Request` (request.go), `Capabilities` (capabilities.go), the
`validate` switch and each harness's `assemble` (harnesses.go), the CLI
flag surface (`cmd/agentsummons/flags.go`), the text renderer in
`cmd/agentsummons/info.go`, the manifest-contract cases in
`capabilities_test.go` (`TestUnsupportedMatchesManifest`), the README
capability matrix, and a `CHANGELOG.md` entry. The hermetic suite catches
a validator/manifest mismatch, and `--json` output picks new
`Capabilities` fields up automatically — but the `info` text output and
the README matrix do not; check those two by hand.

The deeper canary lives in agentminutes: its `drift probe` invokes the
harnesses and parses the fresh transcripts, so one (paid) probe run
exercises both this repo's flag surface (invocation-time failures) and
agentminutes' format surface (parse-time failures).

## Testing notes

Process handling is tested hermetically via the fake-harness scaffolding
in `internal/fakeharness`, shared by the library tests (`run_test.go`) and
the CLI tests (`cmd/agentsummons/cli_test.go`): each package's `TestMain`
lets the test binary double as a harness, and tests copy it into a temp
`PATH` dir under the real binary name (`claude`, `codex`, ...), driving
behavior through `FAKE_*` env vars. This works unchanged on Windows (the
copy gets an `.exe` suffix). The package is test-only scaffolding; nothing
outside `_test.go` files may import it.

The CLI suite locks in the cross-language surface: golden envelope tests
for every `--json` shape, the exit-code mapping, and the flag-to-`Request`
translation.

The live tests are the only ones that execute real harnesses, and only
when `AGENTSUMMONS_LIVE` opts in. They never run in CI.

## JSON contract

The `--json` shapes of `run`, `build`, `info`, and `doctor` carry
`schema_version` and are consumed cross-language. Any shape change needs a
`schema_version` review: additive fields are fine within a version;
renames, removals, or semantic changes bump it. The golden envelope tests
in `cmd/agentsummons/cli_test.go` assert the actual wire keys, so a shape
change fails tests instead of shipping silently.
