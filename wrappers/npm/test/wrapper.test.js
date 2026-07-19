"use strict";

const assert = require("node:assert/strict");
const { execFile } = require("node:child_process");
const path = require("node:path");
const { test } = require("node:test");

const { run, build, binaryPath, RequestError, NotInstalledError } = require("../lib/index");

const FAKE = path.join(__dirname, "fake-agentsummons.js");
const SHIM = path.join(__dirname, "..", "bin", "agentsummons.js");

function withFake(env, fn) {
  const saved = {};
  const entries = { AGENTSUMMONS_BINARY: FAKE, ...env };
  for (const [k, v] of Object.entries(entries)) {
    saved[k] = process.env[k];
    if (v === undefined) delete process.env[k];
    else process.env[k] = v;
  }
  const restore = () => {
    for (const [k, v] of Object.entries(saved)) {
      if (v === undefined) delete process.env[k];
      else process.env[k] = v;
    }
  };
  return fn().finally(restore);
}

test("binaryPath honors AGENTSUMMONS_BINARY", () =>
  withFake({}, async () => {
    assert.equal(binaryPath(), FAKE);
  }));

test("shim passes stdio through and propagates the exit code", () => {
  return new Promise((resolve, reject) => {
    execFile(
      process.execPath,
      [SHIM, "run", "--harness", "claude-code", "-p", "hi"],
      {
        env: {
          ...process.env,
          AGENTSUMMONS_BINARY: FAKE,
          FAKE_STDOUT: "out",
          FAKE_STDERR: "err",
          FAKE_EXIT: "3",
        },
      },
      (err, stdout, stderr) => {
        try {
          assert.equal(err && err.code, 3);
          assert.equal(stdout, "out");
          assert.equal(stderr, "err");
          resolve();
        } catch (assertion) {
          reject(assertion);
        }
      },
    );
  });
});

test("run maps options onto CLI flags and returns the envelope", () =>
  withFake({}, async () => {
    const envelope = await run({
      harness: "claude-code",
      prompt: "hi",
      model: "some-model",
      sessionId: "sid",
      allowedTools: ["Read", "Grep"],
      autoApprove: true,
      jsonOutput: true,
      extraArgs: ["--flag"],
      extraEnv: ["K=V"],
      timeoutMs: 1500,
    });
    assert.equal(envelope.schema_version, 1);
    const argv = envelope.argv;
    assert.equal(argv[0], "run");
    for (const expected of [
      ["--harness", "claude-code"],
      ["--prompt", "hi"],
      ["--model", "some-model"],
      ["--session-id", "sid"],
      ["--allowed-tools", "Read"],
      ["--allowed-tools", "Grep"],
      ["--extra-arg", "--flag"],
      ["--extra-env", "K=V"],
      ["--timeout", "1500ms"],
    ]) {
      const i = argv.indexOf(expected[0]);
      assert.notEqual(i, -1, `missing ${expected[0]}`);
      assert.ok(argv.includes(expected[1], i), `missing value for ${expected[0]}`);
    }
    assert.ok(argv.includes("--auto-approve"));
    assert.ok(argv.includes("--json-output"));
    assert.ok(argv.includes("--json"));
  }));

test("run returns nonzero harness exits as envelope data", () =>
  withFake({ FAKE_EXIT: "2", FAKE_STDERR: "harness broke" }, async () => {
    const envelope = await run({ harness: "claude-code", prompt: "hi" });
    assert.equal(envelope.exit_code, 2);
    assert.equal(envelope.stderr, "harness broke");
  }));

test("run returns timeouts as envelope data", () =>
  withFake({ FAKE_TIMED_OUT: "1", FAKE_STDOUT: "partial" }, async () => {
    const envelope = await run({ harness: "claude-code", prompt: "hi" });
    assert.equal(envelope.timed_out, true);
    assert.equal(envelope.stdout, "partial");
  }));

test("run rejects with RequestError on exit 64", () =>
  withFake({ FAKE_MODE: "fail", FAKE_EXIT: "64", FAKE_STDERR: "bad request" }, async () => {
    await assert.rejects(run({ harness: "nope", prompt: "hi" }), RequestError);
  }));

test("run rejects with NotInstalledError on exit 69", () =>
  withFake({ FAKE_MODE: "fail", FAKE_EXIT: "69", FAKE_STDERR: "not installed" }, async () => {
    await assert.rejects(run({ harness: "claude-code", prompt: "hi" }), NotInstalledError);
  }));

test("run rejects locally on a missing required option", async () => {
  await assert.rejects(run({ prompt: "hi" }), RequestError);
  await assert.rejects(run({ harness: "claude-code" }), RequestError);
});

test("build returns the assembled envelope without executing", () =>
  withFake({}, async () => {
    const envelope = await build({ harness: "claude-code", prompt: "hi" });
    assert.equal(envelope.schema_version, 1);
    assert.equal(envelope.argv[0], "build");
    assert.ok(!envelope.argv.includes("--json"), "build has no --json flag");
  }));
