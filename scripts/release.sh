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

# --- Main orchestration ---

main() {
  local script_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  local project_dir="$script_dir/.."
  local version_file="${RELEASE_VERSION_FILE:-$project_dir/VERSION}"
  local yes_flag=false

  # Parse args
  for arg in "$@"; do
    case "$arg" in
      --yes|-y) yes_flag=true ;;
    esac
  done

  # Preflight checks
  echo "Running preflight checks..."

  check_clean_tree
  echo "  ✓ Working tree is clean"

  check_main_branch
  echo "  ✓ On main branch"

  local version
  version="$(read_version "$version_file")"
  echo "  ✓ Version: $version"

  local tag="v$version"
  check_tag_not_exists "$tag"
  echo "  ✓ Tag $tag does not exist"

  check_gh_auth
  echo "  ✓ gh CLI authenticated"

  echo ""

  # Confirmation
  if [[ "$yes_flag" != true ]]; then
    echo "Release $tag?"
    echo "  - Create annotated tag $tag"
    echo "  - Push to origin"
    echo "  - Create GitHub release with binaries attached"
    echo ""
    read -rp "Proceed? [y/N] " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
      echo "Aborted."
      exit 0
    fi
  fi

  # Build ghost-tab-tui binaries
  echo "Building ghost-tab-tui binaries..."
  local arm64_bin amd64_bin
  arm64_bin="$(mktemp)"
  amd64_bin="$(mktemp)"
  trap 'rm -f "$arm64_bin" "$amd64_bin"' EXIT

  (cd "$project_dir" && GOOS=darwin GOARCH=arm64 go build -o "$arm64_bin" ./cmd/ghost-tab-tui) || {
    echo "Error: failed to build ghost-tab-tui for arm64" >&2; exit 1
  }
  (cd "$project_dir" && GOOS=darwin GOARCH=amd64 go build -o "$amd64_bin" ./cmd/ghost-tab-tui) || {
    echo "Error: failed to build ghost-tab-tui for amd64" >&2; exit 1
  }
  echo "  ✓ Built ghost-tab-tui for darwin/arm64 and darwin/amd64"

  # Tag and push
  echo ""
  echo "Creating tag $tag..."
  git tag -a "$tag" -m "Release $tag"
  echo "Pushing to origin..."
  git push origin main --tags

  # Create GitHub release
  echo "Creating GitHub release..."
  local prev_tag
  prev_tag="$(git describe --tags --abbrev=0 "${tag}^" 2>/dev/null || echo "")"
  local notes
  notes="$(bash "$script_dir/generate-release-notes.sh" "$prev_tag" "$tag")"
  gh release create "$tag" --notes "$notes" \
    "${arm64_bin}#ghost-tab-tui-darwin-arm64" \
    "${amd64_bin}#ghost-tab-tui-darwin-amd64"

  echo ""
  echo "✓ Release $tag complete!"
  echo "  GitHub: https://github.com/JackUait/ghost-tab/releases/tag/$tag"
  echo "  Binaries: ghost-tab-tui-darwin-arm64, ghost-tab-tui-darwin-amd64"
}

# Only run main when executed directly (not sourced for testing)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
