#!/usr/bin/env bats

load test_helper/bats-support/load
load test_helper/bats-assert/load

setup() {
  TEMP_DIR="$(mktemp -d)"
  export TEMP_DIR
  CODEX_CONFIG="$TEMP_DIR/config.toml"
  export CODEX_CONFIG
}

teardown() {
  rm -rf "$TEMP_DIR"
}

@test "Codex config: notify should be string format, not array" {
  # Simulate the Python script that writes Codex config
  python3 - "$CODEX_CONFIG" "1" << 'PYEOF'
import sys
config_path = sys.argv[1]
sound = int(sys.argv[2])

with open(config_path, "w") as f:
    if sound:
        f.write('notify = ["afplay", "/System/Library/Sounds/Bottle.aiff"]\n')
PYEOF

  # Verify the format is correct (should be string, not array)
  run cat "$CODEX_CONFIG"

  # This test should FAIL with current code
  # Current: notify = ["afplay", "/System/Library/Sounds/Bottle.aiff"]
  # Expected: notify = "afplay /System/Library/Sounds/Bottle.aiff"
  refute_line --partial '["afplay"'
  assert_line --partial 'notify = "afplay'
}

@test "Codex config: notify with custom script should be string format" {
  # Simulate setting notify with a bash script (as user has in their config)
  python3 - "$CODEX_CONFIG" << 'PYEOF'
import sys
config_path = sys.argv[1]

with open(config_path, "w") as f:
    f.write('notify = ["bash", "~/.config/ghost-tab/codex-notify.sh"]\n')
PYEOF

  # Verify the format is correct
  run cat "$CODEX_CONFIG"

  # This test should FAIL with current user config
  # Current: notify = ["bash", "~/.config/ghost-tab/codex-notify.sh"]
  # Expected: notify = "bash ~/.config/ghost-tab/codex-notify.sh"
  refute_line --partial '["bash"'
  assert_line --partial 'notify = "bash'
}

@test "Codex config: verify string format is valid TOML" {
  # Write correct format
  cat > "$CODEX_CONFIG" << 'EOF'
notify = "afplay /System/Library/Sounds/Bottle.aiff"
EOF

  # Should contain string format
  run cat "$CODEX_CONFIG"
  assert_success
  assert_line 'notify = "afplay /System/Library/Sounds/Bottle.aiff"'
}
