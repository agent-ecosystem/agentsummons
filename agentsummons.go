package agentsummons

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// ID identifies a supported harness. Values are string-identical to the
// agentminutes harness IDs so consumers of both libraries can map by value.
type ID string

// Supported harnesses, alphabetical.
const (
	Antigravity ID = "antigravity"
	ClaudeCode  ID = "claude-code"
	Codex       ID = "codex"
)

// Harnesses returns the supported harness IDs, alphabetical.
func Harnesses() []ID {
	return []ID{Antigravity, ClaudeCode, Codex}
}

// InfoFor returns the capability manifest for one harness. The manifest is
// the caller's copy: mutating it never touches the spec table.
func InfoFor(id ID) (Capabilities, error) {
	sp, err := specFor(id)
	if err != nil {
		return Capabilities{}, err
	}
	caps := sp.caps
	caps.VersionArgs = append([]string(nil), caps.VersionArgs...)
	caps.BaseArgs = append([]string(nil), caps.BaseArgs...)
	caps.Notes = append([]string(nil), caps.Notes...)
	return caps, nil
}

// Build assembles the command for req as pure data, without executing
// anything. The prompt is always the final Argv element.
func Build(req Request) (Built, error) {
	sp, err := validate(req)
	if err != nil {
		return Built{}, err
	}
	argv := sp.assemble(req)
	return Built{
		Argv:        argv,
		PromptIndex: len(argv) - 1,
		Dir:         req.Workdir,
		EnvAdded:    append([]string(nil), req.ExtraEnv...),
	}, nil
}

// waitDelay bounds how long Run waits for I/O to drain after the context
// kills the harness: agent CLIs spawn subprocesses that can hold the
// output pipes open past the parent's death.
const waitDelay = 10 * time.Second

// Run executes Build's output and reports everything observable about the
// invocation. A nonzero harness exit is data (Result.ExitCode), not an
// error; errors are agentsummons-level: invalid request, binary missing,
// start failure, or context timeout/cancellation — in the context case the
// partial Result is returned alongside the error so callers can archive
// what was captured.
func Run(ctx context.Context, req Request) (*Result, error) {
	built, err := Build(req)
	if err != nil {
		return nil, err
	}
	path, err := exec.LookPath(built.Argv[0])
	if err != nil {
		return nil, &NotInstalledError{Harness: req.Harness, Binary: built.Argv[0], Err: err}
	}
	cmd := exec.CommandContext(ctx, path, built.Argv[1:]...)
	cmd.Dir = built.Dir
	cmd.Env = append(os.Environ(), built.EnvAdded...)
	cmd.WaitDelay = waitDelay
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	res := &Result{
		Harness:     req.Harness,
		Argv:        built.Argv,
		PromptIndex: built.PromptIndex,
		Workdir:     built.Dir,
		EnvAdded:    built.EnvAdded,
		SessionID:   req.SessionID,
		Start:       time.Now(),
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("agentsummons: starting %s: %w", built.Argv[0], err)
	}
	waitErr := cmd.Wait()
	res.End = time.Now()
	res.Stdout = stdout.Bytes()
	res.Stderr = stderr.Bytes()
	res.ExitCode = cmd.ProcessState.ExitCode()

	if ctxErr := ctx.Err(); ctxErr != nil {
		// Windows reports a Kill as exit code 1, indistinguishable from a
		// real harness exit 1; the ExitCode contract is -1 whenever the
		// run was cut short. waitErr == nil means the harness finished
		// before the kill landed — its real code stands then.
		if waitErr != nil {
			res.ExitCode = -1
		}
		return res, ctxErr
	}
	var exitErr *exec.ExitError
	if waitErr != nil && !errors.As(waitErr, &exitErr) {
		return res, fmt.Errorf("agentsummons: waiting on %s: %w", built.Argv[0], waitErr)
	}
	return res, nil
}

// Version runs the harness's version command and extracts the first dotted
// version from its output. It is kept separate from Run so the primitives
// stay orthogonal; the CLI's run command calls it to stamp its envelope.
func Version(ctx context.Context, id ID) (string, error) {
	sp, err := specFor(id)
	if err != nil {
		return "", err
	}
	path, err := exec.LookPath(sp.caps.Binary)
	if err != nil {
		return "", &NotInstalledError{Harness: id, Binary: sp.caps.Binary, Err: err}
	}
	out, err := exec.CommandContext(ctx, path, sp.caps.VersionArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("agentsummons: %s %v: %w", sp.caps.Binary, sp.caps.VersionArgs, err)
	}
	v := versionRe.FindString(string(out))
	if v == "" {
		return "", fmt.Errorf("agentsummons: no dotted version in %s output %q", sp.caps.Binary, bytes.TrimSpace(out))
	}
	return v, nil
}
