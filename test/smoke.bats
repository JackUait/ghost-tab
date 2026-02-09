setup() {
  load 'test_helper/common'
  _common_setup
}

@test "bats infrastructure works" {
  run echo "hello"
  assert_success
  assert_output "hello"
}
