#!/usr/bin/env python3
"""Fake agentsummons binary for wrapper tests; mirrors the npm wrapper's
fake and the Go suite's internal/fakeharness: FAKE_* env vars drive
behavior, and envelopes echo received argv so tests assert flag mapping."""

import json
import os
import sys

args = sys.argv[1:]
exit_code = int(os.environ.get("FAKE_EXIT", "0"))

if os.environ.get("FAKE_MODE") == "fail":
    sys.stderr.write(os.environ.get("FAKE_STDERR", "fake failure\n"))
    sys.exit(exit_code)

if args and args[0] == "run" and "--json" in args:
    timed_out = os.environ.get("FAKE_TIMED_OUT") == "1"
    json.dump(
        {
            "schema_version": 1,
            "harness": "claude-code",
            "argv": args,
            "prompt_index": len(args) - 1,
            "workdir": os.getcwd(),
            "start": "2026-01-01T00:00:00Z",
            "end": "2026-01-01T00:00:01Z",
            "exit_code": exit_code,
            "timed_out": timed_out,
            "stdout": os.environ.get("FAKE_STDOUT", ""),
            "stderr": os.environ.get("FAKE_STDERR", ""),
        },
        sys.stdout,
    )
    sys.exit(75 if timed_out else exit_code)

if args and args[0] == "build":
    json.dump(
        {
            "schema_version": 1,
            "harness": "claude-code",
            "argv": args,
            "prompt_index": len(args) - 1,
            "dir": os.getcwd(),
        },
        sys.stdout,
    )
    sys.exit(0)

sys.stdout.write(os.environ.get("FAKE_STDOUT", ""))
sys.stderr.write(os.environ.get("FAKE_STDERR", ""))
sys.exit(exit_code)
