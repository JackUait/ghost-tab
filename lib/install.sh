#!/bin/bash
# Package installation helpers for the installer.

# Detect CPU architecture: outputs "arm64" or "x86_64"
detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    arm64)   echo "arm64" ;;
    x86_64)  echo "x86_64" ;;
    *)
      error "Unsupported architecture: $arch"
      return 1 ;;
  esac
}

# Get the latest release tag from a GitHub repo (e.g. "v1.2.3")
# Uses the /releases/latest redirect — no API key required.
get_latest_release_tag() {
  local repo="$1" tag
  tag="$(curl -fsSI "https://github.com/$repo/releases/latest" 2>/dev/null \
    | grep -i '^location:' \
    | sed 's|.*/tag/||' \
    | tr -d '[:space:]\r')"
  if [[ -z "$tag" ]]; then
    error "Failed to fetch release tag for $repo"
    return 1
  fi
  echo "$tag"
}

# Download a binary from $url to $dest and make it executable.
install_binary() {
  local url="$1" dest="$2" display_name="$3"
  info "Downloading $display_name..."
  mkdir -p "$(dirname "$dest")"
  if curl -fsSL -o "$dest" "$url"; then
    chmod +x "$dest"
    success "$display_name installed"
  else
    warn "Failed to download $display_name from $url"
    return 1
  fi
}

# Install jq from jqlang/jq GitHub releases.
ensure_jq() {
  if command -v jq &>/dev/null; then
    success "jq already installed"
    return 0
  fi
  local arch jq_arch
  arch="$(detect_arch)" || return 1
  case "$arch" in
    arm64)   jq_arch="macos-arm64" ;;
    x86_64)  jq_arch="macos-amd64" ;;
  esac
  install_binary \
    "https://github.com/jqlang/jq/releases/latest/download/jq-${jq_arch}" \
    "$HOME/.local/bin/jq" \
    "jq"
}

# Install tmux from tmux/tmux-builds GitHub releases.
ensure_tmux() {
  if command -v tmux &>/dev/null; then
    success "tmux already installed"
    return 0
  fi
  local arch tag version tmp_dir url
  arch="$(detect_arch)" || return 1
  tag="$(get_latest_release_tag "tmux/tmux-builds")" || return 1
  version="${tag#v}"
  tmp_dir="$(mktemp -d)"
  # shellcheck disable=SC2064
  trap "rm -rf '$tmp_dir'" RETURN
  url="https://github.com/tmux/tmux-builds/releases/download/${tag}/tmux-${version}-macos-${arch}.tar.gz"
  info "Downloading tmux..."
  if curl -fsSL -o "$tmp_dir/tmux.tar.gz" "$url"; then
    tar -xzf "$tmp_dir/tmux.tar.gz" -C "$tmp_dir" tmux
    mkdir -p "$HOME/.local/bin"
    mv "$tmp_dir/tmux" "$HOME/.local/bin/tmux"
    chmod +x "$HOME/.local/bin/tmux"
    success "tmux installed"
  else
    warn "Failed to install tmux"
    return 1
  fi
}

# Install lazygit from jesseduffield/lazygit GitHub releases.
ensure_lazygit() {
  if command -v lazygit &>/dev/null; then
    success "lazygit already installed"
    return 0
  fi
  local arch tag version tmp_dir url
  arch="$(detect_arch)" || return 1
  tag="$(get_latest_release_tag "jesseduffield/lazygit")" || return 1
  version="${tag#v}"
  tmp_dir="$(mktemp -d)"
  # shellcheck disable=SC2064
  trap "rm -rf '$tmp_dir'" RETURN
  url="https://github.com/jesseduffield/lazygit/releases/download/${tag}/lazygit_${version}_darwin_${arch}.tar.gz"
  info "Downloading lazygit..."
  if curl -fsSL -o "$tmp_dir/lazygit.tar.gz" "$url"; then
    tar -xzf "$tmp_dir/lazygit.tar.gz" -C "$tmp_dir" lazygit
    mkdir -p "$HOME/.local/bin"
    mv "$tmp_dir/lazygit" "$HOME/.local/bin/lazygit"
    chmod +x "$HOME/.local/bin/lazygit"
    success "lazygit installed"
  else
    warn "Failed to install lazygit"
    return 1
  fi
}

# Install broot from Canop/broot GitHub releases.
ensure_broot() {
  if command -v broot &>/dev/null; then
    success "broot already installed"
    return 0
  fi
  local arch broot_arch tag version tmp_dir url
  arch="$(detect_arch)" || return 1
  case "$arch" in
    arm64)   broot_arch="aarch64-apple-darwin" ;;
    x86_64)  broot_arch="x86_64-apple-darwin" ;;
  esac
  tag="$(get_latest_release_tag "Canop/broot")" || return 1
  version="${tag#v}"
  tmp_dir="$(mktemp -d)"
  # shellcheck disable=SC2064
  trap "rm -rf '$tmp_dir'" RETURN
  url="https://github.com/Canop/broot/releases/download/${tag}/broot_${version}.zip"
  info "Downloading broot..."
  if curl -fsSL -o "$tmp_dir/broot.zip" "$url"; then
    unzip -q -d "$tmp_dir" "$tmp_dir/broot.zip"
    mkdir -p "$HOME/.local/bin"
    mv "$tmp_dir/build/${broot_arch}/broot" "$HOME/.local/bin/broot"
    chmod +x "$HOME/.local/bin/broot"
    success "broot installed"
  else
    warn "Failed to install broot"
    return 1
  fi
}

# Install ghost-tab-tui by downloading the pre-built binary from the ghost-tab release.
# Args: share_dir (to read VERSION from)
ensure_ghost_tab_tui() {
  local share_dir="$1"

  if command -v ghost-tab-tui &>/dev/null; then
    success "ghost-tab-tui already available"
    return 0
  fi

  local version arch url
  version="$(tr -d '[:space:]' < "$share_dir/VERSION" 2>/dev/null)"
  if [[ -z "$version" ]]; then
    error "Cannot determine ghost-tab-tui version: VERSION file missing in $share_dir"
    return 1
  fi

  arch="$(detect_arch)" || return 1
  url="https://github.com/JackUait/ghost-tab/releases/download/v${version}/ghost-tab-tui-darwin-${arch}"

  mkdir -p "$HOME/.local/bin"
  install_binary "$url" "$HOME/.local/bin/ghost-tab-tui" "ghost-tab-tui" || return 1
}

# Install base CLI requirements.
ensure_base_requirements() {
  ensure_jq
  ensure_tmux
  ensure_lazygit
  ensure_broot
}

# Install a command-line tool if not already on PATH.
# Usage: ensure_command "cmd" "install_cmd" "post_msg" "display_name"
ensure_command() {
  local cmd="$1" install_cmd="$2" post_msg="$3" display_name="$4"
  if command -v "$cmd" &>/dev/null; then
    success "$display_name already installed"
  else
    info "Installing $display_name..."
    if eval "$install_cmd"; then
      success "$display_name installed"
      [ -n "$post_msg" ] && info "$post_msg"
    else
      warn "$display_name installation failed — install manually: $install_cmd"
    fi
  fi
}
