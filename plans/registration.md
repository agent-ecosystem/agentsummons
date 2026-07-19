# agentsummons: name registration plan

Goal: claim the `agentsummons` name across the namespaces that matter for a
Go-core library with future npm/PyPI wrapper packages. Companion to the
agentminutes plan of the same name; agentsummons started later, so several
rows were already settled by the initial release.

Availability was verified 2026-07-19 (commands at the bottom). Availability
can change at any time; re-verify each namespace just before claiming.

## Status

| Namespace | Status | Priority |
| --- | --- | --- |
| GitHub repo | Done: `github.com/agent-ecosystem/agentsummons`, v0.1.0 released | Done |
| Homebrew | Done: `agent-ecosystem/tap/agentsummons` formula ships with each release | Done |
| agentsummons.dev | Registered | Done |
| npm | Done: 0.0.1 stub published 2026-07-19; real wrapper in `wrappers/npm/` publishes per release | Done |
| PyPI | Done: 0.0.1 stub published 2026-07-19; real wrapper in `wrappers/pypi/` publishes per release | Done |
| GitHub org `agentsummons` | Free as of 2026-07-19; optional defensive claim | Medium |
| crates.io / RubyGems | Skipping (see notes) | Low |

## 1. GitHub

The module path `github.com/agent-ecosystem/agentsummons` is live and baked
into `go.mod`, every importer, and the v0.1.0 release; treat it as
permanent. pkg.go.dev indexes automatically on first fetch after a tagged
release; no registration needed.

The bare `agentsummons` org/user name was unclaimed as of 2026-07-19. Same
recommendation as agentminutes: claiming it is free and blocks the most
confusable namespace. Optional, not urgent.

## 2. npm

Claimed 2026-07-19 with an honest 0.0.1 stub (npm's dispute policy treats
empty placeholders as squatting, so it had a real export and README). The
stub is superseded: `wrappers/npm/` is now the real package — esbuild-style
platform packages (`agentsummons-<platform>-<arch>`) as optionalDependencies
of the main package, published automatically by the release workflow at the
Go release's version. The six platform-package names are claimed by their
first CI publish.

One-time setup: an `NPM_TOKEN` secret (granular automation token) on the
GitHub repo. Publishes use `--provenance`.

## 3. PyPI

Claimed 2026-07-19 with an honest 0.0.1 stub (PEP 541 lets PyPI reclaim
squatted names). Superseded the same way: `wrappers/pypi/` builds
per-platform wheels bundling the binary, published automatically by the
release workflow via PyPI trusted publishing.

One-time setup: on pypi.org, add a trusted publisher to the existing
`agentsummons` project — repository `agent-ecosystem/agentsummons`,
workflow `release.yml`, no environment. No token secret needed.

Notes:
- PyPI normalizes names, so `agentsummons` also blocks `agent-summons` and
  `agent_summons`. No need to register variants.

## 4. Homebrew

Already done, no reservation was needed: goreleaser pushes
`Formula/agentsummons.rb` to `agent-ecosystem/homebrew-tap` on every
release (`.goreleaser.yaml` `brews` section). homebrew-core requires
notability and can come later if ever warranted.

## 5. Skipped registries

- **crates.io:** no Rust component planned, and crates.io norms frown on
  defensive squatting. Skip unless a Rust wrapper becomes real.
- **RubyGems:** no Ruby story. Skip.

## 6. Domain notes

- agentsummons.dev registered 2026-07-16 (Tucows/Hover, alongside
  agentminutes.dev). `.dev` is HSTS-preloaded: the site must serve HTTPS
  from day one.
- agentsummons.com is held by an unrelated third party (NameCheap,
  registered 2026-03-23, parked). Not a conflict; nothing to do unless it
  lapses and the project has traction.
- agentsummons.io was free as of 2026-07-19. Optional; probably skip.

## Re-verification commands

Run each just before registering (404 or zero means still free):

```bash
curl -s -o /dev/null -w '%{http_code}\n' https://registry.npmjs.org/agentsummons
curl -s -o /dev/null -w '%{http_code}\n' https://pypi.org/pypi/agentsummons/json
curl -s -o /dev/null -w '%{http_code}\n' https://api.github.com/users/agentsummons
gh api 'search/repositories?q=agentsummons+in:name' --jq '.total_count'
curl -sL -o /dev/null -w '%{http_code}\n' https://rdap.org/domain/agentsummons.io
```

## Known name-adjacency (no action needed)

- GitHub repo search for `agentsummons` returns exactly one hit: this
  project. No npm/PyPI/product collisions surfaced during verification.
- "Summons" has legal-tech connotations (court summonses); no product named
  "agent summons" was found in that space at decision time.
