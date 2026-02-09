#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
"$SCRIPT_DIR/test/test_helper/bats-core/bin/bats" "$SCRIPT_DIR/test/"*.bats "$@"
