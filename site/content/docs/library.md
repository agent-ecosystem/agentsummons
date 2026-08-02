---
title: Go Library
description: The Run and Build API, error semantics, and escape hatches.
icon: code
weight: 400
---

## Run

To use agentsummons as a library in your project, use the `Run` API:

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

## Error semantics

A nonzero harness exit (`res.ExitCode`) isn't necessarily an error;
it gives you information about the underlying harness operations. On
timeout, the partial `Result` is returned alongside the context error,
giving experiments the ability to archive failures.

Setting a field the harness can't express returns an explicit
`*UnsupportedError` .

## Build

`Build` returns the assembled command as pure data without executing.

## Escape hatches

To pass anything that isn't explicitly modeled, use `ExtraArgs` and
`ExtraEnv`. You might use this if you want the harnesses'
continue-most-recent forms that agentsummons deliberately
doesn't model.

## No policy by default

agentsummons never applies policy on its own. Models, tool restrictions,
sandbox levels, and permission bypass are always the caller's explicit
choice.
