"""Python wrapper around the agentsummons Go CLI, which invokes agent
harnesses (Antigravity CLI, Claude Code, Codex CLI) headlessly.

The wheel bundles the real binary; ``run`` and ``build`` shell out to it
with ``--json`` and return the parsed envelope dict. Envelope keys stay
snake_case: they are the CLI's cross-language schema_version contract.
Set AGENTSUMMONS_BINARY to override which binary is invoked.
"""

import json
import os
import subprocess
import sys


class RequestError(RuntimeError):
    """CLI exit 64: bad request or unsupported capability."""


class NotInstalledError(RuntimeError):
    """CLI exit 69: harness binary not found in PATH (or no bundled binary)."""


def binary_path():
    """Return the path of the agentsummons binary this wrapper invokes."""
    override = os.environ.get("AGENTSUMMONS_BINARY")
    if override:
        return override
    exe = "agentsummons.exe" if sys.platform == "win32" else "agentsummons"
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "bin", exe)
    if not os.path.exists(path):
        raise NotInstalledError(
            "agentsummons: no bundled binary for this platform; install the Go "
            "CLI instead (https://github.com/agent-ecosystem/agentsummons) and "
            "set AGENTSUMMONS_BINARY"
        )
    # Some installers drop the exec bit on package data; restore best-effort.
    if os.name == "posix" and not os.access(path, os.X_OK):
        try:
            os.chmod(path, 0o755)
        except OSError:
            pass
    return path


def _request_args(
    command,
    harness,
    prompt,
    workdir=None,
    model=None,
    session_id=None,
    resume=None,
    allowed_tools=None,
    auto_approve=False,
    json_output=False,
    extra_args=None,
    extra_env=None,
):
    if not harness:
        raise RequestError("agentsummons: harness is required")
    if not prompt:
        raise RequestError("agentsummons: prompt is required")
    args = [command, "--harness", harness, "--prompt", prompt]
    if workdir:
        args += ["--workdir", workdir]
    if model:
        args += ["--model", model]
    if session_id:
        args += ["--session-id", session_id]
    if resume:
        args += ["--resume", resume]
    for tool in allowed_tools or []:
        args += ["--allowed-tools", tool]
    if auto_approve:
        args.append("--auto-approve")
    if json_output:
        args.append("--json-output")
    for arg in extra_args or []:
        args += ["--extra-arg", arg]
    for kv in extra_env or []:
        args += ["--extra-env", kv]
    return args


def _invoke(args):
    proc = subprocess.run(
        [binary_path()] + args, capture_output=True, text=True, check=False
    )
    # run --json writes its envelope even when the CLI exits nonzero
    # (harness failure, timeout), so a parseable envelope wins over the
    # exit code: its exit_code/timed_out fields carry the outcome.
    try:
        envelope = json.loads(proc.stdout)
    except ValueError:
        envelope = None
    if isinstance(envelope, dict) and "schema_version" in envelope:
        return envelope
    detail = (proc.stderr or "no output").strip()
    if proc.returncode == 64:
        raise RequestError(detail)
    if proc.returncode == 69:
        raise NotInstalledError(detail)
    raise RuntimeError("agentsummons: " + detail)


def run(
    harness,
    prompt,
    timeout=None,
    **kwargs,
):
    """Invoke a harness headlessly; returns the ``run --json`` envelope.

    A nonzero harness exit is data (``envelope["exit_code"]``), not an
    exception; a timeout returns ``timed_out: True`` with the partial
    output. ``timeout`` is seconds; 0 disables the CLI's 5-minute default.
    """
    args = _request_args("run", harness, prompt, **kwargs)
    args.append("--json")
    if timeout is not None:
        args += ["--timeout", "%gs" % timeout]
    return _invoke(args)


def build(harness, prompt, **kwargs):
    """Return the assembled argv/dir/env envelope without executing."""
    return _invoke(_request_args("build", harness, prompt, **kwargs))


def _main():
    """Console-script entry: transparent passthrough to the binary."""
    binary = binary_path()
    argv = [binary] + sys.argv[1:]
    if os.name == "posix":
        os.execv(binary, argv)
    raise SystemExit(subprocess.call(argv))
