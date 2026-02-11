setup() {
  load 'test_helper/common'
  _common_setup
}

@test "select_ai_tool_interactive calls ghost-tab-tui select-ai-tool" {
  source "$PROJECT_ROOT/lib/ai-select-tui.sh"

  # Mock ghost-tab-tui
  ghost-tab-tui() {
    if [[ "$1" == "select-ai-tool" ]]; then
      echo '{"tool":"claude","command":"claude","selected":true}'
      return 0
    fi
    return 1
  }
  export -f ghost-tab-tui

  # Mock jq
  jq() {
    if [[ "$2" == ".selected" ]]; then
      echo "true"
    elif [[ "$2" == ".tool" ]]; then
      echo "claude"
    fi
    return 0
  }
  export -f jq

  run select_ai_tool_interactive

  assert_success
  [[ "$_selected_ai_tool" == "claude" ]]
}

@test "select_ai_tool_interactive returns failure when cancelled" {
  source "$PROJECT_ROOT/lib/ai-select-tui.sh"

  # Mock ghost-tab-tui (cancelled)
  ghost-tab-tui() {
    if [[ "$1" == "select-ai-tool" ]]; then
      echo '{"selected":false}'
      return 0
    fi
    return 1
  }
  export -f ghost-tab-tui

  # Mock jq
  jq() {
    if [[ "$2" == ".selected" ]]; then
      echo "false"
    fi
    return 0
  }
  export -f jq

  run select_ai_tool_interactive

  assert_failure
}

@test "select_ai_tool_interactive handles binary missing" {
  source "$PROJECT_ROOT/lib/ai-select-tui.sh"

  # Override command to simulate missing binary
  command() {
    if [[ "$1" == "-v" && "$2" == "ghost-tab-tui" ]]; then
      return 1
    fi
    builtin command "$@"
  }
  export -f command

  run select_ai_tool_interactive

  assert_failure
  assert_output --partial "ghost-tab-tui binary not found"
}

@test "select_ai_tool_interactive handles jq parse failure for selected" {
  source "$PROJECT_ROOT/lib/ai-select-tui.sh"

  # Mock ghost-tab-tui
  ghost-tab-tui() {
    echo '{"tool":"claude","selected":true}'
    return 0
  }
  export -f ghost-tab-tui

  # Mock jq (fails on first call)
  jq() {
    return 1
  }
  export -f jq

  run select_ai_tool_interactive

  assert_failure
  assert_output --partial "Failed to parse AI tool selection response"
}

@test "select_ai_tool_interactive handles jq parse failure for tool" {
  source "$PROJECT_ROOT/lib/ai-select-tui.sh"

  # Mock ghost-tab-tui
  ghost-tab-tui() {
    echo '{"tool":"claude","selected":true}'
    return 0
  }
  export -f ghost-tab-tui

  # Mock jq (first call succeeds, second fails)
  _jq_call_count=0
  jq() {
    _jq_call_count=$((_jq_call_count + 1))
    if [[ $_jq_call_count -eq 1 ]]; then
      echo "true"
      return 0
    fi
    return 1
  }
  export -f jq

  run select_ai_tool_interactive

  assert_failure
  assert_output --partial "Failed to parse selected tool"
}

@test "select_ai_tool_interactive validates against null selected" {
  source "$PROJECT_ROOT/lib/ai-select-tui.sh"

  # Mock ghost-tab-tui
  ghost-tab-tui() {
    echo '{"selected":"null"}'
    return 0
  }
  export -f ghost-tab-tui

  # Mock jq
  jq() {
    if [[ "$2" == ".selected" ]]; then
      echo "null"
    fi
    return 0
  }
  export -f jq

  run select_ai_tool_interactive

  assert_failure
  assert_output --partial "TUI returned invalid selection status"
}

@test "select_ai_tool_interactive validates against null tool" {
  source "$PROJECT_ROOT/lib/ai-select-tui.sh"

  # Mock ghost-tab-tui
  ghost-tab-tui() {
    echo '{"tool":"null","selected":true}'
    return 0
  }
  export -f ghost-tab-tui

  # Mock jq
  _jq_call_count=0
  jq() {
    _jq_call_count=$((_jq_call_count + 1))
    if [[ $_jq_call_count -eq 1 ]]; then
      echo "true"
    else
      echo "null"
    fi
    return 0
  }
  export -f jq

  run select_ai_tool_interactive

  assert_failure
  assert_output --partial "TUI returned invalid tool name"
}

@test "select_ai_tool_interactive validates against empty tool" {
  source "$PROJECT_ROOT/lib/ai-select-tui.sh"

  # Mock ghost-tab-tui
  ghost-tab-tui() {
    echo '{"tool":"","selected":true}'
    return 0
  }
  export -f ghost-tab-tui

  # Mock jq
  _jq_call_count=0
  jq() {
    _jq_call_count=$((_jq_call_count + 1))
    if [[ $_jq_call_count -eq 1 ]]; then
      echo "true"
    else
      echo ""
    fi
    return 0
  }
  export -f jq

  run select_ai_tool_interactive

  assert_failure
  assert_output --partial "TUI returned invalid tool name"
}
