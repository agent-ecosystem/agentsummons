---
title: Containers
description: Baking agentsummons into reproducible eval images.
icon: deployed_code
weight: 700
---

For reproducible eval environments:

1. Bake the harness binaries and the `agentsummons` binary into the image
2. Make `agentsummons run --json` the container command

Stdout carries exactly one envelope, and transcript roots volume-mount out for
[agentminutes](https://github.com/agent-ecosystem/agentminutes) to parse on
the host.

When installed harness versions drift past the most recent validated version,
`RUN agentsummons doctor` fails the image build.

Releases are static binaries (`CGO_ENABLED=0`), so installation in a
Dockerfile is a single `COPY`.

## Example: an eval image

This Dockerfile builds an image whose only job is to run one agent
invocation and emit one envelope:

```dockerfile
FROM node:22-slim

# Install the harnesses this image invokes. Add the others you use.
RUN npm install -g @anthropic-ai/claude-code

# agentsummons is a static binary: install is one COPY.
COPY agentsummons /usr/local/bin/agentsummons

# Fail the build if an installed harness has drifted past the flag
# surface this agentsummons build was validated against.
RUN agentsummons doctor

# Every container run is one invocation: one envelope on stdout.
ENTRYPOINT ["agentsummons", "run", "--json", "--harness", "claude-code"]
```

## Example: running the image

Each run mounts a project in, mounts the transcript root out, and
redirects stdout to capture the envelope:

```sh
docker run --rm \
  -e ANTHROPIC_API_KEY \
  -v "$PWD/project:/work" \
  -v "$PWD/transcripts:/root/.claude/projects" \
  my-eval-image \
  --workdir /work -p "Summarize this repo" --auto-approve \
  > result.json
```

The behaviors these examples encode:

- **The drift gate runs at build time, where it's free.** `doctor` only
  executes version commands, so the build needs no API keys and spends no
  tokens. A drift candidate fails the image build (exit 1) before any
  experiment runs on it. Harnesses that aren't installed in the image
  report `not installed` without failing the build, so a single-harness
  image passes cleanly.
- **The image pins the contract; the caller supplies the work.** The
  `ENTRYPOINT` fixes the harness and the envelope output. Everything
  per-run (prompt, workdir, permissions) arrives as arguments appended by
  `docker run`.
- **Collection is a redirect.** With `--json`, stdout is exactly one
  envelope, so `> result.json` is the entire result-gathering step. No
  log scraping.
- **Transcripts leave the container.** claude-code writes transcripts
  under `~/.claude/projects` (the container runs as root, so
  `/root/.claude/projects`). Mounting that to the host is what lets
  [agentminutes](https://github.com/agent-ecosystem/agentminutes) parse
  the session after the container exits.
- **Credentials enter at run time.** `-e ANTHROPIC_API_KEY` passes the
  key from the caller's environment. Nothing is baked into the image.
