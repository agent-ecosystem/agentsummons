"use strict";

// Node (platform, arch) pairs with a published platform package; mirrors
// PLATFORMS in scripts/build-packages.mjs. Alphabetical. win32-x64 is
// temporarily absent while npm blocks the package name (see the PLATFORMS
// comment); the goreleaser release still builds that binary.
const SUPPORTED = new Set([
  "darwin-arm64",
  "darwin-x64",
  "linux-arm64",
  "linux-x64",
  "win32-arm64",
]);

// binaryPath resolves the agentsummons binary: the AGENTSUMMONS_BINARY
// override first, then the platform package installed via
// optionalDependencies.
function binaryPath() {
  const override = process.env.AGENTSUMMONS_BINARY;
  if (override) return override;
  const key = `${process.platform}-${process.arch}`;
  if (!SUPPORTED.has(key)) {
    const hint =
      key === "win32-x64"
        ? "the npm platform package is temporarily unavailable; download the Windows " +
          "binary from https://github.com/agent-ecosystem/agentsummons/releases and set " +
          "AGENTSUMMONS_BINARY, or use the PyPI package"
        : "install the Go CLI instead (https://github.com/agent-ecosystem/agentsummons) " +
          "and set AGENTSUMMONS_BINARY";
    throw new Error(`agentsummons: no prebuilt binary for ${key}; ${hint}`);
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
