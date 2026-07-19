package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/agent-ecosystem/agentsummons"
	"github.com/spf13/cobra"
)

// defaultTimeout bounds run's harness invocation unless --timeout overrides
// it: an unattended headless run should fail loudly rather than hang a
// pipeline indefinitely.
const defaultTimeout = 5 * time.Minute

// runEnvelope is the run --json output shape.
type runEnvelope struct {
	SchemaVersion  int       `json:"schema_version"`
	Harness        string    `json:"harness"`
	HarnessVersion string    `json:"harness_version,omitempty"`
	DriftHint      string    `json:"drift_hint,omitempty"`
	Argv           []string  `json:"argv"`
	PromptIndex    int       `json:"prompt_index"`
	Workdir        string    `json:"workdir"`
	EnvAdded       []string  `json:"env_added,omitempty"`
	Start          time.Time `json:"start"`
	End            time.Time `json:"end"`
	ExitCode       int       `json:"exit_code"`
	TimedOut       bool      `json:"timed_out,omitempty"`
	SessionID      string    `json:"session_id,omitempty"`
	Stdout         string    `json:"stdout"`
	Stderr         string    `json:"stderr"`
}

func newRunCmd() *cobra.Command {
	var (
		f       requestFlags
		timeout time.Duration
		asJSON  bool
	)
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Invoke a harness headlessly and report what happened",
		Long: `Run invokes the harness once, headlessly, in the given working
directory. By default it is a transparent wrapper: harness stdout and
stderr pass through, and the harness exit code propagates. With --json it
instead emits a single result envelope on stdout (harness output embedded)
— the shape containers and scripts consume.

agentsummons's own failures use exit codes distinct from common harness
codes: 64 bad request or unsupported capability, 69 harness not installed,
75 timeout.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(cmd, &f, timeout, asJSON)
		},
	}
	addRequestFlags(cmd, &f)
	cmd.Flags().DurationVar(&timeout, "timeout", defaultTimeout, "kill the harness after this long (0 = no timeout)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit a JSON result envelope instead of passing output through")
	return cmd
}

func runRun(cmd *cobra.Command, f *requestFlags, timeout time.Duration, asJSON bool) error {
	req, err := f.request()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Stamp the envelope with the installed version, best-effort: a probe
	// failure shouldn't fail the run it decorates.
	var harnessVersion string
	if asJSON {
		harnessVersion, _ = agentsummons.Version(ctx, req.Harness)
	}

	res, runErr := agentsummons.Run(ctx, req)
	if res == nil {
		return requestExit(cmd, runErr)
	}
	// Only a deadline is a timeout; cancellation (no path to it today, but
	// signal-aware contexts are a plausible future) reports as a plain error
	// rather than masquerading as one.
	timedOut := runErr != nil && errors.Is(runErr, context.DeadlineExceeded)
	if runErr != nil && !timedOut {
		return runErr
	}

	if asJSON {
		envelope := runEnvelope{
			SchemaVersion:  schemaVersion,
			Harness:        string(res.Harness),
			HarnessVersion: harnessVersion,
			DriftHint:      driftHint(req.Harness, harnessVersion),
			Argv:           res.Argv,
			PromptIndex:    res.PromptIndex,
			Workdir:        res.Workdir,
			EnvAdded:       res.EnvAdded,
			Start:          res.Start,
			End:            res.End,
			ExitCode:       res.ExitCode,
			TimedOut:       timedOut,
			SessionID:      res.SessionID,
			Stdout:         string(res.Stdout),
			Stderr:         string(res.Stderr),
		}
		if err := writeJSON(cmd, envelope); err != nil {
			return err
		}
	} else {
		if _, err := cmd.OutOrStdout().Write(res.Stdout); err != nil {
			return err
		}
		if _, err := cmd.ErrOrStderr().Write(res.Stderr); err != nil {
			return err
		}
	}

	switch {
	case timedOut:
		if !asJSON {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "agentsummons: harness killed after %s timeout\n", timeout)
		}
		return silentExit(cmd, exitTimeout)
	case res.ExitCode != 0:
		return silentExit(cmd, res.ExitCode)
	}
	return nil
}

// driftHint flags an installed version newer than the flag surface this
// build was validated against.
func driftHint(id agentsummons.ID, installed string) string {
	validated := agentsummons.LastValidated[id]
	if installed == "" || !agentsummons.VersionNewer(installed, validated) {
		return ""
	}
	return fmt.Sprintf("installed %s is newer than flag-surface validated %s; check for an agentsummons update", installed, validated)
}
