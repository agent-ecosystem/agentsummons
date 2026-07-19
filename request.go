package agentsummons

import "time"

// Request is one headless invocation. Zero-value optional fields are
// omitted from the command line; setting a field the harness cannot
// express is a loud *UnsupportedError, never a silent drop. The
// cross-language JSON contract is the CLI's output envelopes (see its
// schema_version), not these structs' own encodings: Result's Stdout and
// Stderr are []byte, which marshal as base64, where the run envelope
// re-emits them as strings.
type Request struct {
	// Harness selects the agent CLI to invoke.
	Harness ID
	// Prompt is the task text. Required.
	Prompt string
	// Workdir is the directory the agent works in. Required: headless
	// runs are deliberate about their working directory.
	Workdir string

	// Model pins the model, in the harness's own syntax (see Capabilities;
	// Antigravity wants the display name from `agy models`, not an API id).
	Model string
	// SessionID presets the session identity. ClaudeCode only.
	SessionID string
	// Resume continues an existing session by explicit ref. Mutually
	// exclusive with SessionID: they are competing identity claims.
	// agentsummons passes the ref through opaquely — discovering a ref the
	// harness never emitted in-band is agentminutes' side of the seam.
	Resume string
	// AllowedTools restricts the agent to the named tools. ClaudeCode only.
	AllowedTools []string
	// JSONOutput asks the harness for machine-readable stdout. Unsupported
	// on Antigravity; the shape differs per harness (see Capabilities).
	JSONOutput bool
	// AutoApprove adds the harness's permission-bypass flags. Never
	// implied, even where headless operation effectively requires it.
	AutoApprove bool

	// ExtraArgs are spliced in verbatim before the prompt — the escape
	// hatch for anything the fields above don't model.
	ExtraArgs []string
	// ExtraEnv is KEY=VALUE pairs appended to the parent environment. An
	// entry without that shape is a loud *InvalidRequestError.
	ExtraEnv []string
}

// Built is an assembled command as pure data: enough for any caller — in
// any language, or wrapping the argv in a container — to execute it
// exactly as Run would.
type Built struct {
	// Argv is the full command line; Argv[0] is the unresolved binary name.
	Argv []string `json:"argv"`
	// PromptIndex is the prompt's position in Argv, so callers can redact
	// the prompt before archiving the command line.
	PromptIndex int `json:"prompt_index"`
	// Dir is the working directory the command must run in.
	Dir string `json:"dir"`
	// EnvAdded is the KEY=VALUE pairs to append to the parent environment.
	EnvAdded []string `json:"env_added,omitempty"`
}

// Result is everything observable about one finished invocation. A nonzero
// harness exit is recorded here, not returned as an error: a failed run is
// still a run, and callers archive failures.
type Result struct {
	Harness     ID        `json:"harness"`
	Argv        []string  `json:"argv"`
	PromptIndex int       `json:"prompt_index"`
	Workdir     string    `json:"workdir"`
	EnvAdded    []string  `json:"env_added,omitempty"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	// ExitCode is the harness's exit code; -1 when the process was killed
	// (context timeout or cancellation).
	ExitCode int    `json:"exit_code"`
	Stdout   []byte `json:"stdout"`
	Stderr   []byte `json:"stderr"`
	// SessionID echoes Request.SessionID when one was preset, else "".
	// Session refs the harness only emits in its own output or transcript
	// store are the caller's to extract.
	SessionID string `json:"session_id,omitempty"`
}
