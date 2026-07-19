"""Wrapper tests against the fake binary; run from wrappers/pypi with
``python3 -m unittest discover -s tests``."""

import os
import sys
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))
import agentsummons  # noqa: E402

FAKE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "fake_agentsummons.py")


class WrapperTest(unittest.TestCase):
    def setUp(self):
        self._saved = os.environ.copy()
        os.environ["AGENTSUMMONS_BINARY"] = FAKE

    def tearDown(self):
        os.environ.clear()
        os.environ.update(self._saved)

    def test_binary_path_honors_override(self):
        self.assertEqual(agentsummons.binary_path(), FAKE)

    def test_run_maps_options_onto_cli_flags(self):
        envelope = agentsummons.run(
            harness="claude-code",
            prompt="hi",
            model="some-model",
            session_id="sid",
            allowed_tools=["Read", "Grep"],
            auto_approve=True,
            json_output=True,
            extra_args=["--flag"],
            extra_env=["K=V"],
            timeout=1.5,
        )
        self.assertEqual(envelope["schema_version"], 1)
        argv = envelope["argv"]
        self.assertEqual(argv[0], "run")
        for flag, value in [
            ("--harness", "claude-code"),
            ("--prompt", "hi"),
            ("--model", "some-model"),
            ("--session-id", "sid"),
            ("--allowed-tools", "Read"),
            ("--extra-arg", "--flag"),
            ("--extra-env", "K=V"),
            ("--timeout", "1.5s"),
        ]:
            self.assertIn(flag, argv)
            self.assertEqual(argv[argv.index(flag) + 1], value)
        self.assertIn("--auto-approve", argv)
        self.assertIn("--json-output", argv)
        self.assertIn("--json", argv)

    def test_run_returns_nonzero_exit_as_data(self):
        os.environ["FAKE_EXIT"] = "2"
        os.environ["FAKE_STDERR"] = "harness broke"
        envelope = agentsummons.run(harness="claude-code", prompt="hi")
        self.assertEqual(envelope["exit_code"], 2)
        self.assertEqual(envelope["stderr"], "harness broke")

    def test_run_returns_timeout_as_data(self):
        os.environ["FAKE_TIMED_OUT"] = "1"
        os.environ["FAKE_STDOUT"] = "partial"
        envelope = agentsummons.run(harness="claude-code", prompt="hi")
        self.assertTrue(envelope["timed_out"])
        self.assertEqual(envelope["stdout"], "partial")

    def test_run_raises_request_error_on_64(self):
        os.environ["FAKE_MODE"] = "fail"
        os.environ["FAKE_EXIT"] = "64"
        with self.assertRaises(agentsummons.RequestError):
            agentsummons.run(harness="nope", prompt="hi")

    def test_run_raises_not_installed_on_69(self):
        os.environ["FAKE_MODE"] = "fail"
        os.environ["FAKE_EXIT"] = "69"
        with self.assertRaises(agentsummons.NotInstalledError):
            agentsummons.run(harness="claude-code", prompt="hi")

    def test_run_rejects_missing_required_locally(self):
        with self.assertRaises(agentsummons.RequestError):
            agentsummons.run(harness="", prompt="hi")
        with self.assertRaises(agentsummons.RequestError):
            agentsummons.run(harness="claude-code", prompt="")

    def test_build_returns_envelope_without_json_flag(self):
        envelope = agentsummons.build(harness="claude-code", prompt="hi")
        self.assertEqual(envelope["schema_version"], 1)
        self.assertEqual(envelope["argv"][0], "build")
        self.assertNotIn("--json", envelope["argv"])


if __name__ == "__main__":
    unittest.main()
