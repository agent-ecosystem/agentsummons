# agentsummons design

Go library + CLI for invoking agent harnesses (Antigravity CLI, Claude Code,
Codex CLI) in headless mode. It owns the flag knowledge that every headless
experiment otherwise rediscovers: which binary, which flags, prompt-passing
mechanics, working-dir mechanics, and which optional capabilities (model
pinning, session IDs, tool allowlists, JSON output) each harness supports and
via what syntax.

Companion to [agentminutes](https://github.com/agent-ecosystem/agentminutes):
agentsummons convenes the meeting, agentminutes takes the minutes.

## Boundary (the one rule that matters)

**agentsummons knows nothing about transcripts.** Not where they live, not how
to find them, not how to parse them. Its contract ends at "the command ran;
here is everything observable about that": argv, env, workdir, timing, exit
code, captured output, and the session ID when the harness let us preset one.

This keeps the dependency graph acyclic: agentminutes imports agentsummons
(for its drift probe), never the reverse. Session *discovery* composes on the
consumer side — a `Result`'s start/end timestamps and workdir are exactly the
inputs `agentminutes sessions --since/--until --cwd` needs.

Corollaries:

- No transcript-root paths anywhere in this repo. Those belong to
  agentminutes `Locator.DefaultRoot`.
- No output parsing beyond capturing bytes. Interpreting stdout (Claude's
  JSON envelope, Codex's JSONL event stream, Antigravity's narration text) is
  the caller's business; the capability manifest documents which shape to
  expect, nothing more.
- No experiment policy. Model choice, tool allowlists, sandbox levels, and
  TLS workarounds are caller decisions; agentsummons provides the syntax and
  the escape hatches, and never applies policy by default.

## Non-goals

- Streaming/interactive sessions. Blocking run, results at the end. A
  streaming stdout writer is a compatible later addition (see Open
  questions), but v0.1 is batch.
- Retry logic, prompt templating, archiving, cost accounting. Callers own
  these; keeping them out is what makes the abstraction survive its second
  consumer.
- Unifying harness exit-code or stdout semantics. We report what happened;
  we do not normalize it.

## Module and layout

Module `github.com/agent-ecosystem/agentsummons`, MIT, same Go version and
lint config as agentminutes (golangci-lint + gofumpt, CI builds
`GOOS=windows`). Core library is stdlib-only so importing it costs nothing;
cobra appears only under `cmd/`.

Single root package — the per-harness knowledge is ~40 lines each, table-
driven, and does not warrant subpackages:

```
agentsummons.go      // facade: Build, Run, Version, Harnesses, InfoFor
request.go          // Request, Built, Result
harnesses.go        // the per-harness spec table — THE knowledge, single adjustment point
capabilities.go     // Capabilities manifest types (JSON-serializable)
versions.go         // LastValidated (flag surface) + VersionNewer
errors.go           // UnsupportedError, NotInstalledError
doc.go
*_test.go           // hermetic golden-argv tests
cmd/agentsummons/    // CLI: run, build, info, doctor
plans/              // this doc, future inventories
```

Harness IDs are string-identical to agentminutes `harness.ID` values
(`antigravity`, `claude-code`, `codex`) so consumers using both can map by
value. Harness lists stay alphabetical everywhere, same convention as
agentminutes. `VersionNewer` is a copy of the agentminutes lenient
comparison (leading-digit segments, unparseable ⇒ equal, never a false
positive) — copied, not imported, per the boundary rule.

## Library API

```go
type ID string

const (
    Antigravity ID = "antigravity"
    ClaudeCode  ID = "claude-code"
    Codex       ID = "codex"
)

// Request is one headless invocation. Zero-value optional fields are
// omitted from the command line; setting a field a harness cannot express
// is a loud *UnsupportedError, never a silent drop.
type Request struct {
    Harness ID
    Prompt  string
    Workdir string // required; headless runs are deliberate about cwd

    Model        string   // optional; harness-specific syntax
    SessionID    string   // optional; only ClaudeCode supports presetting
    Resume       string   // optional; continue an existing session by ref (all three support it)
    AllowedTools []string // optional; only ClaudeCode has the flag
    JSONOutput   bool     // request machine-readable stdout; unsupported on Antigravity
    AutoApprove  bool     // adds the harness's permission-bypass flags; never implied

    // SessionID and Resume are competing identity claims; setting both is an
    // *InvalidRequestError. agentsummons passes Resume through opaquely —
    // DISCOVERING a session ref where the harness doesn't emit one in-band
    // is deliberately out of scope (that's agentminutes' side of the seam).

    ExtraArgs []string // spliced before the prompt, verbatim — the escape hatch
    ExtraEnv  []string // KEY=VALUE appended to os.Environ()
}

// Built is the assembled command, pure data — powers `build`, dry runs,
// and callers that want to exec it themselves.
type Built struct {
    Argv        []string
    PromptIndex int // index of the prompt in Argv, for caller-side redaction
    Dir         string
    EnvAdded    []string
}

func Build(req Request) (Built, error)

// Run executes Build's output. A nonzero harness exit is data, not error:
// Result is returned with ExitCode set. Errors are agentsummons-level
// (invalid request, binary missing). On ctx timeout/cancel the process is
// killed and Run returns the partial Result alongside the ctx error, so
// experiments can archive failures.
func Run(ctx context.Context, req Request) (*Result, error)

type Result struct {
    Harness     ID
    Argv        []string
    PromptIndex int
    Workdir     string
    EnvAdded    []string
    Start, End  time.Time
    ExitCode    int
    Stdout      []byte
    Stderr      []byte
    SessionID   string // the preset ID when one was used, else ""
}

// Version runs the harness's version command and extracts a dotted version.
// Kept separate from Run so the library stays orthogonal; the CLI's run
// command calls it by default to stamp its envelope.
func Version(ctx context.Context, id ID) (string, error)

// Harnesses returns supported IDs, alphabetical. InfoFor returns the
// capability manifest for one harness.
func Harnesses() []ID
func InfoFor(id ID) (Capabilities, error)
```

`Capabilities` is a JSON-serializable manifest: binary name, base args,
which optional Request fields are supported and via which flag, the
auto-approve flags, the expected stdout shape (`json-envelope` /
`jsonl-events` / `text`), and a `notes` list carrying the quirk prose.
Deliberately *descriptive*, not *assembly instructions*: ordering rules
(e.g. Antigravity's prompt-last) live in `Build`, and non-Go callers who
need correct assembly use the CLI's `build` command rather than
reimplementing quirks from the manifest.

## Per-harness knowledge

Transplanted from the two proven implementations (agentminutes
`internal/driftprobe` and harness-llmify-prototype `runner/harnesses/`).
Flag-surface validation versions: Antigravity 1.1.3, Claude Code 2.1.204,
Codex 0.144.1.

### antigravity (`agy`)

- Version: `agy --version`.
- Workdir: `--add-dir <dir>` — agy ignores the process cwd without it. We
  set `cmd.Dir` too, for consistency, but the flag is what works.
- Prompt: `-p <prompt>`, and it MUST be the final arguments: `--print`
  consumes the next argument, so any flag after `-p` is swallowed as prompt
  text. `Build` enforces this ordering; `ExtraArgs` splice in before it.
- AutoApprove: `--dangerously-skip-permissions` — also the only
  non-interactive approval path, so headless agy effectively requires it
  (manifest note; still not implied).
- Model: `--model "<exact display string>"` as printed by `agy models`
  (e.g. `Gemini 3.5 Flash (Medium)`), not an API-style ID.
- Resume: `--conversation <id>` composes with `-p`. Validated live (1.1.3)
  via the composed seam: agentsummons turn 1 → `agentminutes sessions`
  time-window discovery → agentsummons `--conversation` turn 2, with recall
  confirmed in both stdout and the transcript. Caveat: headless agy does
  not emit the conversation ID in-band (open upstream request:
  <https://github.com/google-antigravity/antigravity-cli/issues/7> —
  worth watching; if it lands, the ref becomes in-band and the caveat
  relaxes), so the ref must be discovered
  post-hoc from the brain store — i.e. via agentminutes, not here. `-c`
  (continue most recent) exists but is racy on a shared machine.
  Cross-find for agentminutes: resumed conversations emit a
  `SYSTEM_MESSAGE` record its adapter had never observed (single-turn
  probes only) — recorded in that repo's plans/next-steps.md.
- Unsupported: SessionID (no preset; identity is discoverable only
  post-hoc), AllowedTools (no such flag — any tool restriction is
  prompt-level only), JSONOutput (stdout is step narration + "Summary of
  Work" text).
- Note: Go-based TLS treats `SSL_CERT_FILE` as the *entire* root set — a
  caller pointing it at a local CA must use a combined bundle.
- Note (reported, unvalidated): `-p` under a non-TTY can drop the final
  response from stdout in some versions — another reason callers should
  treat transcripts (via agentminutes) as the authoritative record of what
  the agent said, and stdout as best-effort.

### claudecode (`claude`)

- Version: `claude --version`.
- Workdir: `cmd.Dir`.
- Prompt: `-p <prompt>`, assembled last for consistency.
- AutoApprove: `--dangerously-skip-permissions`.
- Model: `--model <id>` (e.g. `claude-sonnet-5`).
- SessionID: `--session-id <uuid>` — the only harness that can preset its
  session identity; the JSON envelope echoes it back.
- Resume: `--resume <session-id>` composes with `-p`; combined with the
  envelope's `session_id`, multi-turn is fully in-band (no post-hoc
  discovery needed). Validated live (2.1.204): a resumed headless turn
  preserves the session ID, so the preset ref stays valid across turns;
  chaining from each envelope remains the robust habit.
- AllowedTools: `--allowedTools <comma-joined>`. Quirk (observed 2.1.197):
  in headless mode Glob/Grep are not registered unless explicitly named
  here, even with permissions bypassed.
- JSONOutput: `--output-format json` — single JSON envelope on stdout with
  `result`, `session_id`, `usage`, `total_cost_usd`.
- Note: Node-based; self-signed local HTTPS needs
  `NODE_TLS_REJECT_UNAUTHORIZED=0` via ExtraEnv (WebFetch force-upgrades
  HTTP to HTTPS).

### codex (`codex`)

- Version: `codex --version`.
- Base args: `exec --skip-git-repo-check` — headless is the `exec`
  subcommand; the skip flag is required outside a git repo and assumed
  harmless inside one (verify during implementation).
- Workdir: `cmd.Dir`.
- Prompt: bare positional argument, last.
- AutoApprove: `--dangerously-bypass-approvals-and-sandbox`. Finer sandbox
  control (`--sandbox read-only`,
  `-c sandbox_workspace_write.network_access=true`) is policy → ExtraArgs.
- Model: `-m <id>`.
- Resume: `codex exec resume <session-id> [prompt]` — a *subcommand*, not a
  flag, so the base args change shape when Resume is set (`exec resume`
  replaces `exec`). Validated live (0.144.1): the subcommand form works
  with `--skip-git-repo-check` and the other exec flags in place, the ref
  extracted from turn 1's `--json` event stream, recall confirmed via
  `--output-last-message`. This is exactly why assembly lives in `Build`
  code rather than in the manifest as data. The session ID comes from the
  `--json` event stream of the prior turn, or post-hoc via agentminutes.
- JSONOutput: `--json` — a JSONL *event stream*, not a single envelope
  (manifest records the shape so callers parse accordingly).
  `--output-last-message <path>` is the reliable way to capture the final
  answer; ExtraArgs for now (see Open questions).
- Unsupported: SessionID (post-hoc discovery only), AllowedTools (tool
  access is governed by sandbox config, not an allowlist flag).
- Note: Rust-based; do NOT set `SSL_CERT_FILE` to a bare local cert — it
  replaces the trust store rustls uses to reach the backend (observed on
  0.144.1). `CURL_CA_BUNDLE` scopes to shell curl only.

## CLI

`cmd/agentsummons`, cobra, mirrors agentminutes CLI conventions. All JSON
output carries `"schema_version": 1`; the envelope shapes are a
cross-language contract, versioned like the agentminutes schema.

- `agentsummons run --harness <id> -p <prompt> [--workdir DIR] [--model M]
  [--session-id ID] [--allowed-tools A,B] [--auto-approve] [--json-output]
  [--timeout 5m] [--extra-arg X ...] [--extra-env K=V ...] [--json]`
  — executes. Default behavior is a transparent wrapper: harness stdout/
  stderr pass through, harness exit code propagates. `--json` swaps to a
  Result envelope on stdout (harness stdout embedded as a string;
  `prompt_index` included so callers can redact argv before archiving),
  plus `harness_version` from a pre-run version probe and a `drift_hint`
  when installed > validated. agentsummons's own failures use exit codes
  distinct from common harness codes (usage/capability 64, binary missing
  69, timeout 70 — BSD sysexits territory to minimize collision; envelope
  is authoritative for `--json` consumers).
- `agentsummons build <same flags as run>` — prints the `Built` JSON (argv,
  prompt_index, dir, env_added) without executing. This is the
  correct-assembly path for non-Go callers; quirks stay implemented once.
- `agentsummons info [--harness <id>] [--json]` — capability manifest(s).
- `agentsummons doctor [--json]` — for each installed harness: installed
  version vs flag-surface validated version. Free (no tokens). Exit 0 all
  clean/absent, 1 when any installed version is newer than validated (drift
  candidate). Designed as the cheap pre-experiment gate.

## Version validation

`versions.go` holds this repo's own `LastValidated` map: the newest release
whose *flag surface* the spec table was validated against. It is
deliberately separate from agentminutes' table, which validates the
*transcript format* surface; the two drift independently.

Detection is layered, cheapest first:

1. `doctor` — free, compares versions only.
2. Any real invocation — flag removals/renames fail loudly at run time
   with a usage error; `run` surfaces the drift hint when the installed
   version is newer than validated.
3. `agentminutes drift probe` — the paid canary. Once agentminutes'
   driftprobe migrates onto this library, one probe invocation exercises
   both layers: an invocation-time failure is flag drift (agentsummons's to
   fix), a parse-time failure is format drift (agentminutes' to fix).

Updating `LastValidated` means re-checking each harness's `--help` output
and the spec table's flags against the new release, then bumping the entry
(process to be detailed in DEVELOPMENT.md).

## Testing

- **Hermetic:** `Build` is pure, so the core suite is golden-argv tests —
  every harness × every capability combination, plus ordering invariants
  (Antigravity prompt-last even with ExtraArgs) and `UnsupportedError`
  cases. No binaries, no network, runs everywhere including
  `GOOS=windows` builds.
- **Live (env-gated, spends tokens):** `AGENTSUMMONS_LIVE=<ids|all>` enables
  tests that invoke installed harnesses with a trivial prompt ("Reply with
  exactly: pong") and assert exit 0, non-empty output, the JSON envelope
  parses (claudecode), and the preset session ID is echoed (claudecode).
  Same pattern as agentminutes' `AGENTMINUTES_LOCAL_*` gates.
- **CI:** unit + lint + windows cross-build. Live tests never run in CI.

Windows caveat: binaries are resolved via `exec.LookPath`, which honors
PATHEXT (`claude.cmd` etc.); cross-build is CI-enforced but live behavior
on Windows is unvalidated until someone runs the gated tests there.

## Consumers and migration

- **agentminutes** (`internal/driftprobe`): replace the `Runner` table with
  `agentsummons.Build`/`Run` (keeping probe-specific retry and transcript
  logic); replace `Runner.TranscriptRoot` closures with the existing
  `Locator.DefaultRoot`, collapsing the duplicated root knowledge.
- **harness-llmify-prototype**: harness adapter modules shrink to pure
  policy (model choice, tool allowlist, sandbox args, TLS env) composed
  onto `agentsummons run --json` / `agentsummons build`; `doctor` becomes the
  pre-run gate next to the existing `agentminutes drift scan`.
- **Future experiments**: consume the CLI (any language) or the lib (Go).

## Multi-turn

Multi-turn works with this architecture because harness session state lives
on disk in each harness's own store; a "conversation" is just a sequence of
discrete headless invocations that name the same session. agentsummons stays
stateless per call — no daemon, no in-memory conversation object. The turn
loop composes across the seam:

1. Turn 1: `Run` (ClaudeCode: preset `SessionID` to know the ref up front).
2. Obtain the session ref: in-band from stdout (ClaudeCode envelope, Codex
   event stream) or post-hoc via `agentminutes sessions` (Antigravity, or
   whenever stdout wasn't captured in a parseable form).
3. Turn N: `Run` with `Resume` set to the ref.

Support matrix: ClaudeCode fully in-band; Codex in-band via `--json` (or
post-hoc); Antigravity resume flag exists but the ref is only discoverable
post-hoc. All three resume paths are live-validated at the LastValidated
versions: ClaudeCode and Codex by the in-repo live tests, Antigravity by
the composed agentsummons+agentminutes procedure (which cannot be an
in-repo test — it would need transcript knowledge, and a test dependency
on agentminutes would create the module cycle the boundary exists to
prevent). All three also have a continue-most-recent form (`--continue` /
`resume --last` / `-c`) — decided: not modeled. `Resume` requires an
explicit session ref, because continue-most-recent is racy against any
concurrent agent session on the machine (the same reason the agentminutes
drift probe filters foreign transcripts by cwd) and silently continuing
the wrong session is the worst failure mode an experiment can have.
Callers who truly want it can pass the flag via ExtraArgs.

## Containerized environments (evals)

Agent evals often need a reproducible environment for the agent to explore:
seeded content, pinned tool versions, constrained network. Docker slots in
as an execution substrate *around* agentsummons, not a feature *inside* it —
no `Request` changes, no docker knowledge in this repo. Two composition
patterns, both available with the API as designed:

- **agentsummons-in-container (recommended for evals).** Bake the harness
  binaries and the agentsummons binary into the eval image alongside the
  environment under test; the container command is
  `agentsummons run --json ...`. Container stdout then carries exactly one
  JSON envelope (harness stdout embedded), which is the only easy channel
  out of a container — the reason `run --json` must never share stdout
  with anything else. Transcript roots inside the container are
  volume-mounted to the host and parsed there with agentminutes (which
  accepts explicit roots; the *container-side* root paths are agentminutes
  knowledge the eval caller consults — agentsummons stays ignorant).
  Replicability comes from the image digest plus the envelope's
  `harness_version`; record both in eval metadata. `doctor` doubles as a
  bake-time assertion: a `RUN agentsummons doctor` step fails the image
  build when installed harness versions drift past validated.
- **Build-composed.** `Built` is pure data, so a host-side orchestrator can
  wrap it itself: `docker run -w <built.dir> -e <built.env_added...>
  <image> <built.argv...>`. For callers that keep orchestration on the
  host and only containerize the harness process.

Layer split for "env constraints we want to test": image content, tool
versions, resource limits, and coarse network are docker-level; tool
allowlists and sandbox levels are harness-level (`AllowedTools`, codex
sandbox config via ExtraArgs). They compose. One non-obvious limit:
`--network none` is unusable because the harness must reach its model API,
so constraining *agent* egress separately from harness API traffic needs
an egress proxy/allowlist, not a docker switch. And auth is the sharp
edge: credentials must exist inside the container (env keys or mounted
auth state); Antigravity's interactive OAuth makes it the hardest of the
three to containerize.

Constraint this adds to the build: release artifacts must be static
cross-compiled binaries (`CGO_ENABLED=0`; linux/amd64 and linux/arm64 at
minimum, in addition to host platforms) so that dropping agentsummons into
any image is a single `COPY`. The stdlib-only core already guarantees
this stays cheap.

## Open questions

- Vanity import path (`agentsummons.dev/...` via go-import meta tags) vs the
  GitHub module path. Trivial to change before the first public release,
  breaking after; decide before tagging v0.1.0.
- Promote codex `--output-last-message` from ExtraArgs to a first-class
  `Request` field? It is the only reliable final-answer capture for codex,
  which argues for first-class; it is also single-harness, which argues
  against. Revisit after llmify migrates.
- Streaming: add optional `Stdout/Stderr io.Writer` tee fields to `Request`
  when a consumer actually needs live output; contract stays "captured in
  Result" either way.
- Should `run --json` redact the prompt in `argv` by default (with a flag
  to include it), rather than only providing `prompt_index`? Leaning no —
  representational judgment calls belong to callers — but archiving
  workflows may prove the default wrong.
- ClaudeCode `--fork-session` (branching a resumed session instead of
  continuing it) is unmodeled and unexplored; reachable via ExtraArgs if
  an experiment wants it.
