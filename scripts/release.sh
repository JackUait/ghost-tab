#!/bin/bash
set -euo pipefail

# --- Preflight check functions ---

check_clean_tree() {
  if [[ -n "$(git status --porcelain)" ]]; then
    echo "Error: Working tree is not clean. Commit or stash changes first." >&2
    return 1
  fi
}

check_main_branch() {
  local branch
  branch="$(git rev-parse --abbrev-ref HEAD)"
  if [[ "$branch" != "main" ]]; then
    echo "Error: Must be on main branch (currently on '$branch')." >&2
    return 1
  fi
}

read_version() {
  local version_file="$1"
  if [[ ! -f "$version_file" ]]; then
    echo "Error: VERSION file not found at $version_file" >&2
    return 1
  fi
  local version
  version="$(tr -d '[:space:]' < "$version_file")"
  if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: VERSION '$version' is not valid semver (expected X.Y.Z)" >&2
    return 1
  fi
  echo "$version"
}

check_tag_not_exists() {
  local tag="$1"
  if git rev-parse "$tag" &>/dev/null; then
    echo "Error: Tag $tag already exists." >&2
    return 1
  fi
}

check_gh_auth() {
  if ! command -v gh &>/dev/null; then
    echo "Error: gh CLI is not installed. Install with: brew install gh" >&2
    return 1
  fi
  if ! gh auth status &>/dev/null; then
    echo "Error: gh CLI is not authenticated. Run: gh auth login" >&2
    return 1
  fi
}

check_formula_exists() {
  local formula_path="$1"
  if [[ ! -f "$formula_path" ]]; then
    echo "Error: Homebrew formula not found at $formula_path" >&2
    return 1
  fi
}

# Only run main when executed directly (not sourced for testing)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
