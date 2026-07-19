package agentsummons

// Capabilities is the manifest of what one harness supports and via which
// syntax. It is descriptive, not assembly instructions: ordering rules
// (e.g. Antigravity's prompt-must-be-last) live in Build, and non-Go
// callers who need correct assembly should use the CLI's build command
// rather than reassembling argv from this manifest.
//
// The optional-capability fields (Model through JSONOutput) hold the
// harness's syntax and are empty when the harness cannot express that
// capability; Build rejects a Request setting the corresponding field
// with a *UnsupportedError.
type Capabilities struct {
	Harness     ID       `json:"harness"`
	Binary      string   `json:"binary"`
	VersionArgs []string `json:"version_args"`

	// BaseArgs are always present, immediately after the binary in the
	// minimal invocation; a subcommand-shaped capability (Codex's resume)
	// may splice between them.
	BaseArgs []string `json:"base_args,omitempty"`
	// Prompt describes how the prompt is passed. Always supported.
	Prompt string `json:"prompt"`
	// Workdir describes how the working directory takes effect.
	Workdir string `json:"workdir"`

	Model        string `json:"model,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	Resume       string `json:"resume,omitempty"`
	AllowedTools string `json:"allowed_tools,omitempty"`
	JSONOutput   string `json:"json_output,omitempty"`
	AutoApprove  string `json:"auto_approve"`

	// JSONOutputShape is what stdout looks like with JSONOutput set:
	// "envelope" (one JSON object) or "jsonl-events" (a JSONL event
	// stream). Empty when JSONOutput is unsupported; without JSONOutput,
	// every harness's stdout is free text.
	JSONOutputShape string `json:"json_output_shape,omitempty"`

	// Notes carries the quirk prose that doesn't fit a field. Informational
	// only; nothing in it is ever applied automatically.
	Notes []string `json:"notes,omitempty"`
}
