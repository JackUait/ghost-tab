setup() {
  load 'test_helper/common'
  _common_setup
  source "$PROJECT_ROOT/lib/tui.sh"
  source "$PROJECT_ROOT/lib/ai-tools.sh"
  source "$PROJECT_ROOT/lib/menu.sh"
}

# --- draw_menu: function exists ---

@test "draw_menu: function is defined" {
  declare -f draw_menu >/dev/null
}

@test "draw_menu: function is callable (type -t)" {
  [ "$(type -t draw_menu)" = "function" ]
}

# --- draw_menu: right border clears line properly ---

@test "draw_menu: _rbdr function definition is correct" {
  # Bug fix verification: _rbdr() should position first, then print border+clear
  # NOT clear first then position (which would clear content at wrong location)
  source "$PROJECT_ROOT/lib/tui.sh"
  source "$PROJECT_ROOT/lib/ai-tools.sh"
  source "$PROJECT_ROOT/lib/menu.sh"

  # Extract the _rbdr function definition from draw_menu
  # The correct implementation should be: moveto ... printf ...│...K
  # The buggy implementation would be: printf ...K... moveto ... printf ...│

  # Check that the source code has the fix
  rbdr_def=$(grep -A0 "_rbdr()" "$PROJECT_ROOT/lib/menu.sh")

  # Correct pattern: moveto comes before the border character │
  if ! echo "$rbdr_def" | grep -q 'moveto.*│'; then
    echo "ERROR: _rbdr doesn't call moveto before printing border"
    return 1
  fi

  # Verify clear (\033[K) comes after the border in the printf
  if ! echo "$rbdr_def" | grep -q '│.*\\033\[K'; then
    echo "ERROR: _rbdr doesn't have clear after border"
    return 1
  fi

  # Ensure clear does NOT come before moveto (the bug)
  if echo "$rbdr_def" | grep -q '\\033\[K.*moveto'; then
    echo "BUG: _rbdr has clear before moveto"
    return 1
  fi

  return 0
}

@test "draw_menu: clears entire line before printing content" {
  # Bug: Old content can remain visible if new content is shorter
  # Fix: Clear the entire line (or at least to right border) before printing new content

  # Check that after printing left border, we clear before printing content
  # Pattern should be: moveto, print │, clear to EOL or right border, then print content

  # Check menu item rendering (around line 112-128)
  item_render=$(sed -n '111,129p' "$PROJECT_ROOT/lib/menu.sh")

  # After printing left border │, before printing content, we should clear
  # The code should have a pattern like: printf "│" followed by some clearing mechanism
  # Currently it doesn't clear, which causes the bug

  # For now, verify the issue exists by checking that we DON'T clear after left border
  if echo "$item_render" | grep -A1 'printf.*│.*_NC' | grep -q '\\033\[2K\|\\033\[K'; then
    # If we find clearing after left border, the fix is in place
    return 0
  else
    # Expected to fail initially - no clearing after left border
    echo "BUG: No clear after left border before printing content"
    return 1
  fi
}
