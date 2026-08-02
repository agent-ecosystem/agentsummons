---
title: Installation
description: Homebrew, go install, npm, PyPI, and prebuilt binaries.
icon: download
weight: 200
---

## Homebrew

```sh
brew install agent-ecosystem/tap/agentsummons
```

## Go

As a CLI:

```sh
go install github.com/agent-ecosystem/agentsummons/cmd/agentsummons@latest
```

As a library:

```sh
go get github.com/agent-ecosystem/agentsummons
```

## npm and PyPI

For Node and Python projects, the
[npm](https://www.npmjs.com/package/agentsummons) and
[PyPI](https://pypi.org/project/agentsummons/) packages bundle the platform
binary and expose `run`/`build` over the JSON envelope contract:

```sh
npm install agentsummons
```

```sh
pip install agentsummons
```

Wrapper versions always match the Go tag.

## Prebuilt binaries

Static binaries for Darwin, Linux, and Windows are on the
[releases page](https://github.com/agent-ecosystem/agentsummons/releases).
They are built with `CGO_ENABLED=0`, so a single `COPY` works in container
builds.
