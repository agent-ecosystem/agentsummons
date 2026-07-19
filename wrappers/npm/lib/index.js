"use strict";

const { execFile } = require("node:child_process");
const { binaryPath } = require("./binary");

// Harness stdout can be large; well past any real transcript-free output.
const MAX_BUFFER = 64 * 1024 * 1024;

// RequestError mirrors CLI exit 64: bad request or unsupported capability.
class RequestError extends Error {}
// NotInstalledError mirrors CLI exit 69: harness binary not found in PATH.
class NotInstalledError extends Error {}

// requestArgs maps the option object shared by run and build onto the
// CLI's request flags. The flag surface is the contract; anything
// unmodeled goes through extraArgs/extraEnv, same as the Go API.
function requestArgs(command, options) {
  const {
    harness,
    prompt,
    workdir,
    model,
    sessionId,
    resume,
    allowedTools,
    autoApprove,
    jsonOutput,
    extraArgs,
    extraEnv,
  } = options;
  if (!harness) throw new RequestError("agentsummons: options.harness is required");
  if (!prompt) throw new RequestError("agentsummons: options.prompt is required");
  const args = [command, "--harness", harness, "--prompt", prompt];
  if (workdir) args.push("--workdir", workdir);
  if (model) args.push("--model", model);
  if (sessionId) args.push("--session-id", sessionId);
  if (resume) args.push("--resume", resume);
  for (const tool of allowedTools ?? []) args.push("--allowed-tools", tool);
  if (autoApprove) args.push("--auto-approve");
  if (jsonOutput) args.push("--json-output");
  for (const arg of extraArgs ?? []) args.push("--extra-arg", arg);
  for (const kv of extraEnv ?? []) args.push("--extra-env", kv);
  return args;
}

// invoke runs the CLI and parses the single JSON envelope it prints.
// Envelope keys stay snake_case: they are the cross-language
// schema_version contract, not a JS-local shape.
function invoke(args) {
  return new Promise((resolve, reject) => {
    execFile(binaryPath(), args, { maxBuffer: MAX_BUFFER }, (err, stdout, stderr) => {
      // run --json writes its envelope even when the CLI exits nonzero
      // (harness failure, timeout), so a parseable envelope wins over the
      // exit code: the envelope's exit_code/timed_out fields carry it.
      let envelope;
      try {
        envelope = JSON.parse(stdout);
      } catch {
        envelope = null;
      }
      if (envelope && typeof envelope.schema_version === "number") {
        resolve(envelope);
        return;
      }
      const detail = (stderr || (err && err.message) || "no output").trim();
      const code = err && typeof err.code === "number" ? err.code : 0;
      if (code === 64) reject(new RequestError(detail));
      else if (code === 69) reject(new NotInstalledError(detail));
      else reject(new Error(`agentsummons: ${detail}`));
    });
  });
}

// run invokes a harness headlessly and resolves with the run envelope.
// A nonzero harness exit is data (envelope.exit_code), not a rejection;
// a timeout resolves with timed_out=true and the partial output.
// timeoutMs=0 disables the CLI's default 5-minute timeout.
function run(options) {
  try {
    const args = requestArgs("run", options);
    args.push("--json");
    if (options.timeoutMs !== undefined) args.push("--timeout", `${options.timeoutMs}ms`);
    return invoke(args);
  } catch (err) {
    return Promise.reject(err);
  }
}

// build resolves with the assembled argv/dir/env envelope, executing
// nothing.
function build(options) {
  try {
    return invoke(requestArgs("build", options));
  } catch (err) {
    return Promise.reject(err);
  }
}

module.exports = { run, build, binaryPath, RequestError, NotInstalledError };
