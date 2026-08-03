# Changelog

Notable changes to agentsummons. Each version covers the Go module, the
CLI, and the npm/PyPI wrappers together (wrapper versions always match
the Go tag). Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.3.1] - 2026-08-02

### Changed

- Revalidated the flag surface against antigravity 1.1.10
  (`LastValidated`). No flags moved; the live resume loop re-confirmed
  the stable envelope `conversation_id`.
- New antigravity manifest note: since agy 1.1.9, print mode expands
  slash commands and skills in the prompt, so a prompt starting with `/`
  may resolve to a command instead of being sent verbatim;
  `--disable-slash-commands` (via `ExtraArgs`) opts out.

### Fixed

- The release workflow's `--release-notes` file was ignored because the
  goreleaser config also disabled changelog generation, which publishes
  an empty release body; the disable is gone (the flag alone suppresses
  the commit-list changelog). The v0.3.0 release body was patched by
  hand.

## [0.3.0] - 2026-07-30

### Added

- Antigravity gained the JSON output capability: agy 1.1.8 added
  `--output-format json` for print mode, and its result envelope carries
  `conversation_id` in-band (the ask in upstream request
  [antigravity-cli#7](https://github.com/google-antigravity/antigravity-cli/issues/7)).
  `Request.JSONOutput` now assembles `--output-format json` for
  antigravity instead of failing with `*UnsupportedError`.
- Antigravity multi-turn is fully in-band when JSON output is on: take
  `conversation_id` from turn 1's envelope and pass it as `Resume`.
  Post-hoc discovery via agentminutes is only needed in text mode or on
  agy releases older than 1.1.8. A new live test
  (`TestLiveAntigravityResume`) covers the loop, so live coverage now
  spans all three harnesses' resume paths.
- GitHub release notes now come from CHANGELOG.md: the release workflow
  extracts the tag's section and fails the release if it is missing.

### Changed

- Revalidated the flag surface against antigravity 1.1.8, claude-code
  2.1.212, and codex 0.146.0 (`LastValidated`). No flags moved.
- Antigravity manifest notes reworked for 1.1.8: stdout-reliability and
  ref-discovery caveats now scoped to text mode, and resume append
  semantics documented (same conversation, not a fork; the warm-up
  double-conversation trap still applies to post-hoc discovery).

## [0.2.2] - 2026-07-19

### Fixed

- Publish the main npm package by explicit path, so the release workflow
  publishes it alongside the platform packages.

## [0.2.1] - 2026-07-19

### Fixed

- Made the npm publish step idempotent so a re-run release does not fail
  on already-published packages.
- Temporarily dropped win32-x64 from the npm platform matrix: npm
  refuses to create the `agentsummons-win32-x64` package name (support
  ticket open; the matrix entry returns when it resolves).

## [0.2.0] - 2026-07-19

### Added

- npm and PyPI wrapper packages (`wrappers/`): they deliver the platform
  binary and a thin `run`/`build` API over the `--json` envelopes,
  published by the release workflow with versions matching the Go tag.

### Changed

- Revalidated the flag surface against antigravity 1.1.4 and codex
  0.144.6.

## [0.1.0] - 2026-07-19

### Added

- Initial release: Go library (`Build`, `Run`, `Version`, `Harnesses`,
  `InfoFor`) and cobra CLI (`build`, `doctor`, `info`, `run`) for
  invoking Antigravity CLI, Claude Code, and Codex CLI headlessly, with
  typed capability manifests, loud `*UnsupportedError` rejection, and
  `--json` envelopes carrying `schema_version`.

[Unreleased]: https://github.com/agent-ecosystem/agentsummons/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/agent-ecosystem/agentsummons/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/agent-ecosystem/agentsummons/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/agent-ecosystem/agentsummons/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/agent-ecosystem/agentsummons/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/agent-ecosystem/agentsummons/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/agent-ecosystem/agentsummons/releases/tag/v0.1.0
