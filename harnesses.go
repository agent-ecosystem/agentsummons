package agentsummons

import (
	"strconv"
	"strings"
)

// spec holds one harness's invocation knowledge: the capability manifest
// plus the assembly rules. This table is the single adjustment point when
// a harness changes its flags; it moves together with LastValidated in
// versions.go. Alphabetical, like all harness lists.
type spec struct {
	caps Capabilities

	// assemble returns the full argv, binary first and prompt last. It may
	// assume the Request passed shared validation: required fields are
	// present and every set optional field is supported.
	assemble func(req Request) []string
}

var specs = map[ID]spec{
	Antigravity: {
		caps: Capabilities{
			Harness:         Antigravity,
			Binary:          "agy",
			VersionArgs:     []string{"--version"},
			Prompt:          "-p <prompt>, assembled last (--print consumes the next argument)",
			Workdir:         "--add-dir <dir> plus process cwd (agy ignores the cwd without the flag)",
			Model:           `--model "<display name from agy models>"`,
			Resume:          "--conversation <id>; in-band multi-turn with JSONOutput (the envelope carries conversation_id), post-hoc discovery otherwise",
			JSONOutput:      "--output-format json",
			AutoApprove:     "--dangerously-skip-permissions (the only non-interactive approval path)",
			JSONOutputShape: "envelope",
			Notes: []string{
				"the upstream request for an in-band conversation ID (https://github.com/google-antigravity/antigravity-cli/issues/7) landed in 1.1.8 as --output-format json/stream-json; text-mode headless runs still never emit it, so only they need post-hoc discovery (e.g. agentminutes sessions)",
				"one -p invocation writes two conversations (a warm-up plus the real one), so post-hoc discovery must match on content (e.g. the recorded prompt), not recency alone (observed 1.1.4)",
				"--conversation appends to the same conversation transcript rather than forking (observed 1.1.4; envelope conversation_id stability re-confirmed 1.1.8)",
				"text-mode stdout is step narration plus a Summary of Work; --output-format json swaps it for a single result envelope (conversation_id, status, response, usage)",
				"print mode expands slash commands and skills in the prompt since 1.1.9, so a prompt starting with / may resolve to a command instead of being sent verbatim; --disable-slash-commands (via ExtraArgs) opts out",
				"text-mode -p under a non-TTY has been reported to drop the final response from stdout in some versions",
				"no tool-restriction flag exists; any read-only constraint is prompt-level only",
				"Go TLS treats SSL_CERT_FILE as the entire trust store; a local CA needs a combined bundle",
			},
		},
		assemble: func(req Request) []string {
			args := []string{"agy"}
			if req.AutoApprove {
				args = append(args, "--dangerously-skip-permissions")
			}
			// agy ignores the process cwd without --add-dir.
			args = append(args, "--add-dir", req.Workdir)
			if req.Model != "" {
				args = append(args, "--model", req.Model)
			}
			if req.Resume != "" {
				args = append(args, "--conversation", req.Resume)
			}
			if req.JSONOutput {
				args = append(args, "--output-format", "json")
			}
			args = append(args, req.ExtraArgs...)
			// --print consumes the next argument, so -p and the prompt go last.
			return append(args, "-p", req.Prompt)
		},
	},
	ClaudeCode: {
		caps: Capabilities{
			Harness:         ClaudeCode,
			Binary:          "claude",
			VersionArgs:     []string{"--version"},
			Prompt:          "-p <prompt>",
			Workdir:         "process cwd",
			Model:           "--model <id>",
			SessionID:       "--session-id <uuid>",
			Resume:          "--resume <session-id>; fully in-band multi-turn (the JSON envelope echoes session_id)",
			AllowedTools:    "--allowedTools <comma-joined>",
			JSONOutput:      "--output-format json",
			AutoApprove:     "--dangerously-skip-permissions",
			JSONOutputShape: "envelope",
			Notes: []string{
				"headless runs do not register Glob/Grep unless named in --allowedTools (observed 2.1.197)",
				"resume preserves the session_id (observed 2.1.212); chaining refs from each turn's envelope stays correct either way",
				"self-signed local HTTPS needs NODE_TLS_REJECT_UNAUTHORIZED=0 (WebFetch force-upgrades to HTTPS)",
			},
		},
		assemble: func(req Request) []string {
			args := []string{"claude"}
			if req.AutoApprove {
				args = append(args, "--dangerously-skip-permissions")
			}
			if req.Model != "" {
				args = append(args, "--model", req.Model)
			}
			if req.SessionID != "" {
				args = append(args, "--session-id", req.SessionID)
			}
			if req.Resume != "" {
				args = append(args, "--resume", req.Resume)
			}
			if len(req.AllowedTools) > 0 {
				args = append(args, "--allowedTools", strings.Join(req.AllowedTools, ","))
			}
			if req.JSONOutput {
				args = append(args, "--output-format", "json")
			}
			args = append(args, req.ExtraArgs...)
			return append(args, "-p", req.Prompt)
		},
	},
	Codex: {
		caps: Capabilities{
			Harness:         Codex,
			Binary:          "codex",
			VersionArgs:     []string{"--version"},
			BaseArgs:        []string{"exec", "--skip-git-repo-check"},
			Prompt:          "bare positional, last",
			Workdir:         "process cwd",
			Model:           "-m <id>",
			Resume:          "exec resume <session-id> — a subcommand, so the base args change shape",
			JSONOutput:      "--json",
			AutoApprove:     "--dangerously-bypass-approvals-and-sandbox",
			JSONOutputShape: "jsonl-events",
			Notes: []string{
				"--json emits a JSONL event stream, not one envelope; --output-last-message <path> (via ExtraArgs) captures the final answer reliably",
				"exec resume appends to the same rollout file and preserves the session id (observed 0.144.6)",
				"finer sandbox control is caller policy via ExtraArgs: --sandbox read-only, -c sandbox_workspace_write.network_access=true",
				"do not set SSL_CERT_FILE to a bare local cert: it replaces the rustls trust store used to reach the backend (observed 0.144.1)",
			},
		},
		assemble: func(req Request) []string {
			args := []string{"codex", "exec"}
			// Resuming is a subcommand of exec, not a flag.
			if req.Resume != "" {
				args = append(args, "resume")
			}
			args = append(args, "--skip-git-repo-check")
			if req.AutoApprove {
				args = append(args, "--dangerously-bypass-approvals-and-sandbox")
			}
			if req.Model != "" {
				args = append(args, "-m", req.Model)
			}
			if req.JSONOutput {
				args = append(args, "--json")
			}
			args = append(args, req.ExtraArgs...)
			if req.Resume != "" {
				args = append(args, req.Resume)
			}
			return append(args, req.Prompt)
		},
	},
}

// specFor resolves the spec table entry for id, the shared lookup behind
// every entry point that takes a harness ID.
func specFor(id ID) (spec, error) {
	sp, ok := specs[id]
	if !ok {
		return spec{}, &InvalidRequestError{Reason: "unknown harness " + string(id)}
	}
	return sp, nil
}

// validate resolves the harness spec and enforces the Request contract:
// required fields, mutually exclusive identity claims, and loud rejection
// of any optional field the harness cannot express (supported exactly when
// the corresponding Capabilities field is non-empty).
func validate(req Request) (spec, error) {
	sp, err := specFor(req.Harness)
	if err != nil {
		return spec{}, err
	}
	if req.Prompt == "" {
		return spec{}, &InvalidRequestError{Reason: "prompt is required"}
	}
	if req.Workdir == "" {
		return spec{}, &InvalidRequestError{Reason: "workdir is required"}
	}
	if req.SessionID != "" && req.Resume != "" {
		return spec{}, &InvalidRequestError{Reason: "SessionID and Resume are competing identity claims; set at most one"}
	}
	for _, kv := range req.ExtraEnv {
		if i := strings.IndexByte(kv, '='); i <= 0 {
			return spec{}, &InvalidRequestError{Reason: "ExtraEnv entry " + strconv.Quote(kv) + " is not KEY=VALUE"}
		}
	}
	caps := sp.caps
	switch {
	case req.Model != "" && caps.Model == "":
		return spec{}, &UnsupportedError{Harness: req.Harness, Option: "Model"}
	case req.SessionID != "" && caps.SessionID == "":
		return spec{}, &UnsupportedError{Harness: req.Harness, Option: "SessionID"}
	case req.Resume != "" && caps.Resume == "":
		return spec{}, &UnsupportedError{Harness: req.Harness, Option: "Resume"}
	case len(req.AllowedTools) > 0 && caps.AllowedTools == "":
		return spec{}, &UnsupportedError{Harness: req.Harness, Option: "AllowedTools"}
	case req.JSONOutput && caps.JSONOutput == "":
		return spec{}, &UnsupportedError{Harness: req.Harness, Option: "JSONOutput"}
	case req.AutoApprove && caps.AutoApprove == "":
		// Unreachable while every harness documents an AutoApprove path
		// (test-enforced), but kept so a future harness without one rejects
		// loudly instead of silently dropping the field.
		return spec{}, &UnsupportedError{Harness: req.Harness, Option: "AutoApprove"}
	}
	return sp, nil
}
