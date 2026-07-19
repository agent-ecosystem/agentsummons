package agentsummons

import "fmt"

// UnsupportedError reports a Request field set for a harness that cannot
// express it. Loud by design: a capability the harness lacks is never
// silently dropped from the command line.
type UnsupportedError struct {
	Harness ID
	Option  string
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("agentsummons: harness %q does not support %s", e.Harness, e.Option)
}

// InvalidRequestError reports a Request that is malformed regardless of
// harness capabilities (unknown harness, missing prompt or workdir,
// competing identity claims).
type InvalidRequestError struct {
	Reason string
}

func (e *InvalidRequestError) Error() string {
	return "agentsummons: invalid request: " + e.Reason
}

// NotInstalledError reports that a harness binary could not be resolved
// via PATH.
type NotInstalledError struct {
	Harness ID
	Binary  string
	Err     error
}

func (e *NotInstalledError) Error() string {
	return fmt.Sprintf("agentsummons: harness %q binary %q not found in PATH: %v", e.Harness, e.Binary, e.Err)
}

func (e *NotInstalledError) Unwrap() error { return e.Err }
