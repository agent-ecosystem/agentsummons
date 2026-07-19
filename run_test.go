package agentsummons

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agent-ecosystem/agentsummons/internal/fakeharness"
)

// TestMain lets the test binary double as a fake harness: Run tests copy
// the binary into a temp PATH dir under a harness's name and drive its
// behavior through FAKE_* variables (see internal/fakeharness).
func TestMain(m *testing.M) {
	if os.Getenv(fakeharness.EnvEnable) == "1" {
		fakeharness.Main()
		return
	}
	os.Exit(m.Run())
}

func TestRunCapturesEverything(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_STDOUT", "the answer")
	t.Setenv("FAKE_STDERR", "some noise")
	req := Request{Harness: ClaudeCode, Prompt: "hi", Workdir: t.TempDir(), SessionID: "sess-1"}
	res, err := Run(context.Background(), req)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", res.ExitCode)
	}
	if string(res.Stdout) != "the answer" {
		t.Errorf("Stdout = %q", res.Stdout)
	}
	if string(res.Stderr) != "some noise" {
		t.Errorf("Stderr = %q", res.Stderr)
	}
	if res.SessionID != "sess-1" {
		t.Errorf("SessionID = %q, want the preset ref", res.SessionID)
	}
	if res.End.Before(res.Start) {
		t.Errorf("End %v before Start %v", res.End, res.Start)
	}
	if res.PromptIndex != len(res.Argv)-1 || res.Argv[res.PromptIndex] != "hi" {
		t.Errorf("prompt not last in Argv %q", res.Argv)
	}
}

func TestRunNonzeroExitIsData(t *testing.T) {
	fakeharness.Install(t, "codex")
	t.Setenv("FAKE_EXIT", "3")
	res, err := Run(context.Background(), Request{Harness: Codex, Prompt: "hi", Workdir: t.TempDir()})
	if err != nil {
		t.Fatalf("Run returned error for nonzero exit: %v", err)
	}
	if res.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", res.ExitCode)
	}
}

func TestRunUsesWorkdir(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_PRINT_CWD", "1")
	wd := t.TempDir()
	res, err := Run(context.Background(), Request{Harness: ClaudeCode, Prompt: "hi", Workdir: wd})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got, err := filepath.EvalSymlinks(string(res.Stdout))
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", res.Stdout, err)
	}
	want, err := filepath.EvalSymlinks(wd)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", wd, err)
	}
	if got != want {
		t.Errorf("harness cwd = %q, want %q", got, want)
	}
}

func TestRunAppendsExtraEnv(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_ECHO_ENV", "AGENTSUMMONS_TEST_EXTRA")
	res, err := Run(context.Background(), Request{
		Harness: ClaudeCode, Prompt: "hi", Workdir: t.TempDir(),
		ExtraEnv: []string{"AGENTSUMMONS_TEST_EXTRA=visible"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if string(res.Stdout) != "visible" {
		t.Errorf("child saw %q for the extra env var, want %q", res.Stdout, "visible")
	}
}

func TestRunTimeoutReturnsPartialResult(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_STDOUT", "partial")
	t.Setenv("FAKE_SLEEP_MS", "10000")
	// Generous relative to process startup (first exec of a freshly copied
	// binary can be slow), tiny relative to the sleep: no flakes either way.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res, err := Run(ctx, Request{Harness: ClaudeCode, Prompt: "hi", Workdir: t.TempDir()})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run error = %v, want context.DeadlineExceeded", err)
	}
	if res == nil {
		t.Fatal("Run returned nil Result on timeout; want the partial result")
	}
	if res.ExitCode != -1 {
		t.Errorf("ExitCode = %d, want -1 for a killed process", res.ExitCode)
	}
	if string(res.Stdout) != "partial" {
		t.Errorf("Stdout = %q, want the partial output", res.Stdout)
	}
}

func TestRunNotInstalled(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // empty dir: nothing resolvable
	_, err := Run(context.Background(), Request{Harness: ClaudeCode, Prompt: "hi", Workdir: t.TempDir()})
	var nie *NotInstalledError
	if !errors.As(err, &nie) {
		t.Fatalf("Run error = %v, want *NotInstalledError", err)
	}
	if nie.Harness != ClaudeCode || nie.Binary != "claude" {
		t.Errorf("NotInstalledError = %+v", nie)
	}
}

func TestRunStartFailure(t *testing.T) {
	fakeharness.Install(t, "claude")
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	res, err := Run(context.Background(), Request{Harness: ClaudeCode, Prompt: "hi", Workdir: missing})
	if err == nil || !strings.Contains(err.Error(), "starting") {
		t.Fatalf("Run error = %v, want a start failure", err)
	}
	if res != nil {
		t.Errorf("Run result = %+v, want nil when the process never started", res)
	}
}

func TestVersionUnknownHarness(t *testing.T) {
	_, err := Version(context.Background(), "gemini")
	var ire *InvalidRequestError
	if !errors.As(err, &ire) {
		t.Fatalf("Version error = %v, want *InvalidRequestError", err)
	}
}

func TestVersionNotInstalled(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // empty dir: nothing resolvable
	_, err := Version(context.Background(), ClaudeCode)
	var nie *NotInstalledError
	if !errors.As(err, &nie) {
		t.Fatalf("Version error = %v, want *NotInstalledError", err)
	}
}

func TestVersionCommandFails(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_EXIT", "1")
	if _, err := Version(context.Background(), ClaudeCode); err == nil {
		t.Fatal("Version accepted a failing version command")
	}
}

func TestVersionExtractsDotted(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_STDOUT", "2.1.204 (Claude Code)")
	v, err := Version(context.Background(), ClaudeCode)
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if v != "2.1.204" {
		t.Errorf("Version = %q, want 2.1.204", v)
	}
}

func TestVersionNoDottedVersion(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_STDOUT", "no version here")
	if _, err := Version(context.Background(), ClaudeCode); err == nil {
		t.Fatal("Version accepted output with no dotted version")
	}
}
