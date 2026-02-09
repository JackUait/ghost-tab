setup() {
  load 'test_helper/common'
  _common_setup
  source "$PROJECT_ROOT/lib/tui.sh"
  source "$PROJECT_ROOT/lib/ai-select.sh"
  TEST_TMP="$(mktemp -d)"
}

teardown() {
  rm -rf "$TEST_TMP"
}

# --- save_ai_tool_preference ---

@test "save_ai_tool_preference: claude first when selected" {
  save_ai_tool_preference 1 1 0 0 "$TEST_TMP"
  run cat "$TEST_TMP/ai-tool"
  assert_output "claude"
}

@test "save_ai_tool_preference: codex when claude not selected" {
  save_ai_tool_preference 0 1 0 0 "$TEST_TMP"
  run cat "$TEST_TMP/ai-tool"
  assert_output "codex"
}

@test "save_ai_tool_preference: copilot when claude and codex not selected" {
  save_ai_tool_preference 0 0 1 0 "$TEST_TMP"
  run cat "$TEST_TMP/ai-tool"
  assert_output "copilot"
}

@test "save_ai_tool_preference: opencode as last fallback" {
  save_ai_tool_preference 0 0 0 1 "$TEST_TMP"
  run cat "$TEST_TMP/ai-tool"
  assert_output "opencode"
}

@test "save_ai_tool_preference: creates parent directory" {
  save_ai_tool_preference 1 0 0 0 "$TEST_TMP/nested/dir"
  [ -f "$TEST_TMP/nested/dir/ai-tool" ]
  run cat "$TEST_TMP/nested/dir/ai-tool"
  assert_output "claude"
}

@test "save_ai_tool_preference: does nothing when none selected" {
  save_ai_tool_preference 0 0 0 0 "$TEST_TMP"
  [ ! -f "$TEST_TMP/ai-tool" ]
}

# --- run_ai_tool_select ---

@test "run_ai_tool_select: function is defined" {
  declare -f run_ai_tool_select >/dev/null
}
