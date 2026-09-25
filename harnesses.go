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
				"--conversation appends to the same conversation transcript rather than forking (observed 1.1.4; envelope conversation_id stability re-confirmed 1.2.11)",
				"text-mode stdout is step narration plus a Summary of Work; --output-format json swaps it for a single result envelope (conversation_id, status, response, usage; since 1.1.27 refused tool actions are listed as denied_actions instead of being skipped silently)",
				"print mode expands slash commands and skills in the prompt since 1.1.9, so a prompt starting with / may resolve to a command instead of being sent verbatim; --disable-slash-commands (via ExtraArgs) opts out; since 1.1.11/1.1.12 read-only commands answer without an agent turn and interactive-only ones fail loudly",
				"text-mode -p under a non-TTY has been reported to drop the final response from stdout in some versions; 1.1.18 fixed the related dropped-stream case, which previously reported an empty response as a clean exit 0 and now exits non-zero; since 1.1.28 fatal headless errors carry a stable error: marker on stderr, and since 1.2.6 agent and model API failures instead print a structured AGY_ERROR: {...} JSON line on stderr (status, error code, retryability, error ID) and exit 3 rather than 1; since 1.2.10 a run that streamed part of a response before failing also exits 3 with the AGY_ERROR line instead of 0, and the JSON error output carries the partial response",
				"print mode waits at most --print-timeout: the default was 5m through 1.2.5 and is unlimited since 1.2.6 (0 waits for the turn to complete); when a limit is set, expiry returns the partial output and exits 0 with only a stderr warning (since 1.1.28), so an explicit --print-timeout (via ExtraArgs) can truncate in a way that looks like success; since 1.2.9 still-running background tasks are waited for until that deadline (capped at 30 minutes) instead of being cancelled about 5s after the agent goes idle",
				"through 1.2.8 headless runs could leave daemon background processes holding stdout open after exit, so a reader waiting for EOF could hang; 1.2.9 terminates them when the run ends",
				"--effort low|medium|high|max (since 1.2.11) is reachable via ExtraArgs; there is no Request field for it",
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
				"resume preserves the session_id (observed 2.1.212; re-confirmed 2.1.274); chaining refs from each turn's envelope stays correct either way",
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
				"exec resume appends to the same rollout file and preserves the session id (observed 0.144.6; re-confirmed 0.157.0); a separate exec fork subcommand (present by 0.154.0) forks into a new session instead, but subcommands are not reachable via ExtraArgs, so Resume always means append here",
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
	Copilot: {
		caps: Capabilities{
			Harness:         Copilot,
			Binary:          "copilot",
			VersionArgs:     []string{"--version"},
			Prompt:          "-p <prompt>",
			Workdir:         "process cwd (-C <dir> also exists, via ExtraArgs)",
			Model:           "--model <id> (e.g. claude-sonnet-5, gpt-5.3-codex, or auto)",
			SessionID:       "--session-id <uuid>",
			Resume:          "--resume=<session-id>; fully in-band multi-turn (the final result event carries sessionId)",
			AllowedTools:    "--available-tools=<comma-joined>",
			JSONOutput:      "--output-format json",
			AutoApprove:     "--allow-all (the documented equivalent of --yolo: --allow-all-tools --allow-all-paths --allow-all-urls)",
			JSONOutputShape: "jsonl-events",
			Notes: []string{
				"--output-format json emits a JSONL event stream whose final line is a result event carrying sessionId, exitCode, and usage; the answer text is the content of assistant.message events (observed 1.0.88)",
				"resume appends to the same session and preserves the session id (observed 1.0.88); --resume takes an optional value (bare --resume opens an interactive picker), so it is passed as one --resume=<id> argument",
				"--session-id resumes an existing session when the UUID already exists instead of failing, so a preset ref must be fresh",
				"without a bypass flag a headless run still executes read-only tools, but every write (file creation, shell redirection, touch) fails with \"Permission denied and could not request permission from user\" and the run still exits 0 with a result event, so a task that needed writes can look like a clean success (observed 1.0.88)",
				"--allow-all-tools alone (via ExtraArgs) auto-approves tools while keeping path and URL checks; COPILOT_ALLOW_ALL=true (via ExtraEnv) additionally trusts the working directory, which loads its hooks, plugins, and MCP servers",
				"--available-tools also drops the built-in GitHub MCP server's tools, not just built-in tools (observed 1.0.88: view,grep left exactly those two); --allow-tool/--deny-tool (via ExtraArgs) shape permissions rather than availability",
				"the built-in GitHub MCP server connects on every run; --disable-builtin-mcps (via ExtraArgs) skips it",
				"auto-update is on by default outside CI (detected via CI, BUILD_NUMBER, RUN_ID, or SYSTEM_COLLECTIONURI); --no-auto-update (via ExtraArgs) or COPILOT_AUTO_UPDATE=false (via ExtraEnv) pins the installed release",
				"-s/--silent (via ExtraArgs) strips the stats footer from text-mode stdout; --usage-output-file <path> writes final usage statistics as JSON",
				"non-interactive auth: COPILOT_GITHUB_TOKEN, GH_TOKEN, or GITHUB_TOKEN (in that order of precedence) override stored credentials",
			},
		},
		assemble: func(req Request) []string {
			args := []string{"copilot"}
			if req.AutoApprove {
				args = append(args, "--allow-all")
			}
			if req.Model != "" {
				args = append(args, "--model", req.Model)
			}
			if req.SessionID != "" {
				args = append(args, "--session-id", req.SessionID)
			}
			if req.Resume != "" {
				// --resume's value is optional, so join it to keep the
				// argument unambiguous.
				args = append(args, "--resume="+req.Resume)
			}
			if len(req.AllowedTools) > 0 {
				// Variadic flag: the joined form keeps it from swallowing
				// ExtraArgs or the prompt.
				args = append(args, "--available-tools="+strings.Join(req.AllowedTools, ","))
			}
			if req.JSONOutput {
				args = append(args, "--output-format", "json")
			}
			args = append(args, req.ExtraArgs...)
			return append(args, "-p", req.Prompt)
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
