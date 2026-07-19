"use strict";

// Node (platform, arch) pairs with a published platform package; mirrors
// the goreleaser build matrix. Alphabetical.
const SUPPORTED = new Set([
  "darwin-arm64",
  "darwin-x64",
  "linux-arm64",
  "linux-x64",
  "win32-arm64",
  "win32-x64",
]);

// binaryPath resolves the agentsummons binary: the AGENTSUMMONS_BINARY
// override first, then the platform package installed via
// optionalDependencies.
function binaryPath() {
  const override = process.env.AGENTSUMMONS_BINARY;
  if (override) return override;
  const key = `${process.platform}-${process.arch}`;
  if (!SUPPORTED.has(key)) {
    throw new Error(
      `agentsummons: no prebuilt binary for ${key}; install the Go CLI instead ` +
        "(https://github.com/agent-ecosystem/agentsummons) and set AGENTSUMMONS_BINARY",
    );
  }
  const exe = process.platform === "win32" ? "agentsummons.exe" : "agentsummons";
  try {
    return require.resolve(`agentsummons-${key}/bin/${exe}`);
  } catch {
    throw new Error(
      `agentsummons: platform package agentsummons-${key} is missing; it installs ` +
        "automatically as an optional dependency, so reinstall with optional " +
        "dependencies enabled, or set AGENTSUMMONS_BINARY to a binary you provide",
    );
  }
}

module.exports = { binaryPath };
