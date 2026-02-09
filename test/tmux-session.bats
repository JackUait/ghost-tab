setup() {
  load 'test_helper/common'
  _common_setup
  source "$PROJECT_ROOT/lib/process.sh"
  source "$PROJECT_ROOT/lib/tmux-session.sh"
  TEST_TMP="$(mktemp -d)"
}

teardown() {
  rm -rf "$TEST_TMP"
}

# --- build_ai_launch_cmd ---

@test "build_ai_launch_cmd: claude passes args through" {
  run build_ai_launch_cmd "claude" "/usr/bin/claude" "" "" "" "--resume"
  assert_output '/usr/bin/claude --resume'
}

@test "build_ai_launch_cmd: codex uses --cd flag" {
  run build_ai_launch_cmd "codex" "" "/usr/bin/codex" "" "" "/my/project"
  assert_output '/usr/bin/codex --cd "/my/project"'
}

@test "build_ai_launch_cmd: copilot has no extra args" {
  run build_ai_launch_cmd "copilot" "" "" "/usr/bin/copilot" ""
  assert_output "/usr/bin/copilot"
}

@test "build_ai_launch_cmd: opencode passes project dir" {
  run build_ai_launch_cmd "opencode" "" "" "" "/usr/bin/opencode" "/my/project"
  assert_output '/usr/bin/opencode "/my/project"'
}

# --- cleanup_tmux_session ---

@test "cleanup_tmux_session: calls kill and tmux kill-session" {
  _calls=()
  kill() { _calls+=("kill:$*"); return 0; }
  export -f kill

  tmux() {
    if [[ "$1" == "list-panes" ]]; then
      echo "12345"
    elif [[ "$1" == "kill-session" ]]; then
      _calls+=("kill-session:$*")
    fi
    return 0
  }
  export -f tmux

  kill_tree() { _calls+=("kill_tree:$*"); return 0; }
  export -f kill_tree

  sleep() { return 0; }
  export -f sleep

  cleanup_tmux_session "test-session" "99999" "tmux"
}

@test "cleanup_tmux_session: handles missing session gracefully" {
  tmux() { return 1; }
  export -f tmux
  kill() { return 0; }
  export -f kill
  sleep() { return 0; }
  export -f sleep
  kill_tree() { return 0; }
  export -f kill_tree

  run cleanup_tmux_session "nonexistent" "99999" "tmux"
  [ "$status" -eq 0 ]
}
