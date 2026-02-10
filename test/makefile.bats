#!/usr/bin/env bats

load test_helper/bats-support/load
load test_helper/bats-assert/load

setup() {
  # Clean before tests
  make clean 2>/dev/null || true
}

teardown() {
  # Clean after tests
  make clean 2>/dev/null || true
  rm -f "$HOME/.local/bin/ghost-tab-tui"
}

@test "make build creates binary" {
  run make build
  assert_success
  assert_output --partial "Building ghost-tab-tui"
  assert_output --partial "✓ Built bin/ghost-tab-tui"
  assert [ -f bin/ghost-tab-tui ]
  assert [ -x bin/ghost-tab-tui ]
}

@test "make clean removes binary" {
  make build
  run make clean
  assert_success
  refute [ -f bin/ghost-tab-tui ]
}

@test "make help shows targets" {
  run make help
  assert_success
  assert_output --partial "make build"
  assert_output --partial "make install"
  assert_output --partial "make test"
}
