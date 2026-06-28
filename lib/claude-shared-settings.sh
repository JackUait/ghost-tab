#!/bin/bash
# Claude shared-settings helper — pure, no side effects on source.
# A native Claude "account" is isolated by its own CLAUDE_CONFIG_DIR (see
# lib/claude-accounts.sh), so by default a non-Default login sees NONE of the
# standard ~/.claude login's settings: status line, permission mode, skills,
# subagents, slash commands, hooks, model, keybindings, plugins — all gone.
#
# sync_claude_shared_settings symlinks a curated allowlist of *settings* items
# from the standard login's config dir into a per-account CLAUDE_CONFIG_DIR, so
# every login shares ONE set of settings while each keeps its own credentials,
# identity (.claude.json), and session/runtime state. Symlinks (not copies) mean
# editing a setting once — in the standard login — propagates to all accounts,
# and re-running at every launch self-heals any drift (e.g. if Claude rewrote a
# settings file in place, severing the link).

# Items shared across logins. Each is a settings/customization artifact, never
# credentials or per-login state. Files that don't exist in the source are
# skipped, so this list can name everything Claude might use without harm.
WISP_DECK_CLAUDE_SHARED_ITEMS=(
  settings.json          # permissions (incl. defaultMode), model, hooks, env, statusLine, plugins
  settings.local.json    # machine-local permission overrides
  CLAUDE.md              # global user memory
  keybindings.json       # custom key bindings
  skills                 # custom skills
  commands               # custom slash commands
  agents                 # custom subagents
  plugins                # installed plugins + marketplace config
  statusline-wrapper.sh  # status line entrypoint referenced by settings.json
  statusline-command.sh  # status line helpers
  statusline-helpers.sh
  subagent-statusline.sh         # subagent panel row entrypoint (settings.json)
  subagent-statusline-helpers.sh # subagent panel row renderer
  tab-spinner-start.sh   # notification/tab hooks referenced by settings.json
  tab-spinner-stop.sh
)

# sync_claude_shared_settings <source_dir> <account_dir>
# Link every existing shared item from source_dir into account_dir, replacing any
# account-local copy with a symlink to the source. No-op (exit 0) when either dir
# is missing or when the two are the same path (the Default login uses the source
# dir directly and must be left untouched). Never reads, writes, or removes any
# item outside the allowlist, so credentials and session state are safe.
sync_claude_shared_settings() {
  local source_dir="$1" account_dir="$2" item src dest
  [ -n "$source_dir" ] && [ -n "$account_dir" ] || return 0
  [ -d "$source_dir" ] && [ -d "$account_dir" ] || return 0
  [ "$source_dir" = "$account_dir" ] && return 0

  for item in "${WISP_DECK_CLAUDE_SHARED_ITEMS[@]}"; do
    src="$source_dir/$item"
    dest="$account_dir/$item"
    [ -e "$src" ] || continue
    # Drop any existing account-local file/dir/symlink (removing a symlink-to-dir
    # deletes only the link, never the shared target), then point at the source.
    rm -rf "$dest"
    ln -sfn "$src" "$dest"
  done
}
