#!/usr/bin/env bats
# Manual integration test - demonstrates sleep feature with 10-second timeout
# Run: ./run-tests.sh test/integration-sleep.bats
# Then wait 10 seconds to see ghost sleep, press any key to wake

@test "sleep feature integration test (manual)" {
  skip "Manual test only - requires visual inspection"
  # This test documents the expected behavior
  # Actual integration testing requires running ghost-tab
}
