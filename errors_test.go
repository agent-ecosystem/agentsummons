package agentsummons

import (
	"errors"
	"strings"
	"testing"
)

// TestErrorStrings pins the identifying content of each typed error's
// message: which harness, which field, and (for NotInstalledError) which
// binary — the parts a caller or log reader diagnoses from.
func TestErrorStrings(t *testing.T) {
	lookupErr := errors.New("nope")
	cases := []struct {
		name string
		err  error
		want []string
	}{
		{
			"unsupported",
			&UnsupportedError{Harness: Antigravity, Option: "JSONOutput"},
			[]string{"antigravity", "JSONOutput"},
		},
		{
			"invalid request",
			&InvalidRequestError{Reason: "prompt is required"},
			[]string{"invalid request", "prompt is required"},
		},
		{
			"not installed",
			&NotInstalledError{Harness: ClaudeCode, Binary: "claude", Err: lookupErr},
			[]string{"claude-code", `"claude"`, "PATH", "nope"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := tc.err.Error()
			for _, want := range tc.want {
				if !strings.Contains(msg, want) {
					t.Errorf("Error() = %q, want it to contain %q", msg, want)
				}
			}
		})
	}
}

func TestNotInstalledUnwrap(t *testing.T) {
	lookupErr := errors.New("nope")
	err := &NotInstalledError{Harness: ClaudeCode, Binary: "claude", Err: lookupErr}
	if !errors.Is(err, lookupErr) {
		t.Error("NotInstalledError does not unwrap to the lookup error")
	}
}
