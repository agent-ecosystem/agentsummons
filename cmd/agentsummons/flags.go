package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/agent-ecosystem/agentsummons"
	"github.com/spf13/cobra"
)

// harnessHelpList renders the supported harness IDs for flag help text,
// e.g. `"antigravity", "claude-code", or "codex"` — derived so a new
// harness appears in the help automatically.
func harnessHelpList() string {
	ids := agentsummons.Harnesses()
	quoted := make([]string, len(ids))
	for i, id := range ids {
		quoted[i] = strconv.Quote(string(id))
	}
	if len(quoted) == 1 {
		return quoted[0]
	}
	return strings.Join(quoted[:len(quoted)-1], ", ") + ", or " + quoted[len(quoted)-1]
}

// requestFlags is the invocation surface shared by run and build.
type requestFlags struct {
	harness      string
	prompt       string
	workdir      string
	model        string
	sessionID    string
	resume       string
	allowedTools []string
	autoApprove  bool
	jsonOutput   bool
	extraArgs    []string
	extraEnv     []string
}

func addRequestFlags(cmd *cobra.Command, f *requestFlags) {
	cmd.Flags().StringVar(&f.harness, "harness", "", "harness to invoke: "+harnessHelpList()+" (required)")
	cmd.Flags().StringVarP(&f.prompt, "prompt", "p", "", "prompt text (required)")
	cmd.Flags().StringVar(&f.workdir, "workdir", "", "working directory for the agent (default: current directory)")
	cmd.Flags().StringVar(&f.model, "model", "", "pin the model, in the harness's own syntax (see info)")
	cmd.Flags().StringVar(&f.sessionID, "session-id", "", "preset the session identity (claude-code only)")
	cmd.Flags().StringVar(&f.resume, "resume", "", "continue an existing session by explicit ref (see info per harness)")
	cmd.Flags().StringSliceVar(&f.allowedTools, "allowed-tools", nil, "restrict the agent to these tools (claude-code only)")
	cmd.Flags().BoolVar(&f.autoApprove, "auto-approve", false, "add the harness's permission-bypass flags (never implied)")
	cmd.Flags().BoolVar(&f.jsonOutput, "json-output", false, "ask the harness for machine-readable stdout")
	cmd.Flags().StringArrayVar(&f.extraArgs, "extra-arg", nil, "extra argument spliced in before the prompt (repeatable)")
	cmd.Flags().StringArrayVar(&f.extraEnv, "extra-env", nil, "KEY=VALUE added to the harness environment (repeatable)")
	_ = cmd.MarkFlagRequired("harness")
	_ = cmd.MarkFlagRequired("prompt")
}

func (f *requestFlags) request() (agentsummons.Request, error) {
	workdir := f.workdir
	if workdir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return agentsummons.Request{}, fmt.Errorf("resolving current directory for --workdir: %w", err)
		}
		workdir = wd
	}
	return agentsummons.Request{
		Harness:      agentsummons.ID(f.harness),
		Prompt:       f.prompt,
		Workdir:      workdir,
		Model:        f.model,
		SessionID:    f.sessionID,
		Resume:       f.resume,
		AllowedTools: f.allowedTools,
		JSONOutput:   f.jsonOutput,
		AutoApprove:  f.autoApprove,
		ExtraArgs:    f.extraArgs,
		ExtraEnv:     f.extraEnv,
	}, nil
}

// requestExit maps an agentsummons-level error to the CLI's exit codes.
func requestExit(cmd *cobra.Command, err error) error {
	var (
		ue  *agentsummons.UnsupportedError
		ire *agentsummons.InvalidRequestError
		nie *agentsummons.NotInstalledError
	)
	switch {
	case errors.As(err, &ue), errors.As(err, &ire):
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err)
		return silentExit(cmd, exitUsage)
	case errors.As(err, &nie):
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err)
		return silentExit(cmd, exitUnavailable)
	}
	return err
}
