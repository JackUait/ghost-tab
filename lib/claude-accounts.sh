#!/bin/bash
# Claude account helpers — pure, no side effects on source.
# An "account" is a native Claude login isolated by its own CLAUDE_CONFIG_DIR,
# so multiple subscriptions (work, personal) can stay logged in at once and be
# switched between by relaunching `claude` under a different config dir.
# Storage: <root>/claude-accounts/<dir>/ holds the login, named in
# <root>/claude-accounts.list (label:dir), with the active dir in
# <root>/claude-account. The Default account (absent/"default" pointer) is the
# user's standard ~/.claude login (Keychain) — no CLAUDE_CONFIG_DIR is set.

# load_claude_accounts <list_file> — prints valid label:dir lines (skips blanks/comments).
load_claude_accounts() {
  local file="$1" line
  [ ! -f "$file" ] && return 0
  while IFS= read -r line; do
    [[ -z "$line" || "$line" == \#* ]] && continue
    echo "$line"
  done < "$file"
}

# get_active_claude_account <pointer_file> — prints active dir name or empty (Default).
get_active_claude_account() {
  local file="$1" line
  [ -f "$file" ] || return 0
  IFS= read -r line < "$file" || true
  line="${line//[[:space:]]/}"
  [ "$line" = "default" ] && return 0
  printf '%s\n' "$line"
}

# get_active_claude_account_name <pointer_file> <list_file> — prints the active
# account's display label, mirroring the menu's ACCOUNT row. Default (no active
# pointer) and any dir not found in the list read as "Default".
get_active_claude_account_name() {
  local pointer_file="$1" list_file="$2" active label dir
  active="$(get_active_claude_account "$pointer_file")"
  if [ -n "$active" ] && [ -f "$list_file" ]; then
    while IFS=: read -r label dir; do
      [[ -z "$label" || "$label" == \#* ]] && continue
      if [ "$dir" = "$active" ]; then
        printf '%s\n' "$label"
        return 0
      fi
    done < "$list_file"
  fi
  printf 'Default\n'
}

# set_active_claude_account <pointer_file> <dir> — empty/default removes the file.
set_active_claude_account() {
  local file="$1" dirname="$2"
  if [ -z "$dirname" ] || [ "$dirname" = "default" ]; then
    rm -f "$file"
    return 0
  fi
  mkdir -p "$(dirname "$file")"
  printf '%s\n' "$dirname" > "$file"
}

# resolve_claude_account_dir <accounts_dir> <pointer_file> — abs path iff the
# active account directory exists. Default (empty pointer) resolves to empty so
# the launcher leaves CLAUDE_CONFIG_DIR unset and Claude uses the Keychain login.
resolve_claude_account_dir() {
  local accounts_dir="$1" pointer_file="$2" active
  active="$(get_active_claude_account "$pointer_file")"
  [ -z "$active" ] && return 0
  local path="$accounts_dir/$active"
  [ -d "$path" ] && printf '%s\n' "$path"
}

# add_claude_account [config_root] — interactive: prompt for a label, register
# the account via the ghost-tab-tui CLI, run `claude auth login` under its
# isolated CLAUDE_CONFIG_DIR (browser OAuth), then make it the active account.
# Adding a native login needs an interactive terminal, so wrapper.sh runs this
# between menu loops (after the menu exits with the "add-account" action).
add_claude_account() {
  local cfg_root="${1:-${XDG_CONFIG_HOME:-$HOME/.config}/ghost-tab}"
  local accounts_dir="$cfg_root/claude-accounts"
  local list_file="$cfg_root/claude-accounts.list"
  local pointer_file="$cfg_root/claude-account"

  printf '\nAdd a Claude login\n'
  local label=""
  printf 'Label (e.g. Work, Personal): '
  IFS= read -r label || true
  label="${label#"${label%%[![:space:]]*}"}" # ltrim
  label="${label%"${label##*[![:space:]]}"}" # rtrim
  if [ -z "$label" ]; then
    printf 'Cancelled.\n'
    return 1
  fi

  local dir
  if ! dir="$(ghost-tab-tui claude-account add --list "$list_file" --accounts-dir "$accounts_dir" --label "$label")"; then
    printf 'Failed to register account.\n' >&2
    return 1
  fi
  dir="${dir//[[:space:]]/}"
  if [ -z "$dir" ]; then
    printf 'Failed to register account.\n' >&2
    return 1
  fi

  printf "Opening Claude login for '%s'…\n" "$label"
  CLAUDE_CONFIG_DIR="$accounts_dir/$dir" claude auth login || \
    printf 'Login did not complete; you can retry by switching to this account later.\n' >&2

  # Make the new account active so the next launch uses it.
  set_active_claude_account "$pointer_file" "$dir"
  printf "Added login '%s'.\n" "$label"
  return 0
}

# Other mutations (remove) live in Go — the single source of truth — exposed as
# `ghost-tab-tui claude-account <action>`. See internal/claudeaccount.
