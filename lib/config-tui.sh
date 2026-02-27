#!/bin/bash
# Config menu TUI dispatcher
# Uses ghost-tab-tui config-menu subcommand in a loop

# Source dependencies if not already loaded
_config_tui_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/tui.sh
[ "$(type -t header 2>/dev/null)" = "function" ] || source "$_config_tui_dir/tui.sh"
# shellcheck source=lib/terminal-select-tui.sh
[ "$(type -t select_terminal_interactive 2>/dev/null)" = "function" ] || source "$_config_tui_dir/terminal-select-tui.sh"
# shellcheck source=lib/terminals/registry.sh
[ "$(type -t get_terminal_display_name 2>/dev/null)" = "function" ] || source "$_config_tui_dir/terminals/registry.sh"
# shellcheck source=lib/project-actions-tui.sh
[ "$(type -t add_project_interactive 2>/dev/null)" = "function" ] || source "$_config_tui_dir/project-actions-tui.sh"
# shellcheck source=lib/project-actions.sh
[ "$(type -t add_project_to_file 2>/dev/null)" = "function" ] || source "$_config_tui_dir/project-actions.sh"
# shellcheck source=lib/ai-select-tui.sh
[ "$(type -t select_ai_tool_interactive 2>/dev/null)" = "function" ] || source "$_config_tui_dir/ai-select-tui.sh"
# shellcheck source=lib/settings-menu-tui.sh
[ "$(type -t settings_menu_interactive 2>/dev/null)" = "function" ] || source "$_config_tui_dir/settings-menu-tui.sh"

# Interactive config menu loop.
# Calls ghost-tab-tui config-menu, dispatches on action, loops until quit.
config_menu_interactive() {
  if ! command -v ghost-tab-tui &>/dev/null; then
    error "ghost-tab-tui binary not found. Please reinstall."
    return 1
  fi

  local config_dir="${XDG_CONFIG_HOME:-$HOME/.config}/ghost-tab"

  while true; do
    local result
    if ! result=$(ghost-tab-tui config-menu 2>/dev/null); then
      return 1
    fi

    local action
    if ! action=$(echo "$result" | jq -r '.action' 2>/dev/null); then
      error "Failed to parse config menu response"
      return 1
    fi

    case "$action" in
      manage-terminals)
        export GHOST_TAB_TERMINAL_PREF="$config_dir/terminal"
        if select_terminal_interactive; then
          # shellcheck disable=SC2154
          echo "$_selected_terminal" > "$config_dir/terminal"
          success "Terminal set to $(get_terminal_display_name "$_selected_terminal")"
          echo ""
          read -rsn1 -p "Press any key to continue..." </dev/tty
        fi
        ;;
      manage-projects)
        if add_project_interactive; then
          # shellcheck disable=SC2154
          add_project_to_file "$_add_project_name" "$_add_project_path" "$config_dir/projects"
          success "Added project: $_add_project_name"
          echo ""
          read -rsn1 -p "Press any key to continue..." </dev/tty
        fi
        ;;
      select-ai-tools)
        if select_ai_tool_interactive; then
          # shellcheck disable=SC2154
          echo "$_selected_ai_tool" > "$config_dir/ai-tool"
          success "Default AI tool set to $_selected_ai_tool"
          echo ""
          read -rsn1 -p "Press any key to continue..." </dev/tty
        fi
        ;;
      display-settings)
        settings_menu_interactive
        ;;
      reinstall)
        local script_dir
        script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
        if [ -f "$script_dir/bin/ghost-tab" ]; then
          exec bash "$script_dir/bin/ghost-tab"
        else
          error "Installer not found. Re-clone the repository."
          echo ""
          read -rsn1 -p "Press any key to continue..." </dev/tty
        fi
        ;;
      quit|"")
        return 0
        ;;
      *)
        error "Unknown action: $action"
        ;;
    esac
  done
}
