#!/usr/bin/env bats
# shellcheck disable=SC2218  # Test mocks override functions sourced from lib/logo-animation.sh

setup() {
  load 'test_helper/common'
  _common_setup
  source "$PROJECT_ROOT/lib/tui.sh"
  source "$PROJECT_ROOT/lib/logo-animation.sh"

  # Set up temp directory for flag files
  TEMP_DIR="$(mktemp -d)"
  export TMPDIR="$TEMP_DIR"
}

teardown() {
  # Clean up any running animation processes
  if [ -n "$_LOGO_ANIM_PID" ]; then
    kill "$_LOGO_ANIM_PID" 2>/dev/null || true
    wait "$_LOGO_ANIM_PID" 2>/dev/null || true
  fi

  # Remove flag files
  rm -f /tmp/ghost-tab-anim-*

  # Clean up temp directory
  rm -rf "$TEMP_DIR"
}

# ============================================================================
# Helper Functions (_c and _r)
# ============================================================================

@test "_c: returns ANSI color code for given number" {
  result="$(_c 209)"
  [[ "$result" == $'\033[38;5;209m' ]]
}

@test "_r: returns ANSI reset code" {
  result="$(_r)"
  [[ "$result" == $'\033[0m' ]]
}

# ============================================================================
# Logo Art Functions - Claude
# ============================================================================

@test "logo_art_claude: sets _LOGO_LINES array with 15 elements" {
  logo_art_claude
  assert [ "${#_LOGO_LINES[@]}" -eq 15 ]
}

@test "logo_art_claude: sets _LOGO_HEIGHT to 15" {
  logo_art_claude
  assert [ "$_LOGO_HEIGHT" -eq 15 ]
}

@test "logo_art_claude: sets _LOGO_WIDTH to 28" {
  logo_art_claude
  assert [ "$_LOGO_WIDTH" -eq 28 ]
}

@test "logo_art_claude: all lines are exactly 28 visible characters" {
  logo_art_claude

  for i in $(seq 0 14); do
    local line="${_LOGO_LINES[$i]}"
    # Strip ANSI codes
    line=$(echo -e "$line" | sed 's/\x1b\[[0-9;]*m//g')
    local len=${#line}
    assert [ "$len" -eq 28 ]
  done
}

@test "logo_art_claude: contains colored blocks" {
  logo_art_claude

  # Verify it contains block characters (▄ or █)
  local combined="${_LOGO_LINES[*]}"
  [[ "$combined" =~ ▄ ]] || [[ "$combined" =~ █ ]]
}

# ============================================================================
# Logo Art Functions - Claude Sleeping
# ============================================================================

@test "logo_art_claude_sleeping: sets correct array size" {
  logo_art_claude_sleeping
  assert [ "${#_LOGO_LINES[@]}" -eq 15 ]
  assert [ "$_LOGO_HEIGHT" -eq 15 ]
  assert [ "$_LOGO_WIDTH" -eq 28 ]
}

@test "logo_art_claude_sleeping: all lines are exactly 28 visible characters" {
  logo_art_claude_sleeping

  for i in $(seq 0 14); do
    local line="${_LOGO_LINES[$i]}"
    line=$(echo -e "$line" | sed 's/\x1b\[[0-9;]*m//g')
    local len=${#line}
    assert [ "$len" -eq 28 ]
  done
}

@test "logo_art_claude_sleeping: has closed eyes (▬ pattern)" {
  logo_art_claude_sleeping

  # Line 5 should have closed eyes
  [[ "${_LOGO_LINES[5]}" =~ ▬ ]]
}

# ============================================================================
# Logo Art Functions - Codex
# ============================================================================

@test "logo_art_codex: sets correct dimensions" {
  logo_art_codex
  assert [ "${#_LOGO_LINES[@]}" -eq 15 ]
  assert [ "$_LOGO_HEIGHT" -eq 15 ]
  assert [ "$_LOGO_WIDTH" -eq 28 ]
}

@test "logo_art_codex: all lines are exactly 28 visible characters" {
  logo_art_codex

  for i in $(seq 0 14); do
    local line="${_LOGO_LINES[$i]}"
    line=$(echo -e "$line" | sed 's/\x1b\[[0-9;]*m//g')
    local len=${#line}
    assert [ "$len" -eq 28 ]
  done
}

@test "logo_art_codex_sleeping: sets correct dimensions" {
  logo_art_codex_sleeping
  assert [ "${#_LOGO_LINES[@]}" -eq 15 ]
  assert [ "$_LOGO_HEIGHT" -eq 15 ]
  assert [ "$_LOGO_WIDTH" -eq 28 ]
}

@test "logo_art_codex_sleeping: has closed eyes" {
  logo_art_codex_sleeping
  [[ "${_LOGO_LINES[5]}" =~ ▬ ]]
}

# ============================================================================
# Logo Art Functions - Copilot
# ============================================================================

@test "logo_art_copilot: sets correct dimensions" {
  logo_art_copilot
  assert [ "${#_LOGO_LINES[@]}" -eq 15 ]
  assert [ "$_LOGO_HEIGHT" -eq 15 ]
  assert [ "$_LOGO_WIDTH" -eq 28 ]
}

@test "logo_art_copilot: all lines are exactly 28 visible characters" {
  logo_art_copilot

  for i in $(seq 0 14); do
    local line="${_LOGO_LINES[$i]}"
    line=$(echo -e "$line" | sed 's/\x1b\[[0-9;]*m//g')
    local len=${#line}
    assert [ "$len" -eq 28 ]
  done
}

@test "logo_art_copilot_sleeping: sets correct dimensions" {
  logo_art_copilot_sleeping
  assert [ "${#_LOGO_LINES[@]}" -eq 15 ]
  assert [ "$_LOGO_HEIGHT" -eq 15 ]
  assert [ "$_LOGO_WIDTH" -eq 28 ]
}

@test "logo_art_copilot_sleeping: has closed eyes" {
  logo_art_copilot_sleeping
  [[ "${_LOGO_LINES[6]}" =~ ▬ ]]
}

# ============================================================================
# Logo Art Functions - OpenCode
# ============================================================================

@test "logo_art_opencode: sets correct dimensions" {
  logo_art_opencode
  assert [ "${#_LOGO_LINES[@]}" -eq 15 ]
  assert [ "$_LOGO_HEIGHT" -eq 15 ]
  assert [ "$_LOGO_WIDTH" -eq 28 ]
}

@test "logo_art_opencode: all lines are exactly 28 visible characters" {
  logo_art_opencode

  for i in $(seq 0 14); do
    local line="${_LOGO_LINES[$i]}"
    line=$(echo -e "$line" | sed 's/\x1b\[[0-9;]*m//g')
    local len=${#line}
    assert [ "$len" -eq 28 ]
  done
}

@test "logo_art_opencode_sleeping: sets correct dimensions" {
  logo_art_opencode_sleeping
  assert [ "${#_LOGO_LINES[@]}" -eq 15 ]
  assert [ "$_LOGO_HEIGHT" -eq 15 ]
  assert [ "$_LOGO_WIDTH" -eq 28 ]
}

@test "logo_art_opencode_sleeping: has closed eyes" {
  logo_art_opencode_sleeping
  [[ "${_LOGO_LINES[5]}" =~ ▬ ]]
}

# ============================================================================
# Drawing Functions - draw_logo
# ============================================================================

@test "draw_logo: executes without error for claude" {
  # Mock moveto to avoid terminal dependencies
  moveto() { echo "MOVE:$1:$2"; }

  run draw_logo 10 5 "claude"
  assert_success
}

@test "draw_logo: executes without error for all tools" {
  moveto() { echo "MOVE:$1:$2"; }

  for tool in claude codex copilot opencode; do
    run draw_logo 10 5 "$tool"
    assert_success
  done
}

@test "draw_logo: calls logo_art function for specified tool" {
  moveto() { :; }

  # Draw claude logo - should populate _LOGO_LINES
  draw_logo 10 5 "claude"

  # Verify _LOGO_LINES was populated
  assert [ "${#_LOGO_LINES[@]}" -eq 15 ]
}

# ============================================================================
# Drawing Functions - clear_logo_area
# ============================================================================

@test "clear_logo_area: executes without error" {
  moveto() { echo "MOVE:$1:$2"; }

  run clear_logo_area 10 5 15 28
  assert_success
}

@test "clear_logo_area: clears specified rectangular area" {
  local move_calls=0
  moveto() {
    move_calls=$((move_calls + 1))
  }

  # Clear a 3-line area
  clear_logo_area 10 5 3 28

  # Should call moveto 3 times (once per line)
  assert [ "$move_calls" -eq 3 ]
}

# ============================================================================
# Drawing Functions - draw_zzz
# ============================================================================

@test "draw_zzz: executes without error" {
  moveto() { :; }
  logo_art_claude  # Set _LOGO_WIDTH

  run draw_zzz 5 10
  assert_success
}

@test "draw_zzz: renders three z characters at different positions" {
  local move_count=0
  moveto() {
    move_count=$((move_count + 1))
  }
  logo_art_claude  # Set _LOGO_WIDTH=28

  draw_zzz 5 10

  # Should call moveto 3 times (for z, Z, Z)
  assert [ "$move_count" -eq 3 ]
}

# ============================================================================
# Drawing Functions - clear_zzz
# ============================================================================

@test "clear_zzz: executes without error" {
  moveto() { :; }
  logo_art_claude  # Set _LOGO_WIDTH

  run clear_zzz 5 10
  assert_success
}

@test "clear_zzz: clears three line positions" {
  local move_count=0
  moveto() {
    move_count=$((move_count + 1))
  }
  logo_art_claude  # Set _LOGO_WIDTH=28

  clear_zzz 5 10

  # Should call moveto 3 times (for 3 lines of zzz)
  assert [ "$move_count" -eq 3 ]
}

# ============================================================================
# Drawing Functions - draw_logo_sleeping
# ============================================================================

@test "draw_logo_sleeping: executes without error for claude" {
  moveto() { :; }

  run draw_logo_sleeping 10 5 "claude"
  assert_success
}

@test "draw_logo_sleeping: calls sleeping variant function" {
  moveto() { :; }

  draw_logo_sleeping 10 5 "claude"

  # Verify sleeping variant was loaded (should have closed eyes)
  [[ "${_LOGO_LINES[5]}" =~ ▬ ]]
}

@test "draw_logo_sleeping: works for all tools" {
  moveto() { :; }

  for tool in claude codex copilot opencode; do
    run draw_logo_sleeping 10 5 "$tool"
    assert_success
  done
}

# ============================================================================
# Animation Lifecycle - start_logo_animation
# ============================================================================

@test "start_logo_animation: creates flag file" {
  # Suppress terminal output by redirecting to /dev/null
  moveto() { :; }
  export -f moveto
  draw_logo() { :; }
  export -f draw_logo
  clear_logo_area() { :; }
  export -f clear_logo_area

  start_logo_animation 10 5 "claude" >/dev/null 2>&1

  # Flag file should exist
  assert [ -f "/tmp/ghost-tab-anim-$$" ]

  # Clean up (stop_logo_animation may return non-zero due to wait on killed process)
  stop_logo_animation || true
}

@test "start_logo_animation: starts background process" {
  moveto() { :; }
  export -f moveto
  draw_logo() { :; }
  export -f draw_logo
  clear_logo_area() { :; }
  export -f clear_logo_area

  start_logo_animation 10 5 "claude" >/dev/null 2>&1

  # _LOGO_ANIM_PID should be set
  assert [ -n "$_LOGO_ANIM_PID" ]

  # Process should be running
  ps -p "$_LOGO_ANIM_PID" > /dev/null

  # Clean up (stop_logo_animation may return non-zero due to wait on killed process)
  stop_logo_animation || true
  sleep 0.1  # Give it time to clean up
}

@test "start_logo_animation: sets global variables" {
  moveto() { :; }
  export -f moveto
  draw_logo() { :; }
  export -f draw_logo
  clear_logo_area() { :; }
  export -f clear_logo_area

  start_logo_animation 10 5 "claude" >/dev/null 2>&1

  assert [ "$_LOGO_CUR_ROW" -eq 10 ]
  assert [ "$_LOGO_CUR_COL" -eq 5 ]
  assert [ "$_LOGO_CUR_TOOL" = "claude" ]

  # Clean up (stop_logo_animation may return non-zero due to wait on killed process)
  stop_logo_animation || true
  sleep 0.1  # Give it time to clean up
}

# ============================================================================
# Animation Lifecycle - stop_logo_animation
# ============================================================================

@test "stop_logo_animation: removes flag file" {
  moveto() { :; }
  export -f moveto
  draw_logo() { :; }
  export -f draw_logo
  clear_logo_area() { :; }
  export -f clear_logo_area

  start_logo_animation 10 5 "claude" >/dev/null 2>&1
  local flagfile="/tmp/ghost-tab-anim-$$"
  assert [ -f "$flagfile" ]

  stop_logo_animation || true

  # Flag file should be gone
  assert [ ! -f "$flagfile" ]
}

@test "stop_logo_animation: kills background process" {
  moveto() { :; }
  export -f moveto
  draw_logo() { :; }
  export -f draw_logo
  clear_logo_area() { :; }
  export -f clear_logo_area

  start_logo_animation 10 5 "claude" >/dev/null 2>&1
  local pid="$_LOGO_ANIM_PID"

  # Verify process is running
  ps -p "$pid" > /dev/null

  stop_logo_animation || true

  # Give it a moment to terminate
  sleep 0.3

  # Process should be gone
  run ps -p "$pid"
  assert_failure
}

@test "stop_logo_animation: unsets _LOGO_ANIM_PID" {
  moveto() { :; }
  export -f moveto
  draw_logo() { :; }
  export -f draw_logo
  clear_logo_area() { :; }
  export -f clear_logo_area

  start_logo_animation 10 5 "claude" >/dev/null 2>&1
  assert [ -n "$_LOGO_ANIM_PID" ]

  stop_logo_animation || true

  # Variable should be unset
  assert [ -z "$_LOGO_ANIM_PID" ]
}

@test "stop_logo_animation: handles no running animation gracefully" {
  moveto() { :; }

  # Stop without starting - should not error
  run stop_logo_animation
  assert_success
}

# ============================================================================
# Animation Lifecycle - Multiple Cycles
# ============================================================================

@test "animation: multiple start/stop cycles work" {
  moveto() { :; }
  export -f moveto
  draw_logo() { :; }
  export -f draw_logo
  clear_logo_area() { :; }
  export -f clear_logo_area

  # First cycle
  start_logo_animation 10 5 "claude" >/dev/null 2>&1
  local pid1="$_LOGO_ANIM_PID"
  ps -p "$pid1" > /dev/null
  stop_logo_animation || true
  sleep 0.3
  run ps -p "$pid1"
  assert_failure

  # Second cycle
  start_logo_animation 10 5 "codex" >/dev/null 2>&1
  local pid2="$_LOGO_ANIM_PID"
  ps -p "$pid2" > /dev/null
  stop_logo_animation || true
  sleep 0.3
  run ps -p "$pid2"
  assert_failure

  # Third cycle
  start_logo_animation 10 5 "copilot" >/dev/null 2>&1
  local pid3="$_LOGO_ANIM_PID"
  ps -p "$pid3" > /dev/null
  stop_logo_animation || true
  sleep 0.3
  run ps -p "$pid3"
  assert_failure
}

@test "animation: rapid start/stop works" {
  moveto() { :; }
  export -f moveto
  draw_logo() { :; }
  export -f draw_logo
  clear_logo_area() { :; }
  export -f clear_logo_area

  start_logo_animation 10 5 "claude" >/dev/null 2>&1
  stop_logo_animation || true
  sleep 0.1

  start_logo_animation 10 5 "claude" >/dev/null 2>&1
  stop_logo_animation || true
  sleep 0.1

  # No zombies should remain
  run stop_logo_animation
  assert_success
}

# ============================================================================
# Animation Lifecycle - No Zombie Processes
# ============================================================================

@test "animation: no zombie processes after stop" {
  moveto() { :; }
  export -f moveto
  draw_logo() { :; }
  export -f draw_logo
  clear_logo_area() { :; }
  export -f clear_logo_area

  start_logo_animation 10 5 "claude" >/dev/null 2>&1
  local pid="$_LOGO_ANIM_PID"

  stop_logo_animation || true
  sleep 0.5

  # Check if process is truly gone (not zombie)
  run ps -p "$pid"
  assert_failure

  # Double-check no zombie state
  if ps -p "$pid" -o state= 2>/dev/null | grep -q 'Z'; then
    fail "Zombie process detected"
  fi
}

# ============================================================================
# Sleep/Wake Transitions - initiate_sleep_transition
# ============================================================================

@test "initiate_sleep_transition: executes without error" {
  # Mock all dependencies
  moveto() { :; }
  stop_logo_animation() { :; }
  draw_logo_sleeping() { :; }
  draw_zzz() { :; }
  sleep() { :; }  # Mock sleep to speed up test

  # Set required globals
  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"
  _ghost_sleeping=0

  run initiate_sleep_transition
  assert_success
}

@test "initiate_sleep_transition: sets _ghost_sleeping to 1" {
  moveto() { :; }
  stop_logo_animation() { :; }
  draw_logo_sleeping() { :; }
  draw_zzz() { :; }
  sleep() { :; }

  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"
  _ghost_sleeping=0

  initiate_sleep_transition

  assert [ "$_ghost_sleeping" -eq 1 ]
}

@test "initiate_sleep_transition: calls stop_logo_animation" {
  moveto() { :; }
  local stop_called=0
  stop_logo_animation() { stop_called=1; }
  draw_logo_sleeping() { :; }
  draw_zzz() { :; }
  sleep() { :; }

  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"

  initiate_sleep_transition

  assert [ "$stop_called" -eq 1 ]
}

@test "initiate_sleep_transition: calls draw_logo_sleeping" {
  moveto() { :; }
  stop_logo_animation() { :; }
  local draw_sleeping_called=0
  draw_logo_sleeping() { draw_sleeping_called=1; }
  draw_zzz() { :; }
  sleep() { :; }

  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"

  initiate_sleep_transition

  assert [ "$draw_sleeping_called" -eq 1 ]
}

@test "initiate_sleep_transition: calls draw_zzz" {
  moveto() { :; }
  stop_logo_animation() { :; }
  draw_logo_sleeping() { :; }
  local draw_zzz_called=0
  draw_zzz() { draw_zzz_called=1; }
  sleep() { :; }

  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"

  initiate_sleep_transition

  assert [ "$draw_zzz_called" -eq 1 ]
}

# ============================================================================
# Sleep/Wake Transitions - wake_ghost
# ============================================================================

@test "wake_ghost: executes without error" {
  moveto() { :; }
  clear_zzz() { :; }
  start_logo_animation() { :; }

  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"
  _ghost_sleeping=1
  SECONDS=100

  run wake_ghost
  assert_success
}

@test "wake_ghost: sets _ghost_sleeping to 0" {
  moveto() { :; }
  clear_zzz() { :; }
  start_logo_animation() { :; }

  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"
  _ghost_sleeping=1
  SECONDS=100

  wake_ghost

  assert [ "$_ghost_sleeping" -eq 0 ]
}

@test "wake_ghost: updates _last_interaction" {
  moveto() { :; }
  clear_zzz() { :; }
  start_logo_animation() { :; }

  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"
  _ghost_sleeping=1
  SECONDS=100
  _last_interaction=50

  wake_ghost

  assert [ "$_last_interaction" -eq 100 ]
}

@test "wake_ghost: calls clear_zzz" {
  moveto() { :; }
  local clear_zzz_called=0
  clear_zzz() { clear_zzz_called=1; }
  start_logo_animation() { :; }

  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"
  _ghost_sleeping=1
  SECONDS=100

  wake_ghost

  assert [ "$clear_zzz_called" -eq 1 ]
}

@test "wake_ghost: calls start_logo_animation" {
  moveto() { :; }
  clear_zzz() { :; }
  local start_called=0
  start_logo_animation() { start_called=1; }

  _logo_row=10
  _logo_col=5
  SELECTED_AI_TOOL="claude"
  _ghost_sleeping=1
  SECONDS=100

  wake_ghost

  assert [ "$start_called" -eq 1 ]
}

# ============================================================================
# Bob Offsets Array
# ============================================================================

@test "_BOB_OFFSETS: array is defined" {
  assert [ -n "${_BOB_OFFSETS[*]}" ]
}

@test "_BOB_OFFSETS: has 14 elements" {
  assert [ "${#_BOB_OFFSETS[@]}" -eq 14 ]
}

@test "_BOB_OFFSETS: contains only 0 and 1 values" {
  for offset in "${_BOB_OFFSETS[@]}"; do
    assert [ "$offset" -eq 0 ] || [ "$offset" -eq 1 ]
  done
}

@test "_BOB_MAX: is set to 1" {
  assert [ "$_BOB_MAX" -eq 1 ]
}
