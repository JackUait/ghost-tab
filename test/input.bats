setup() {
  load 'test_helper/common'
  _common_setup
  source "$PROJECT_ROOT/lib/input.sh"
}

@test "parse_esc_sequence: up arrow" {
  result="$(printf '[A' | parse_esc_sequence)"
  [[ "$result" == "A" ]]
}

@test "parse_esc_sequence: down arrow" {
  result="$(printf '[B' | parse_esc_sequence)"
  [[ "$result" == "B" ]]
}

@test "parse_esc_sequence: left arrow" {
  result="$(printf '[D' | parse_esc_sequence)"
  [[ "$result" == "D" ]]
}

@test "parse_esc_sequence: right arrow" {
  result="$(printf '[C' | parse_esc_sequence)"
  [[ "$result" == "C" ]]
}

@test "parse_esc_sequence: SGR mouse left click" {
  result="$(printf '[<0;15;3M' | parse_esc_sequence)"
  [[ "$result" == "click:3" ]]
}

@test "parse_esc_sequence: SGR mouse left click different row" {
  result="$(printf '[<0;22;10M' | parse_esc_sequence)"
  [[ "$result" == "click:10" ]]
}

@test "parse_esc_sequence: ignores mouse release" {
  result="$(printf '[<0;15;3m' | parse_esc_sequence)"
  [[ "$result" == "" ]]
}

@test "parse_esc_sequence: ignores right click" {
  result="$(printf '[<2;15;3M' | parse_esc_sequence)"
  [[ "$result" == "" ]]
}

@test "parse_esc_sequence: ignores middle click" {
  result="$(printf '[<1;15;3M' | parse_esc_sequence)"
  [[ "$result" == "" ]]
}
