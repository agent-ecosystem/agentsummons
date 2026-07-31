package agentsummons

import (
	"errors"
	"reflect"
	"testing"
)

func TestBuildGolden(t *testing.T) {
	cases := []struct {
		name string
		req  Request
		want []string
	}{
		{
			name: "antigravity minimal",
			req:  Request{Harness: Antigravity, Prompt: "hi", Workdir: "/w"},
			want: []string{"agy", "--add-dir", "/w", "-p", "hi"},
		},
		{
			name: "antigravity full",
			req: Request{
				Harness: Antigravity, Prompt: "hi", Workdir: "/w",
				AutoApprove: true, Model: "Gemini 3.5 Flash (Medium)",
				Resume: "conv-1", JSONOutput: true,
				ExtraArgs: []string{"--print-timeout", "20m"},
			},
			want: []string{
				"agy", "--dangerously-skip-permissions", "--add-dir", "/w",
				"--model", "Gemini 3.5 Flash (Medium)", "--conversation", "conv-1",
				"--output-format", "json", "--print-timeout", "20m", "-p", "hi",
			},
		},
		{
			name: "claude-code minimal",
			req:  Request{Harness: ClaudeCode, Prompt: "hi", Workdir: "/w"},
			want: []string{"claude", "-p", "hi"},
		},
		{
			name: "claude-code full",
			req: Request{
				Harness: ClaudeCode, Prompt: "hi", Workdir: "/w",
				AutoApprove: true, Model: "claude-sonnet-5", SessionID: "u-1",
				AllowedTools: []string{"Read", "Grep", "Glob"}, JSONOutput: true,
				ExtraArgs: []string{"--verbose"},
			},
			want: []string{
				"claude", "--dangerously-skip-permissions", "--model", "claude-sonnet-5",
				"--session-id", "u-1", "--allowedTools", "Read,Grep,Glob",
				"--output-format", "json", "--verbose", "-p", "hi",
			},
		},
		{
			name: "claude-code resume",
			req:  Request{Harness: ClaudeCode, Prompt: "hi", Workdir: "/w", Resume: "s-1"},
			want: []string{"claude", "--resume", "s-1", "-p", "hi"},
		},
		{
			name: "codex minimal",
			req:  Request{Harness: Codex, Prompt: "hi", Workdir: "/w"},
			want: []string{"codex", "exec", "--skip-git-repo-check", "hi"},
		},
		{
			name: "codex full",
			req: Request{
				Harness: Codex, Prompt: "hi", Workdir: "/w",
				AutoApprove: true, Model: "gpt-5.6-terra", JSONOutput: true,
				ExtraArgs: []string{"--output-last-message", "/tmp/last.txt"},
			},
			want: []string{
				"codex", "exec", "--skip-git-repo-check",
				"--dangerously-bypass-approvals-and-sandbox", "-m", "gpt-5.6-terra",
				"--json", "--output-last-message", "/tmp/last.txt", "hi",
			},
		},
		{
			name: "codex resume is a subcommand",
			req:  Request{Harness: Codex, Prompt: "hi", Workdir: "/w", Resume: "sess-1", Model: "gpt-5.6-terra"},
			want: []string{"codex", "exec", "resume", "--skip-git-repo-check", "-m", "gpt-5.6-terra", "sess-1", "hi"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			built, err := Build(tc.req)
			if err != nil {
				t.Fatalf("Build: %v", err)
			}
			if !reflect.DeepEqual(built.Argv, tc.want) {
				t.Errorf("Argv = %q, want %q", built.Argv, tc.want)
			}
			if built.PromptIndex != len(built.Argv)-1 {
				t.Errorf("PromptIndex = %d, want %d (prompt must be last)", built.PromptIndex, len(built.Argv)-1)
			}
			if got := built.Argv[built.PromptIndex]; got != tc.req.Prompt {
				t.Errorf("Argv[PromptIndex] = %q, want the prompt %q", got, tc.req.Prompt)
			}
			if built.Dir != tc.req.Workdir {
				t.Errorf("Dir = %q, want %q", built.Dir, tc.req.Workdir)
			}
		})
	}
}

func TestBuildCopiesExtraEnv(t *testing.T) {
	env := []string{"A=1", "B=2"}
	built, err := Build(Request{Harness: ClaudeCode, Prompt: "hi", Workdir: "/w", ExtraEnv: env})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !reflect.DeepEqual(built.EnvAdded, env) {
		t.Fatalf("EnvAdded = %q, want %q", built.EnvAdded, env)
	}
	env[0] = "A=mutated"
	if built.EnvAdded[0] != "A=1" {
		t.Error("EnvAdded aliases the caller's slice; want a copy")
	}
}

func TestBuildInvalidRequests(t *testing.T) {
	cases := []struct {
		name string
		req  Request
	}{
		{"unknown harness", Request{Harness: "gemini", Prompt: "hi", Workdir: "/w"}},
		{"empty prompt", Request{Harness: ClaudeCode, Workdir: "/w"}},
		{"empty workdir", Request{Harness: ClaudeCode, Prompt: "hi"}},
		{"session-id and resume together", Request{Harness: ClaudeCode, Prompt: "hi", Workdir: "/w", SessionID: "a", Resume: "b"}},
		{"extra env without equals", Request{Harness: ClaudeCode, Prompt: "hi", Workdir: "/w", ExtraEnv: []string{"NOEQUALS"}}},
		{"extra env with empty key", Request{Harness: ClaudeCode, Prompt: "hi", Workdir: "/w", ExtraEnv: []string{"=value"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Build(tc.req)
			var ire *InvalidRequestError
			if !errors.As(err, &ire) {
				t.Fatalf("Build error = %v, want *InvalidRequestError", err)
			}
		})
	}
}

func TestBuildUnsupportedOptions(t *testing.T) {
	cases := []struct {
		name   string
		req    Request
		option string
	}{
		{"antigravity session-id", Request{Harness: Antigravity, Prompt: "hi", Workdir: "/w", SessionID: "u"}, "SessionID"},
		{"antigravity allowed-tools", Request{Harness: Antigravity, Prompt: "hi", Workdir: "/w", AllowedTools: []string{"Read"}}, "AllowedTools"},
		{"codex session-id", Request{Harness: Codex, Prompt: "hi", Workdir: "/w", SessionID: "u"}, "SessionID"},
		{"codex allowed-tools", Request{Harness: Codex, Prompt: "hi", Workdir: "/w", AllowedTools: []string{"Read"}}, "AllowedTools"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Build(tc.req)
			var ue *UnsupportedError
			if !errors.As(err, &ue) {
				t.Fatalf("Build error = %v, want *UnsupportedError", err)
			}
			if ue.Option != tc.option {
				t.Errorf("Option = %q, want %q", ue.Option, tc.option)
			}
			if ue.Harness != tc.req.Harness {
				t.Errorf("Harness = %q, want %q", ue.Harness, tc.req.Harness)
			}
		})
	}
}
