package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// agentsummons-level exit codes, from the BSD sysexits range to minimize
// collision with harness exit codes, which run propagates verbatim.
const (
	exitUsage       = 64 // EX_USAGE: bad request or unsupported capability
	exitUnavailable = 69 // EX_UNAVAILABLE: harness binary not found in PATH
	exitTimeout     = 75 // EX_TEMPFAIL: harness killed by --timeout
)

// exitError carries an exit code through cobra to main.
type exitError struct{ code int }

func (e *exitError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}

// silentExit returns an exit-code sentinel after the command has already
// said everything, so cobra doesn't print a second error line.
func silentExit(cmd *cobra.Command, code int) error {
	cmd.SilenceErrors = true
	return &exitError{code: code}
}
