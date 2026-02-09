setup() {
  load 'test_helper/common'
  _common_setup
  source "$PROJECT_ROOT/lib/projects.sh"
  TEST_DIR="$(mktemp -d)"
}

teardown() {
  rm -rf "$TEST_DIR"
}

# --- parse_project_name ---

@test "parse_project_name: extracts name before colon" {
  run parse_project_name "myapp:/Users/me/myapp"
  assert_output "myapp"
}

@test "parse_project_name: handles name with spaces" {
  run parse_project_name "my app:/Users/me/my app"
  assert_output "my app"
}

# --- parse_project_path ---

@test "parse_project_path: extracts path after first colon" {
  run parse_project_path "myapp:/Users/me/myapp"
  assert_output "/Users/me/myapp"
}

@test "parse_project_path: handles paths with colons" {
  run parse_project_path "myapp:/Users/me/path:with:colons"
  assert_output "/Users/me/path:with:colons"
}

# --- load_projects ---

@test "load_projects: reads name:path lines" {
  cat > "$TEST_DIR/projects" << 'EOF'
app1:/path/to/app1
app2:/path/to/app2
EOF
  run load_projects "$TEST_DIR/projects"
  assert_line --index 0 "app1:/path/to/app1"
  assert_line --index 1 "app2:/path/to/app2"
}

@test "load_projects: skips blank lines" {
  cat > "$TEST_DIR/projects" << 'EOF'
app1:/path/to/app1

app2:/path/to/app2
EOF
  run load_projects "$TEST_DIR/projects"
  assert_output --partial "app1:/path/to/app1"
  assert_output --partial "app2:/path/to/app2"
  [ "$(echo "$output" | wc -l | tr -d ' ')" -eq 2 ]
}

@test "load_projects: skips comment lines" {
  cat > "$TEST_DIR/projects" << 'EOF'
# This is a comment
app1:/path/to/app1
EOF
  run load_projects "$TEST_DIR/projects"
  assert_output "app1:/path/to/app1"
}

@test "load_projects: returns empty for missing file" {
  run load_projects "$TEST_DIR/nonexistent"
  assert_output ""
}

# --- path_expand ---

@test "path_expand: converts ~ to HOME" {
  run path_expand "~/projects/app"
  assert_output "$HOME/projects/app"
}

@test "path_expand: leaves absolute paths unchanged" {
  run path_expand "/usr/local/bin"
  assert_output "/usr/local/bin"
}

@test "path_expand: leaves relative paths unchanged" {
  run path_expand "relative/path"
  assert_output "relative/path"
}

# --- path_truncate ---

@test "path_truncate: returns short paths unchanged" {
  run path_truncate "~/short" 38
  assert_output "~/short"
}

@test "path_truncate: truncates long paths with ellipsis" {
  long_path="~/very/long/deeply/nested/project/directory/name"
  result="$(path_truncate "$long_path" 30)"
  [[ "$result" == *"..."* ]]
  [ "${#result}" -le 30 ]
}

@test "path_truncate: preserves start and end of path" {
  long_path="~/very/long/deeply/nested/project/directory/name"
  result="$(path_truncate "$long_path" 30)"
  [[ "$result" == "~/"* ]]
  [[ "$result" == *"name" ]]
}

@test "path_truncate: respects max width parameter" {
  long_path="~/this/is/a/really/long/path/that/goes/on/forever"
  result20="$(path_truncate "$long_path" 20)"
  result40="$(path_truncate "$long_path" 40)"
  [ "${#result20}" -le 20 ]
  [ "${#result40}" -le 40 ]
}
