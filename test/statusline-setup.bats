setup() {
  load 'test_helper/common'
  _common_setup
  source "$PROJECT_ROOT/lib/tui.sh"
  source "$PROJECT_ROOT/lib/settings-json.sh"
  source "$PROJECT_ROOT/lib/statusline-setup.sh"
  TEST_TMP="$(mktemp -d)"

  # Create a fake share dir with template files
  SHARE_DIR="$TEST_TMP/share"
  mkdir -p "$SHARE_DIR/templates"
  echo "mock-settings" > "$SHARE_DIR/templates/ccstatusline-settings.json"
  echo "mock-command" > "$SHARE_DIR/templates/statusline-command.sh"
  echo "mock-wrapper" > "$SHARE_DIR/templates/statusline-wrapper.sh"

  # Create fake home dirs
  FAKE_HOME="$TEST_TMP/home"
  mkdir -p "$FAKE_HOME/.config/ccstatusline"
  mkdir -p "$FAKE_HOME/.claude"
}

teardown() {
  rm -rf "$TEST_TMP"
}

@test "setup_statusline: copies config and scripts when npm available" {
  # Mock npm to succeed for all calls
  npm() { return 0; }
  export -f npm

  # Mock _has_npm to return true
  _has_npm() { return 0; }
  export -f _has_npm

  setup_statusline "$SHARE_DIR" "$FAKE_HOME/.claude/settings.json" "$FAKE_HOME"
  [ -f "$FAKE_HOME/.config/ccstatusline/settings.json" ]
  [ -f "$FAKE_HOME/.claude/statusline-command.sh" ]
  [ -f "$FAKE_HOME/.claude/statusline-wrapper.sh" ]
  [ -x "$FAKE_HOME/.claude/statusline-command.sh" ]
  [ -x "$FAKE_HOME/.claude/statusline-wrapper.sh" ]
}

@test "setup_statusline: skips when npm not available and brew fails" {
  # Mock _has_npm to return false (npm not found)
  _has_npm() { return 1; }
  export -f _has_npm
  brew() { return 1; }
  export -f brew

  run setup_statusline "$SHARE_DIR" "$FAKE_HOME/.claude/settings.json" "$FAKE_HOME"
  [ "$status" -eq 0 ]
  [ ! -f "$FAKE_HOME/.claude/statusline-command.sh" ]
}

@test "setup_statusline: reports already installed" {
  npm() {
    if [[ "$1" == "list" ]]; then return 0; fi
    return 0
  }
  export -f npm
  _has_npm() { return 0; }
  export -f _has_npm

  run setup_statusline "$SHARE_DIR" "$FAKE_HOME/.claude/settings.json" "$FAKE_HOME"
  assert_output --partial "already installed"
}

@test "setup_statusline: warns and skips when npm install fails" {
  _has_npm() { return 0; }
  export -f _has_npm
  npm() {
    if [[ "$1" == "list" ]]; then return 1; fi
    if [[ "$1" == "install" ]]; then return 1; fi
    return 0
  }
  export -f npm

  run setup_statusline "$SHARE_DIR" "$FAKE_HOME/.claude/settings.json" "$FAKE_HOME"
  [ "$status" -eq 0 ]
  assert_output --partial "Failed to install"
  [ ! -f "$FAKE_HOME/.claude/statusline-command.sh" ]
  [ ! -f "$FAKE_HOME/.claude/statusline-wrapper.sh" ]
}

@test "setup_statusline: installs ccstatusline and copies files on fresh install" {
  _has_npm() { return 0; }
  export -f _has_npm
  # First npm list call returns 1 (not installed), subsequent calls return 0
  _npm_list_call_count=0
  export _npm_list_call_count
  npm() {
    if [[ "$1" == "list" ]]; then
      _npm_list_call_count=$((_npm_list_call_count + 1))
      export _npm_list_call_count
      if [[ "$_npm_list_call_count" -eq 1 ]]; then return 1; fi
      return 0
    fi
    if [[ "$1" == "install" ]]; then return 0; fi
    return 0
  }
  export -f npm

  run setup_statusline "$SHARE_DIR" "$FAKE_HOME/.claude/settings.json" "$FAKE_HOME"
  assert_output --partial "ccstatusline installed"
  refute_output --partial "already installed"
  [ -f "$FAKE_HOME/.config/ccstatusline/settings.json" ]
  [ -f "$FAKE_HOME/.claude/statusline-command.sh" ]
  [ -f "$FAKE_HOME/.claude/statusline-wrapper.sh" ]
}

@test "setup_statusline: calls merge_claude_settings after file copy" {
  _has_npm() { return 0; }
  export -f _has_npm
  npm() { return 0; }
  export -f npm

  local claude_settings="$TEST_TMP/claude-settings/settings.json"

  run setup_statusline "$SHARE_DIR" "$claude_settings" "$FAKE_HOME"
  [ "$status" -eq 0 ]
  [ -f "$claude_settings" ]
  grep -q '"statusLine"' "$claude_settings"
}
