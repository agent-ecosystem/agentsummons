---
title: Flag Drift
description: The validated flag surface and the doctor command.
icon: monitor_heart
weight: 800
---

When harness versions change, flags and invocation behavior may change. We
refer to this as "drift" in agentsummons.

`LastValidated` (in `versions.go`) records the newest release each harness's
spec was validated against. The `agentsummons doctor` command compares the
harness version you're running against what has been validated to work correctly
with agentsummons.

Newer installed releases may still work. Removed flags fail with an error, but
agentsummons may not expose additive flags until a future release.
`doctor` and `run --json`'s `drift_hint` tell you when you're past validated
coverage.

## Example: what doctor reports

This is a `doctor` run on a machine where two harnesses match their
validated versions and codex has moved ahead:

```
$ agentsummons doctor
antigravity  installed 1.1.8, validated 1.1.8 — clean
claude-code  installed 2.1.212, validated 2.1.212 — clean
codex        installed 0.148.2 > validated 0.146.0 — drift candidate; check for an agentsummons update
```

This run exits 1 because of the drift candidate, which is what makes
`doctor` usable as a gate: a CI step or a `RUN agentsummons doctor` line
in a Dockerfile fails before spending any tokens (see
[Containers](/docs/containers/)). Harnesses that aren't installed report
`not installed` without affecting the exit code.

## Example: the drift hint at run time

You don't have to run `doctor` to find out. When a run's harness is past
validated coverage, the `run --json` envelope carries the same
information in its `drift_hint` field:

```json
"harness": "codex",
"harness_version": "0.148.2",
"drift_hint": "installed 0.148.2 is newer than flag-surface validated 0.146.0; check for an agentsummons update",
```

What drift means in practice:

- **The run still executes.** A drift candidate is a statement about
  validation coverage. agentsummons doesn't refuse a newer harness; it
  tells you its spec table hasn't been re-checked against that release.
- **The failure mode is loud, and the gap is additive.** If the newer
  harness removed a flag, the run fails visibly with the harness's own
  error. The gap is in new capabilities; flags added in the newer release
  stay unexposed until agentsummons revalidates.
- **The cheap response is a smoke test.** Run a short, low-stakes prompt
  on the drifted harness before committing a token-heavy experiment to
  it, and check whether a newer agentsummons release has already
  revalidated.

See `DEVELOPMENT.md` in the repo for the revalidation loop.
