#!/usr/bin/env node
"use strict";

// Fake agentsummons binary for wrapper tests, in the spirit of the Go
// suite's internal/fakeharness: behavior is driven by FAKE_* env vars, and
// envelopes echo received argv so tests can assert flag mapping.
const args = process.argv.slice(2);
const exit = Number(process.env.FAKE_EXIT || 0);

if (process.env.FAKE_MODE === "fail") {
  process.stderr.write(process.env.FAKE_STDERR || "fake failure\n");
  process.exit(exit);
}

if (args[0] === "run" && args.includes("--json")) {
  process.stdout.write(
    JSON.stringify({
      schema_version: 1,
      harness: "claude-code",
      argv: args,
      prompt_index: args.length - 1,
      workdir: process.cwd(),
      start: "2026-01-01T00:00:00Z",
      end: "2026-01-01T00:00:01Z",
      exit_code: exit,
      timed_out: process.env.FAKE_TIMED_OUT === "1",
      stdout: process.env.FAKE_STDOUT || "",
      stderr: process.env.FAKE_STDERR || "",
    }) + "\n",
  );
  process.exit(process.env.FAKE_TIMED_OUT === "1" ? 75 : exit);
}

if (args[0] === "build") {
  process.stdout.write(
    JSON.stringify({
      schema_version: 1,
      harness: "claude-code",
      argv: args,
      prompt_index: args.length - 1,
      dir: process.cwd(),
    }) + "\n",
  );
  process.exit(0);
}

// Passthrough mode: behave like a harness the shim wraps transparently.
process.stdout.write(process.env.FAKE_STDOUT || "");
process.stderr.write(process.env.FAKE_STDERR || "");
process.exit(exit);
