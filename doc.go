// Package agentsummons invokes agent harnesses (Antigravity CLI, Claude
// Code, Codex CLI) in headless mode, owning the per-harness flag knowledge
// so callers don't have to.
//
// agentsummons knows nothing about the transcripts a harness writes: its
// contract ends at "the command ran; here is everything observable about
// that". Finding and parsing session transcripts is the companion library
// agentminutes' job (github.com/agent-ecosystem/agentminutes) — agentsummons
// convenes the meeting, agentminutes takes the minutes.
package agentsummons
