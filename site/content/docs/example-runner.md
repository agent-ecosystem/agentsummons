---
title: "Example: An Experiment Runner"
description: A worked example of agentsummons driving a multi-harness benchmark.
icon: science
weight: 750
---

This page walks through a real use of agentsummons: the
[Agent Skill Implementation](https://agentskillimplementation.com)
benchmark runner, which asks the same questions of Antigravity CLI,
Claude Code, and Codex CLI and compares their answers. Every check in
that project runs the same probe skills on every harness, so the runner
needs identical invocation semantics across all three. agentsummons is the
layer that provides them.

## The shape of the problem

A benchmark turn looks simple: start a session in a prepared project
directory, send a prompt, wait for the result, then go find the transcript
the harness wrote. Doing that identically across three harnesses is where
the complexity lives:

- Each harness has a different binary, flag surface, and argument order.
- Multi-turn checks need a session ref from turn 1 to resume in turn 2,
  and each harness hands that ref back differently.
- Some turns need tool permissions, and each harness grants them
  differently.
- A hung harness has to be killed without losing the partial result,
  because a timeout is itself a finding worth archiving.

The runner encodes all of that once, in one function that works for every
harness.

## One turn, any harness

This is the heart of the runner, adapted from the real code:

```go
// runTurn invokes one turn of a session, on any harness.
// sessionID is empty on the opening turn; later turns pass the ref
// captured from turn 1.
func runTurn(ctx context.Context, harness agentsummons.ID,
    project, prompt, sessionID string) (*agentsummons.Result, error) {

    req := agentsummons.Request{
        Harness: harness,
        Prompt:  prompt,
        Workdir: project,
    }

    if sessionID == "" {
        // Opening turn. For claude-code, preset the session identity so
        // the transcript is addressable before the run even starts.
        if harness == agentsummons.ClaudeCode {
            req.SessionID = uuid.NewString()
        }
    } else {
        // Follow-up turn. Resume appends to the same session on all
        // three harnesses; agentsummons maps this to --resume,
        // --conversation, or the exec resume subcommand as needed.
        req.Resume = sessionID
    }

    // Grant only what the turn needs. The flag translation is
    // per-harness; the policy decision stays here, in the caller.
    switch harness {
    case agentsummons.ClaudeCode:
        req.AllowedTools = []string{"Skill"}
    case agentsummons.Antigravity, agentsummons.Codex:
        req.AutoApprove = true
    }

    // A hung harness is killed at the deadline, and the partial Result
    // comes back alongside the context error for archiving.
    runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    return agentsummons.Run(runCtx, req)
}
```

The behaviors this encodes, made explicit:

- **Harness selection is data.** The harness is a value
  (`agentsummons.ID`), so "run this check on all three harnesses" is a
  loop, and adding a harness to the benchmark is adding an element to a
  slice.
- **Identity is preset where possible.** Claude Code accepts a session ID
  up front, so the runner knows the transcript's address before invoking.
  The other harnesses hand back their ref in the first turn's output
  instead.
- **Resume is uniform.** The caller stores one string from turn 1 and
  sets `Resume` on every later turn. The three different resume syntaxes
  live in agentsummons, and the append-rather-than-fork behavior is
  documented in [Multi-Turn Sessions](/docs/multi-turn/).
- **Permission policy stays in the caller.** agentsummons never
  auto-approves on its own. The runner decides which turns get
  permissions; agentsummons translates that decision per harness.
- **Timeouts return data.** The check records exit code, timestamps, and
  partial output even when the harness had to be killed.

## After the run: finding the transcript

A `Result` carries the session ID, timestamps, and workdir, and those are
exactly what the companion tool
[agentminutes](https://github.com/agent-ecosystem/agentminutes) needs to
locate the transcript the harness wrote. When the ID was preset, as in the
Claude Code path above, lookup is direct:

```go
res, err := runTurn(ctx, agentsummons.ClaudeCode, project, prompt, "")
if err != nil {
    return err
}
loc, err := agentminutes.LocatorFor(harness.ClaudeCode)
if err != nil {
    return err
}
root, err := loc.DefaultRoot() // e.g. ~/.claude/projects
if err != nil {
    return err
}
// The transcript is addressable by the identity the run preset.
ref, err := loc.Locate(root, res.SessionID)
```

Harnesses without preset identity must be found instead of addressed. The
runner scans the transcript root with `loc.Scan` filtered to the run's
time window (`res.Start` to `res.End`, plus slack) and matches candidates
against `res.Workdir`. Either way, the fields agentsummons reports are the
fields the lookup needs.

This is the deliberate division of labor: agentsummons convenes the
meeting, agentminutes takes the minutes. The benchmark's findings then
cite the located transcript, which is what makes every claim on the
Agent Skill Implementation site traceable to a real session.

## When to reach for agentsummons

The pattern above generalizes past benchmarks. agentsummons fits whenever
you're invoking coding agents as infrastructure:

- **Experiments and evals** that need the same prompt run across
  harnesses, with results you can compare and archive.
- **CI jobs** that gate on an agent's exit code, where a
  `RUN agentsummons doctor` step catches harness drift before tokens are
  spent (see [Containers](/docs/containers/)).
- **Cross-language callers** that use `agentsummons build` to assemble
  correct argv without reimplementing per-harness quirks (see
  [CLI](/docs/cli/)).
- **Batch pipelines** that fan a work queue out to agent sessions and
  consume one JSON envelope per run.

If you're invoking one harness interactively, you don't need agentsummons.
The moment a script, runner, or second harness enters the picture, the
lore it owns starts paying for itself.
