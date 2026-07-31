package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/agent-ecosystem/agentsummons"
	"github.com/agent-ecosystem/agentsummons/internal/fakeharness"
)

func TestMain(m *testing.M) {
	if os.Getenv(fakeharness.EnvEnable) == "1" {
		fakeharness.Main()
		return
	}
	os.Exit(m.Run())
}

// execute runs the CLI in-process and reports what main would: the
// captured stdout/stderr and the process exit code.
func execute(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	root := newRootCmd()
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		var ee *exitError
		if errors.As(err, &ee) {
			return out.String(), errOut.String(), ee.code
		}
		return out.String(), errOut.String(), 1
	}
	return out.String(), errOut.String(), 0
}

// decodeJSON unmarshals one envelope into a map so tests assert the actual
// wire keys — the cross-language contract — not Go struct fields.
func decodeJSON(t *testing.T, s string) map[string]any {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("output is not a JSON envelope: %v\n%s", err, s)
	}
	return v
}

func TestBuildEnvelope(t *testing.T) {
	stdout, _, code := execute(t, "build",
		"--harness", "claude-code", "-p", "hi", "--workdir", "/w",
		"--extra-arg", "--verbose", "--extra-env", "A=1")
	if code != 0 {
		t.Fatalf("exit %d, stdout:\n%s", code, stdout)
	}
	envelope := decodeJSON(t, stdout)
	scalars := map[string]any{
		"schema_version": float64(schemaVersion),
		"harness":        "claude-code",
		"prompt_index":   float64(3),
		"dir":            "/w",
	}
	for key, want := range scalars {
		if envelope[key] != want {
			t.Errorf("%s = %v, want %v", key, envelope[key], want)
		}
	}
	wantArgv := []any{"claude", "--verbose", "-p", "hi"}
	if !reflect.DeepEqual(envelope["argv"], wantArgv) {
		t.Errorf("argv = %v, want %v (extra args before the prompt)", envelope["argv"], wantArgv)
	}
	if want := []any{"A=1"}; !reflect.DeepEqual(envelope["env_added"], want) {
		t.Errorf("env_added = %v, want %v", envelope["env_added"], want)
	}
}

func TestBuildDefaultsWorkdirToCwd(t *testing.T) {
	stdout, _, code := execute(t, "build", "--harness", "claude-code", "-p", "hi")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	if dir := decodeJSON(t, stdout)["dir"]; dir != wd {
		t.Errorf("dir = %v, want the current directory %q", dir, wd)
	}
}

func TestRequestErrorExitCodes(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		code       int
		stderrWant string
	}{
		{
			"unsupported capability",
			[]string{"build", "--harness", "antigravity", "-p", "hi", "--session-id", "s"},
			exitUsage, "does not support SessionID",
		},
		{
			"unknown harness",
			[]string{"build", "--harness", "gemini", "-p", "hi"},
			exitUsage, "unknown harness",
		},
		{
			"competing identity claims",
			[]string{"build", "--harness", "claude-code", "-p", "hi", "--session-id", "a", "--resume", "b"},
			exitUsage, "at most one",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, stderr, code := execute(t, tc.args...)
			if code != tc.code {
				t.Errorf("exit %d, want %d", code, tc.code)
			}
			if !strings.Contains(stderr, tc.stderrWant) {
				t.Errorf("stderr = %q, want it to contain %q", stderr, tc.stderrWant)
			}
		})
	}
}

func TestRunNotInstalledExit(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // empty dir: nothing resolvable
	_, stderr, code := execute(t, "run", "--harness", "claude-code", "-p", "hi", "--workdir", t.TempDir())
	if code != exitUnavailable {
		t.Errorf("exit %d, want %d", code, exitUnavailable)
	}
	if !strings.Contains(stderr, "not found in PATH") {
		t.Errorf("stderr = %q, want the not-installed message", stderr)
	}
}

func TestRunPassthrough(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_STDOUT", "the answer")
	t.Setenv("FAKE_STDERR", "some noise")
	t.Setenv("FAKE_EXIT", "3")
	stdout, stderr, code := execute(t, "run", "--harness", "claude-code", "-p", "hi", "--workdir", t.TempDir())
	if stdout != "the answer" {
		t.Errorf("stdout = %q, want the harness output passed through", stdout)
	}
	if stderr != "some noise" {
		t.Errorf("stderr = %q, want the harness stderr passed through", stderr)
	}
	if code != 3 {
		t.Errorf("exit %d, want the harness exit code 3 propagated", code)
	}
}

func TestRunTimeoutExit(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_SLEEP_MS", "10000")
	// 2s is generous relative to process startup but tiny relative to the
	// sleep: no flakes either way.
	_, stderr, code := execute(t, "run", "--harness", "claude-code", "-p", "hi",
		"--workdir", t.TempDir(), "--timeout", "2s")
	if code != exitTimeout {
		t.Errorf("exit %d, want %d", code, exitTimeout)
	}
	if !strings.Contains(stderr, "timeout") {
		t.Errorf("stderr = %q, want the timeout message", stderr)
	}
}

func TestRunJSONEnvelope(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_VERSION_STDOUT", "9999.0.0")
	t.Setenv("FAKE_STDOUT", "pong")
	stdout, _, code := execute(t, "run", "--harness", "claude-code", "-p", "hi",
		"--workdir", t.TempDir(), "--session-id", "s-1", "--json")
	if code != 0 {
		t.Fatalf("exit %d, stdout:\n%s", code, stdout)
	}
	envelope := decodeJSON(t, stdout)
	scalars := map[string]any{
		"schema_version":  float64(schemaVersion),
		"harness":         "claude-code",
		"harness_version": "9999.0.0",
		"exit_code":       float64(0),
		"stdout":          "pong",
		"session_id":      "s-1",
	}
	for key, want := range scalars {
		if envelope[key] != want {
			t.Errorf("%s = %v, want %v", key, envelope[key], want)
		}
	}
	if hint, _ := envelope["drift_hint"].(string); !strings.Contains(hint, "9999.0.0") {
		t.Errorf("drift_hint = %q, want it to flag the newer installed version", hint)
	}
	if _, ok := envelope["timed_out"]; ok {
		t.Error("timed_out present on a run that did not time out; want it omitted")
	}
	for _, key := range []string{"argv", "start", "end", "workdir"} {
		if _, ok := envelope[key]; !ok {
			t.Errorf("envelope missing %q", key)
		}
	}
}

func TestRunJSONTimedOut(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_VERSION_STDOUT", "1.0.0")
	t.Setenv("FAKE_STDOUT", "partial")
	t.Setenv("FAKE_SLEEP_MS", "10000")
	stdout, _, code := execute(t, "run", "--harness", "claude-code", "-p", "hi",
		"--workdir", t.TempDir(), "--timeout", "2s", "--json")
	if code != exitTimeout {
		t.Errorf("exit %d, want %d", code, exitTimeout)
	}
	envelope := decodeJSON(t, stdout)
	if envelope["timed_out"] != true {
		t.Errorf("timed_out = %v, want true", envelope["timed_out"])
	}
	if envelope["exit_code"] != float64(-1) {
		t.Errorf("exit_code = %v, want -1 for a killed harness", envelope["exit_code"])
	}
	if envelope["stdout"] != "partial" {
		t.Errorf("stdout = %v, want the partial output captured", envelope["stdout"])
	}
}

func TestDoctorStatuses(t *testing.T) {
	cases := []struct {
		name       string
		version    string // FAKE_STDOUT for the version probe
		code       int
		stdoutWant string
	}{
		{"clean", agentsummons.LastValidated[agentsummons.ClaudeCode], 0, "clean"},
		{"drift candidate", "9999.0.0", exitDoctorDrift, "drift candidate"},
		{"probe failure", "no version here", exitDoctorProbeFailed, "version probe failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeharness.Install(t, "claude")
			t.Setenv("FAKE_STDOUT", tc.version)
			stdout, _, code := execute(t, "doctor")
			if code != tc.code {
				t.Errorf("exit %d, want %d", code, tc.code)
			}
			if !strings.Contains(stdout, tc.stdoutWant) {
				t.Errorf("stdout = %q, want it to contain %q", stdout, tc.stdoutWant)
			}
			// Only claude is on PATH; the others must report not installed
			// without affecting the exit code.
			if !strings.Contains(stdout, "not installed") {
				t.Errorf("stdout = %q, want the uninstalled harnesses reported", stdout)
			}
		})
	}
}

func TestDoctorJSON(t *testing.T) {
	fakeharness.Install(t, "claude")
	t.Setenv("FAKE_STDOUT", "9999.0.0")
	stdout, _, code := execute(t, "doctor", "--json")
	if code != exitDoctorDrift {
		t.Errorf("exit %d, want %d", code, exitDoctorDrift)
	}
	envelope := decodeJSON(t, stdout)
	if envelope["schema_version"] != float64(schemaVersion) {
		t.Errorf("schema_version = %v, want %d", envelope["schema_version"], schemaVersion)
	}
	reports, _ := envelope["harnesses"].([]any)
	if len(reports) != len(agentsummons.Harnesses()) {
		t.Fatalf("harnesses has %d entries, want %d:\n%s", len(reports), len(agentsummons.Harnesses()), stdout)
	}
	statuses := map[string]string{}
	for _, r := range reports {
		report := r.(map[string]any)
		statuses[report["harness"].(string)], _ = report["status"].(string)
	}
	want := map[string]string{
		"antigravity": statusNotInstalled,
		"claude-code": statusDrift,
		"codex":       statusNotInstalled,
	}
	if !reflect.DeepEqual(statuses, want) {
		t.Errorf("statuses = %v, want %v", statuses, want)
	}
}

func TestInfoText(t *testing.T) {
	stdout, _, code := execute(t, "info")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	for _, want := range []string{"antigravity (agy)", "claude-code (claude)", "codex (codex)", "(unsupported)"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("info output missing %q", want)
		}
	}
}

func TestInfoJSON(t *testing.T) {
	stdout, _, code := execute(t, "info", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var envelope infoEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if envelope.SchemaVersion != schemaVersion {
		t.Errorf("schema_version = %d, want %d", envelope.SchemaVersion, schemaVersion)
	}
	if len(envelope.Harnesses) != len(agentsummons.Harnesses()) {
		t.Fatalf("harnesses has %d entries, want %d", len(envelope.Harnesses), len(agentsummons.Harnesses()))
	}
	for i, id := range agentsummons.Harnesses() {
		if envelope.Harnesses[i].Harness != id {
			t.Errorf("harnesses[%d] = %q, want %q (alphabetical)", i, envelope.Harnesses[i].Harness, id)
		}
	}
}

func TestInfoSingleHarness(t *testing.T) {
	stdout, _, code := execute(t, "info", "--harness", "codex", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var envelope infoEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if len(envelope.Harnesses) != 1 || envelope.Harnesses[0].Harness != agentsummons.Codex {
		t.Errorf("harnesses = %+v, want just codex", envelope.Harnesses)
	}
}

func TestInfoUnknownHarnessExit(t *testing.T) {
	_, stderr, code := execute(t, "info", "--harness", "gemini")
	if code != exitUsage {
		t.Errorf("exit %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr, "unknown harness") {
		t.Errorf("stderr = %q, want the unknown-harness message", stderr)
	}
}

func TestDriftHint(t *testing.T) {
	validated := agentsummons.LastValidated[agentsummons.ClaudeCode]
	cases := []struct {
		name      string
		installed string
		wantHint  bool
	}{
		{"no installed version", "", false},
		{"older than validated", "0.0.1", false},
		{"equal to validated", validated, false},
		{"newer than validated", "9999.0.0", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hint := driftHint(agentsummons.ClaudeCode, tc.installed)
			if got := hint != ""; got != tc.wantHint {
				t.Errorf("driftHint(%q) = %q, want hint: %v", tc.installed, hint, tc.wantHint)
			}
			if tc.wantHint && !strings.Contains(hint, tc.installed) {
				t.Errorf("driftHint = %q, want it to name the installed version", hint)
			}
		})
	}
}

func TestExitErrorString(t *testing.T) {
	if got, want := (&exitError{code: 70}).Error(), "exit code 70"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestHarnessHelpList(t *testing.T) {
	if got, want := harnessHelpList(), `"antigravity", "claude-code", or "codex"`; got != want {
		t.Errorf("harnessHelpList() = %s, want %s", got, want)
	}
}
