#!/usr/bin/env bats

load test_helper/bats-support/load
load test_helper/bats-assert/load

setup() {
  load 'test_helper/common'
  _common_setup

  # Create test environment
  TEST_TMP="$(mktemp -d)"
  TEST_WRAPPER_DIR="$TEST_TMP/wrapper"
  TEST_BIN_DIR="$TEST_TMP/bin"
  TEST_SHARE_DIR="$TEST_TMP/share"

  # Set up minimal wrapper environment
  mkdir -p "$TEST_WRAPPER_DIR/lib"
  mkdir -p "$TEST_BIN_DIR"
  mkdir -p "$TEST_SHARE_DIR/cmd/ghost-tab-tui"

  # Add test bin to PATH
  export PATH="$TEST_BIN_DIR:$PATH"

  # Copy real lib files for sourcing
  cp "$PROJECT_ROOT/lib/tui.sh" "$TEST_WRAPPER_DIR/lib/"
  cp "$PROJECT_ROOT/lib/ai-tools.sh" "$TEST_WRAPPER_DIR/lib/"
  cp "$PROJECT_ROOT/lib/projects.sh" "$TEST_WRAPPER_DIR/lib/"
  cp "$PROJECT_ROOT/lib/process.sh" "$TEST_WRAPPER_DIR/lib/"
  cp "$PROJECT_ROOT/lib/input.sh" "$TEST_WRAPPER_DIR/lib/"
  cp "$PROJECT_ROOT/lib/update.sh" "$TEST_WRAPPER_DIR/lib/"
  cp "$PROJECT_ROOT/lib/menu-tui.sh" "$TEST_WRAPPER_DIR/lib/"
  cp "$PROJECT_ROOT/lib/project-actions.sh" "$TEST_WRAPPER_DIR/lib/"
  cp "$PROJECT_ROOT/lib/tmux-session.sh" "$TEST_WRAPPER_DIR/lib/"
  cp "$PROJECT_ROOT/lib/settings-menu-tui.sh" "$TEST_WRAPPER_DIR/lib/"

  # Create minimal ghost-tab-tui source for testing builds
  cat > "$TEST_SHARE_DIR/cmd/ghost-tab-tui/main.go" <<'EOF'
package main
import "fmt"
func main() {
  fmt.Println("ghost-tab-tui test version")
}
EOF

  # Create minimal go.mod at SHARE_DIR root
  cat > "$TEST_SHARE_DIR/go.mod" <<'EOF'
module github.com/user/ghost-tab

go 1.21
EOF

  # Create wrapper test script
  cat > "$TEST_WRAPPER_DIR/test-wrapper.sh" <<'WRAPPER_EOF'
#!/bin/bash
export PATH="$PATH"

# Self-healing: Check if ghost-tab-tui exists, rebuild if missing
TUI_BIN="$HOME/.local/bin/ghost-tab-tui"
if ! command -v ghost-tab-tui &>/dev/null; then
  # Simple inline rebuild without TUI functions (not loaded yet)
  if command -v go &>/dev/null; then
    printf 'Rebuilding ghost-tab-tui...\n' >&2
    mkdir -p "$HOME/.local/bin"
    # Build from module root with relative path to cmd
    if (cd "$SHARE_DIR" && go build -o "$HOME/.local/bin/ghost-tab-tui" ./cmd/ghost-tab-tui) 2>/dev/null; then
      printf 'ghost-tab-tui rebuilt successfully\n' >&2
      export PATH="$HOME/.local/bin:$PATH"
    else
      printf '\033[31mError:\033[0m Failed to rebuild ghost-tab-tui\n' >&2
      printf 'Run \033[1mghost-tab\033[0m to reinstall.\n' >&2
      printf 'Press any key to exit...\n' >&2
      read -rsn1
      exit 1
    fi
  else
    printf '\033[31mError:\033[0m ghost-tab-tui binary not found and Go not installed\n' >&2
    printf 'Run \033[1mghost-tab\033[0m to reinstall.\n' >&2
    printf 'Press any key to exit...\n' >&2
    read -rsn1
    exit 1
  fi
fi

# Now proceed with normal wrapper startup
_WRAPPER_DIR="$(cd "$(dirname "$0")" && pwd)"

if [ ! -d "$_WRAPPER_DIR/lib" ]; then
  printf '\033[31mError:\033[0m Ghost Tab libraries not found\n' >&2
  exit 1
fi

# Load minimal libs for testing
for _gt_lib in tui; do
  # shellcheck disable=SC1090
  source "$_WRAPPER_DIR/lib/${_gt_lib}.sh"
done

success "Wrapper started successfully"
echo "tui-command: $(command -v ghost-tab-tui)"
WRAPPER_EOF

  chmod +x "$TEST_WRAPPER_DIR/test-wrapper.sh"
}

teardown() {
  rm -rf "$TEST_TMP"
}

@test "self-healing: continues normally when ghost-tab-tui exists" {
  # Create a fake ghost-tab-tui binary
  cat > "$TEST_BIN_DIR/ghost-tab-tui" <<'EOF'
#!/bin/bash
echo "ghost-tab-tui v1.0.0"
EOF
  chmod +x "$TEST_BIN_DIR/ghost-tab-tui"

  export SHARE_DIR="$TEST_SHARE_DIR"

  run bash "$TEST_WRAPPER_DIR/test-wrapper.sh"

  assert_success
  assert_output --partial "Wrapper started successfully"
  refute_output --partial "Rebuilding"
}

@test "self-healing: rebuilds ghost-tab-tui when missing and Go available" {
  # Skip if Go not installed on test system
  if ! command -v go &>/dev/null; then
    skip "Go not installed on test system"
  fi

  # Create isolated environment without ghost-tab-tui
  export SHARE_DIR="$TEST_SHARE_DIR"
  export HOME="$TEST_TMP/home"
  # Remove both .local/bin and current directory from PATH
  export PATH="/usr/local/bin:/usr/bin:/bin:$(command -v go | xargs dirname)"
  mkdir -p "$HOME/.local/bin"

  run bash "$TEST_WRAPPER_DIR/test-wrapper.sh"

  assert_success
  assert_output --partial "Rebuilding ghost-tab-tui"
  assert_output --partial "ghost-tab-tui rebuilt successfully"
  assert_output --partial "Wrapper started successfully"

  # Verify binary was created
  [ -f "$HOME/.local/bin/ghost-tab-tui" ]
}

@test "self-healing: fails gracefully when ghost-tab-tui missing and Go not installed" {
  # Minimal PATH without go or ghost-tab-tui
  export PATH="/usr/local/bin:/usr/bin:/bin"
  export SHARE_DIR="$TEST_SHARE_DIR"
  export HOME="$TEST_TMP/home"

  # Need to simulate user pressing a key
  run bash -c "echo '' | bash '$TEST_WRAPPER_DIR/test-wrapper.sh'"

  assert_failure
  assert_output --partial "ghost-tab-tui binary not found and Go not installed"
  assert_output --partial "ghost-tab"
  assert_output --partial "reinstall"
}

@test "self-healing: fails gracefully when rebuild fails" {
  # Skip if Go not installed on test system
  if ! command -v go &>/dev/null; then
    skip "Go not installed on test system"
  fi

  # Create invalid Go source that won't compile
  cat > "$TEST_SHARE_DIR/cmd/ghost-tab-tui/main.go" <<'EOF'
package main
this is invalid go code!
EOF

  # Minimal PATH with go but not ghost-tab-tui
  export PATH="/usr/local/bin:/usr/bin:/bin:$(command -v go | xargs dirname)"
  export SHARE_DIR="$TEST_SHARE_DIR"
  export HOME="$TEST_TMP/home"
  mkdir -p "$HOME/.local/bin"

  # Simulate user pressing key
  run bash -c "echo '' | bash '$TEST_WRAPPER_DIR/test-wrapper.sh'"

  assert_failure
  assert_output --partial "Failed to rebuild ghost-tab-tui"
  assert_output --partial "ghost-tab"
  assert_output --partial "reinstall"
}

@test "self-healing: adds rebuilt binary to PATH" {
  # Skip if Go not installed on test system
  if ! command -v go &>/dev/null; then
    skip "Go not installed on test system"
  fi

  # Minimal PATH with go but not ghost-tab-tui
  export PATH="/usr/local/bin:/usr/bin:/bin:$(command -v go | xargs dirname)"
  export SHARE_DIR="$TEST_SHARE_DIR"
  export HOME="$TEST_TMP/home"
  mkdir -p "$HOME/.local/bin"

  run bash "$TEST_WRAPPER_DIR/test-wrapper.sh"

  assert_success
  assert_output --partial "Rebuilding ghost-tab-tui"
  assert_output --regexp "tui-command:.*\.local/bin/ghost-tab-tui"
}

@test "self-healing: does not add noticeable latency when binary exists" {
  # Create a fake ghost-tab-tui binary
  cat > "$TEST_BIN_DIR/ghost-tab-tui" <<'EOF'
#!/bin/bash
echo "ghost-tab-tui v1.0.0"
EOF
  chmod +x "$TEST_BIN_DIR/ghost-tab-tui"

  export SHARE_DIR="$TEST_SHARE_DIR"

  # Time the execution (should be fast, < 1 second for check)
  start=$(date +%s)
  run bash "$TEST_WRAPPER_DIR/test-wrapper.sh"
  end=$(date +%s)
  elapsed=$((end - start))

  assert_success
  # Should complete in under 2 seconds (very generous)
  [ "$elapsed" -lt 2 ]
}
