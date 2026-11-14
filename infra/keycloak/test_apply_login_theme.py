"""Behavior tests for the container setup script, without Docker or credentials."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("apply-login-theme.sh")


class ApplyLoginThemeTest(unittest.TestCase):
    def run_setup(self, current="keycloak", failure=""):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "state").write_text(current)
            cli = root / "kcadm.sh"
            cli.write_text("""#!/usr/bin/env python3
import os, pathlib, sys, time
root = pathlib.Path(os.environ['TEST_ROOT'])
action = sys.argv[1]
with (root / 'calls').open('a') as log:
    log.write(action + '\\n')
if action == os.environ.get('TEST_FAILURE'):
    print('sensitive-cli-diagnostic', file=sys.stderr)
    sys.exit(1)
if action == 'config':
    config = pathlib.Path(sys.argv[sys.argv.index('--config') + 1])
    config.write_text('test-session')
    (root / 'session-path').write_text(str(config))
elif action == 'get':
    if os.environ.get('TEST_FAILURE') == 'timeout':
        time.sleep(20)
    print((root / 'state').read_text())
elif action == 'update':
    assert '--no-merge' in sys.argv
    assert sys.argv[2] == 'realms/guild-banker'
    assert sys.argv[sys.argv.index('-s') + 1] == 'loginTheme=guild-banker'
    if os.environ.get('TEST_FAILURE') != 'confirmation':
        (root / 'state').write_text('guild-banker')
""")
            cli.chmod(0o700)
            if failure == "timeout":
                timer = root / "timeout"
                timer.write_text("""#!/usr/bin/env python3
import os, sys
assert '60s' in sys.argv
args = ['0.4s' if arg == '60s' else arg for arg in sys.argv[1:]]
os.execv('/usr/bin/timeout', ['timeout', *args])
""")
                timer.chmod(0o700)
            result = subprocess.run(
                ["bash", str(SCRIPT)], capture_output=True, text=True, timeout=10,
                env={**os.environ, "PATH": directory + ":" + os.environ["PATH"],
                     "TEST_ROOT": directory, "TEST_FAILURE": failure,
                     "KC_BOOTSTRAP_ADMIN_USERNAME": "admin",
                     "KC_BOOTSTRAP_ADMIN_PASSWORD": "test-secret-never-print"},
            )
            calls = (root / "calls").read_text().splitlines()
            if (root / "session-path").exists():
                self.assertFalse(Path((root / "session-path").read_text()).exists())
            self.assertNotIn("test-secret-never-print", result.stdout + result.stderr)
            self.assertNotIn("sensitive-cli-diagnostic", result.stdout + result.stderr)
            return result, calls

    def test_selects_only_theme_and_confirms(self):
        result, calls = self.run_setup()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["config", "get", "update", "get"])
        self.assertIn("selected and confirmed", result.stdout)

    def test_already_selected_does_not_write(self):
        result, calls = self.run_setup(current="guild-banker")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["config", "get"])
        self.assertIn("already selected", result.stdout)

    def test_admin_read_and_update_failures_are_explicit(self):
        for failure in ["config", "get", "update"]:
            with self.subTest(failure=failure):
                result, calls = self.run_setup(failure=failure)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("unable to", result.stderr)
                self.assertEqual(calls[-1], failure)

    def test_failed_confirmation_is_not_success(self):
        result, _ = self.run_setup(failure="confirmation")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("could not be confirmed", result.stderr)

    def test_timeout_exits_and_removes_temporary_session(self):
        result, calls = self.run_setup(failure="timeout")
        self.assertEqual(result.returncode, 124)
        self.assertEqual(calls, ["config", "get"])


if __name__ == "__main__":
    unittest.main()
