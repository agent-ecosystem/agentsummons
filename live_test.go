package agentsummons

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Live tests invoke the real installed harnesses and spend real tokens.
// Gate: AGENTSUMMONS_LIVE=all, or a comma-separated ID list, e.g.
//
//	AGENTSUMMONS_LIVE=claude-code go test -run TestLive -v
//
// They validate the flag knowledge in harnesses.go against the installed
// releases — run them (and update LastValidated) when regrounding the spec
// table per DEVELOPMENT.md.
func liveGate(t *testing.T, id ID) {
	t.Helper()
	gate := os.Getenv("AGENTSUMMONS_LIVE")
	if gate == "" {
		t.Skip("live test: set AGENTSUMMONS_LIVE=all or a harness list to enable (spends tokens)")
	}
	enabled := gate == "all"
	for _, part := range strings.Split(gate, ",") {
		if ID(strings.TrimSpace(part)) == id {
			enabled = true
		}
	}
	if !enabled {
		t.Skipf("live test: %q not in AGENTSUMMONS_LIVE=%q", id, gate)
	}
	if _, err := Version(context.Background(), id); err != nil {
		t.Skipf("live test: %v", err)
	}
}

func liveCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	t.Cleanup(cancel)
	return ctx
}

func newUUID(t *testing.T) string {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("uuid: %v", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func TestLiveAntigravity(t *testing.T) {
	liveGate(t, Antigravity)
	res, err := Run(liveCtx(t), Request{
		Harness: Antigravity, Prompt: "Reply with exactly: pong", Workdir: t.TempDir(),
		AutoApprove: true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit %d, stderr: %s", res.ExitCode, res.Stderr)
	}
	// agy text-mode stdout under a non-TTY is best-effort (see
	// Capabilities notes); a clean exit is the strongest assertion
	// available here.
}

// TestLiveAntigravityResume validates the in-band multi-turn loop that
// landed in agy 1.1.8: turn 1's JSON envelope carries conversation_id,
// turn 2 resumes with --conversation and echoes the same id (append, not
// fork). Envelope parsing is deliberately caller-side: the library never
// interprets harness output.
func TestLiveAntigravityResume(t *testing.T) {
	liveGate(t, Antigravity)
	wd := t.TempDir()
	res, err := Run(liveCtx(t), Request{
		Harness: Antigravity, Prompt: "Reply with exactly: pong", Workdir: wd,
		JSONOutput: true, AutoApprove: true,
	})
	if err != nil {
		t.Fatalf("turn 1: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("turn 1 exit %d, stderr: %s", res.ExitCode, res.Stderr)
	}
	var envelope struct {
		ConversationID string `json:"conversation_id"`
		Response       string `json:"response"`
	}
	if err := json.Unmarshal(res.Stdout, &envelope); err != nil {
		t.Fatalf("turn 1 stdout is not a JSON envelope: %v\n%s", err, res.Stdout)
	}
	if envelope.ConversationID == "" {
		t.Fatalf("turn 1 envelope has no conversation_id (agy < 1.1.8?):\n%s", res.Stdout)
	}
	if !strings.Contains(strings.ToLower(envelope.Response), "pong") {
		t.Errorf("turn 1 response = %q, want it to contain pong", envelope.Response)
	}

	res2, err := Run(liveCtx(t), Request{
		Harness: Antigravity, Prompt: "Reply with exactly the same word you replied with before.",
		Workdir: wd, Resume: envelope.ConversationID, JSONOutput: true, AutoApprove: true,
	})
	if err != nil {
		t.Fatalf("turn 2 (resume %s): %v", envelope.ConversationID, err)
	}
	if res2.ExitCode != 0 {
		t.Fatalf("turn 2 exit %d, stderr: %s", res2.ExitCode, res2.Stderr)
	}
	var envelope2 struct {
		ConversationID string `json:"conversation_id"`
		Response       string `json:"response"`
	}
	if err := json.Unmarshal(res2.Stdout, &envelope2); err != nil {
		t.Fatalf("turn 2 stdout is not a JSON envelope: %v\n%s", err, res2.Stdout)
	}
	if envelope2.ConversationID != envelope.ConversationID {
		t.Errorf("turn 2 conversation_id = %q, want %q (resume should append, not fork)", envelope2.ConversationID, envelope.ConversationID)
	}
	if !strings.Contains(strings.ToLower(envelope2.Response), "pong") {
		t.Errorf("turn 2 response = %q, want it to recall pong (context carried across resume)", envelope2.Response)
	}
	t.Logf("resume conversation_id: turn1=%s turn2=%s", envelope.ConversationID, envelope2.ConversationID)
}

// TestLiveClaudeCode validates the full in-band multi-turn loop: preset
// session ID, JSON envelope echo, then resume for a second turn.
func TestLiveClaudeCode(t *testing.T) {
	liveGate(t, ClaudeCode)
	wd := t.TempDir()
	sid := newUUID(t)
	res, err := Run(liveCtx(t), Request{
		Harness: ClaudeCode, Prompt: "Reply with exactly: pong", Workdir: wd,
		SessionID: sid, JSONOutput: true, AutoApprove: true,
	})
	if err != nil {
		t.Fatalf("turn 1: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("turn 1 exit %d, stderr: %s", res.ExitCode, res.Stderr)
	}
	var envelope struct {
		Result    string `json:"result"`
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(res.Stdout, &envelope); err != nil {
		t.Fatalf("turn 1 stdout is not a JSON envelope: %v\n%s", err, res.Stdout)
	}
	if envelope.SessionID != sid {
		t.Errorf("envelope session_id = %q, want preset %q", envelope.SessionID, sid)
	}
	if !strings.Contains(strings.ToLower(envelope.Result), "pong") {
		t.Errorf("turn 1 result = %q, want it to contain pong", envelope.Result)
	}

	res2, err := Run(liveCtx(t), Request{
		Harness: ClaudeCode, Prompt: "Reply with exactly the same word you replied with before.",
		Workdir: wd, Resume: envelope.SessionID, JSONOutput: true, AutoApprove: true,
	})
	if err != nil {
		t.Fatalf("turn 2 (resume): %v", err)
	}
	if res2.ExitCode != 0 {
		t.Fatalf("turn 2 exit %d, stderr: %s", res2.ExitCode, res2.Stderr)
	}
	var envelope2 struct {
		Result    string `json:"result"`
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(res2.Stdout, &envelope2); err != nil {
		t.Fatalf("turn 2 stdout is not a JSON envelope: %v\n%s", err, res2.Stdout)
	}
	if !strings.Contains(strings.ToLower(envelope2.Result), "pong") {
		t.Errorf("turn 2 result = %q, want it to recall pong (context carried across resume)", envelope2.Result)
	}
	t.Logf("resume session_id: turn1=%s turn2=%s (chaining rule: use each turn's envelope)", envelope.SessionID, envelope2.SessionID)
}

func TestLiveCodex(t *testing.T) {
	liveGate(t, Codex)
	res, err := Run(liveCtx(t), Request{
		Harness: Codex, Prompt: "Reply with exactly: pong", Workdir: t.TempDir(),
		AutoApprove: true,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit %d, stderr: %s", res.ExitCode, res.Stderr)
	}
	if len(res.Stdout) == 0 {
		t.Error("empty stdout")
	}
}

// TestLiveCodexResume validates the exec resume subcommand shape: the
// session ref comes in-band from turn 1's JSONL event stream, and turn 2's
// answer from --output-last-message (the reliable capture per the manifest
// notes). The stream parsing here is deliberately caller-side: the library
// never interprets harness output.
func TestLiveCodexResume(t *testing.T) {
	liveGate(t, Codex)
	wd := t.TempDir()
	res, err := Run(liveCtx(t), Request{
		Harness: Codex, Prompt: "Reply with exactly: pong", Workdir: wd,
		JSONOutput: true, AutoApprove: true,
	})
	if err != nil {
		t.Fatalf("turn 1: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("turn 1 exit %d, stderr: %s", res.ExitCode, res.Stderr)
	}
	ref := codexSessionRef(res.Stdout)
	if ref == "" {
		t.Fatalf("no session/thread/rollout id in turn 1's JSONL stream:\n%s", res.Stdout)
	}

	lastMessage := filepath.Join(t.TempDir(), "last.txt")
	res2, err := Run(liveCtx(t), Request{
		Harness: Codex, Prompt: "Reply with exactly the same word you replied with before.",
		Workdir: wd, Resume: ref, AutoApprove: true,
		ExtraArgs: []string{"--output-last-message", lastMessage},
	})
	if err != nil {
		t.Fatalf("turn 2 (resume %s): %v", ref, err)
	}
	if res2.ExitCode != 0 {
		t.Fatalf("turn 2 exit %d, stderr: %s", res2.ExitCode, res2.Stderr)
	}
	answer, err := os.ReadFile(lastMessage)
	if err != nil {
		t.Fatalf("reading --output-last-message file: %v", err)
	}
	if !strings.Contains(strings.ToLower(string(answer)), "pong") {
		t.Errorf("turn 2 answer = %q, want it to recall pong (context carried across resume)", answer)
	}
	t.Logf("resume ref: %s; turn 2 answer: %s", ref, bytes.TrimSpace(answer))
}

// codexSessionRef leniently scans a JSONL event stream for the first
// session/thread/rollout id, wherever it nests.
func codexSessionRef(stream []byte) string {
	for _, line := range bytes.Split(stream, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var v any
		if json.Unmarshal(line, &v) != nil {
			continue
		}
		if id := findSessionKey(v); id != "" {
			return id
		}
	}
	return ""
}

func findSessionKey(v any) string {
	switch vv := v.(type) {
	case map[string]any:
		for _, key := range []string{"session_id", "thread_id", "rollout_id"} {
			if s, ok := vv[key].(string); ok && s != "" {
				return s
			}
		}
		for _, child := range vv {
			if id := findSessionKey(child); id != "" {
				return id
			}
		}
	case []any:
		for _, child := range vv {
			if id := findSessionKey(child); id != "" {
				return id
			}
		}
	}
	return ""
}
