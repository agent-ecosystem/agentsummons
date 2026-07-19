package agentsummons

import (
	"errors"
	"sort"
	"testing"
)

func TestHarnessesAlphabetical(t *testing.T) {
	ids := Harnesses()
	if !sort.SliceIsSorted(ids, func(i, j int) bool { return ids[i] < ids[j] }) {
		t.Errorf("Harnesses() = %v, want alphabetical", ids)
	}
}

// TestTablesPaired enforces that the spec table and the flag-surface
// validation table cover exactly the supported harnesses: knowledge and
// its validation record move together.
func TestTablesPaired(t *testing.T) {
	for _, id := range Harnesses() {
		if _, ok := specs[id]; !ok {
			t.Errorf("specs missing %q", id)
		}
		if _, ok := LastValidated[id]; !ok {
			t.Errorf("LastValidated missing %q", id)
		}
	}
	if len(specs) != len(Harnesses()) {
		t.Errorf("specs has %d entries, want %d", len(specs), len(Harnesses()))
	}
	if len(LastValidated) != len(Harnesses()) {
		t.Errorf("LastValidated has %d entries, want %d", len(LastValidated), len(Harnesses()))
	}
}

func TestInfoForUnknownHarness(t *testing.T) {
	_, err := InfoFor("gemini")
	var ire *InvalidRequestError
	if !errors.As(err, &ire) {
		t.Fatalf("InfoFor error = %v, want *InvalidRequestError", err)
	}
}

func TestCapabilitiesComplete(t *testing.T) {
	for _, id := range Harnesses() {
		caps, err := InfoFor(id)
		if err != nil {
			t.Fatalf("InfoFor(%q): %v", id, err)
		}
		if caps.Harness != id {
			t.Errorf("%s: caps.Harness = %q", id, caps.Harness)
		}
		if caps.Binary == "" || len(caps.VersionArgs) == 0 {
			t.Errorf("%s: Binary/VersionArgs must be set", id)
		}
		// Universally supported surfaces must always be documented.
		if caps.Prompt == "" || caps.Workdir == "" || caps.AutoApprove == "" || caps.Model == "" || caps.Resume == "" {
			t.Errorf("%s: Prompt, Workdir, AutoApprove, Model, and Resume descriptions must be non-empty", id)
		}
		if (caps.JSONOutput == "") != (caps.JSONOutputShape == "") {
			t.Errorf("%s: JSONOutput and JSONOutputShape must be set together", id)
		}
	}
}

// TestInfoForCopies enforces that InfoFor hands out the caller's own copy:
// mutating a returned manifest's slices must never corrupt the spec table.
func TestInfoForCopies(t *testing.T) {
	for _, id := range Harnesses() {
		caps, err := InfoFor(id)
		if err != nil {
			t.Fatalf("InfoFor(%q): %v", id, err)
		}
		caps.VersionArgs[0] = "mutated"
		if len(caps.BaseArgs) > 0 {
			caps.BaseArgs[0] = "mutated"
		}
		if len(caps.Notes) > 0 {
			caps.Notes[0] = "mutated"
		}
		fresh, err := InfoFor(id)
		if err != nil {
			t.Fatalf("InfoFor(%q): %v", id, err)
		}
		if fresh.VersionArgs[0] == "mutated" {
			t.Errorf("%s: VersionArgs aliases the spec table; want a copy", id)
		}
		if len(fresh.BaseArgs) > 0 && fresh.BaseArgs[0] == "mutated" {
			t.Errorf("%s: BaseArgs aliases the spec table; want a copy", id)
		}
		if len(fresh.Notes) > 0 && fresh.Notes[0] == "mutated" {
			t.Errorf("%s: Notes aliases the spec table; want a copy", id)
		}
	}
}

// TestBaseArgsMatchAssembly enforces that the descriptive manifest agrees
// with what assemble actually emits: the minimal invocation's argv is the
// binary followed immediately by BaseArgs.
func TestBaseArgsMatchAssembly(t *testing.T) {
	for _, id := range Harnesses() {
		caps, err := InfoFor(id)
		if err != nil {
			t.Fatalf("InfoFor(%q): %v", id, err)
		}
		built, err := Build(Request{Harness: id, Prompt: "p", Workdir: "/w"})
		if err != nil {
			t.Fatalf("Build(%q minimal): %v", id, err)
		}
		if built.Argv[0] != caps.Binary {
			t.Errorf("%s: Argv[0] = %q, want the manifest Binary %q", id, built.Argv[0], caps.Binary)
		}
		for i, arg := range caps.BaseArgs {
			if 1+i >= len(built.Argv) || built.Argv[1+i] != arg {
				t.Errorf("%s: Argv = %q, want BaseArgs %q right after the binary", id, built.Argv, caps.BaseArgs)
				break
			}
		}
	}
}

// TestUnsupportedMatchesManifest enforces the contract that a Request
// field is accepted exactly when the corresponding Capabilities field is
// non-empty — the manifest and the validator can never disagree.
func TestUnsupportedMatchesManifest(t *testing.T) {
	for _, id := range Harnesses() {
		caps, err := InfoFor(id)
		if err != nil {
			t.Fatalf("InfoFor(%q): %v", id, err)
		}
		base := Request{Harness: id, Prompt: "p", Workdir: "/w"}
		cases := []struct {
			option    string
			set       func(*Request)
			supported bool
		}{
			{"Model", func(r *Request) { r.Model = "m" }, caps.Model != ""},
			{"SessionID", func(r *Request) { r.SessionID = "s" }, caps.SessionID != ""},
			{"Resume", func(r *Request) { r.Resume = "r" }, caps.Resume != ""},
			{"AllowedTools", func(r *Request) { r.AllowedTools = []string{"Read"} }, caps.AllowedTools != ""},
			{"JSONOutput", func(r *Request) { r.JSONOutput = true }, caps.JSONOutput != ""},
			{"AutoApprove", func(r *Request) { r.AutoApprove = true }, caps.AutoApprove != ""},
		}
		for _, tc := range cases {
			req := base
			tc.set(&req)
			_, err := Build(req)
			if tc.supported && err != nil {
				t.Errorf("%s: %s is in the manifest but Build rejected it: %v", id, tc.option, err)
			}
			if !tc.supported && err == nil {
				t.Errorf("%s: %s is not in the manifest but Build accepted it", id, tc.option)
			}
		}
	}
}
