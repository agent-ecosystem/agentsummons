// Package fakeharness lets a test binary double as a fake agent harness:
// a package's TestMain hands control to Main when EnvEnable is set, and
// Install copies the running test binary into a temp PATH dir under a
// harness's binary name. Behavior is driven through FAKE_* environment
// variables, keeping process-handling tests hermetic and cross-platform
// (no shell scripts, no real harnesses). Test-only; never imported by
// non-test code.
package fakeharness

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// EnvEnable marks a process as running in fake-harness mode.
const EnvEnable = "AGENTSUMMONS_TEST_FAKE_HARNESS"

// Main is the fake harness's entry point. It never returns.
//
// FAKE_VERSION_STDOUT, when set and the process was invoked with exactly
// --version, is printed alone with an immediate zero exit — so version
// probes stay fast and distinct while FAKE_SLEEP_MS and FAKE_EXIT shape
// the main invocation. The remaining variables apply to every invocation:
// FAKE_STDOUT and FAKE_STDERR are echoed verbatim, FAKE_PRINT_CWD=1
// prints the working directory, FAKE_ECHO_ENV names a variable whose
// value to print, FAKE_SLEEP_MS delays exit, and FAKE_EXIT sets the code.
func Main() {
	if v := os.Getenv("FAKE_VERSION_STDOUT"); v != "" && len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Print(v)
		os.Exit(0)
	}
	fmt.Print(os.Getenv("FAKE_STDOUT"))
	fmt.Fprint(os.Stderr, os.Getenv("FAKE_STDERR"))
	if os.Getenv("FAKE_PRINT_CWD") == "1" {
		wd, err := os.Getwd()
		if err != nil {
			os.Exit(98)
		}
		fmt.Print(wd)
	}
	if key := os.Getenv("FAKE_ECHO_ENV"); key != "" {
		fmt.Print(os.Getenv(key))
	}
	if ms, _ := strconv.Atoi(os.Getenv("FAKE_SLEEP_MS")); ms > 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
	code, _ := strconv.Atoi(os.Getenv("FAKE_EXIT"))
	os.Exit(code)
}

// Install copies the running test binary into a fresh dir under the given
// binary name, points PATH at only that dir, and enables fake mode for
// child processes.
func Install(t testing.TB, name string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatalf("reading test binary: %v", err)
	}
	dir := t.TempDir()
	dst := filepath.Join(dir, name)
	if runtime.GOOS == "windows" {
		dst += ".exe"
	}
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatalf("installing fake harness: %v", err)
	}
	t.Setenv("PATH", dir)
	t.Setenv(EnvEnable, "1")
}
