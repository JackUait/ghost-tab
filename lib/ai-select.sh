#!/bin/bash
# AI tool multi-select menu and preference saving.
# Depends on: tui.sh (color variables)

# Save the first selected AI tool to the preference file.
# Priority: claude > codex > copilot > opencode.
# Usage: save_ai_tool_preference <sel_claude> <sel_codex> <sel_copilot> <sel_opencode> <pref_dir>
save_ai_tool_preference() {
  local sel_claude="$1" sel_codex="$2" sel_copilot="$3" sel_opencode="$4" pref_dir="$5"
  local tool=""

  if [ "$sel_claude" -eq 1 ]; then
    tool="claude"
  elif [ "$sel_codex" -eq 1 ]; then
    tool="codex"
  elif [ "$sel_copilot" -eq 1 ]; then
    tool="copilot"
  elif [ "$sel_opencode" -eq 1 ]; then
    tool="opencode"
  fi

  if [ -n "$tool" ]; then
    mkdir -p "$pref_dir"
    echo "$tool" > "$pref_dir/ai-tool"
  fi
}

# Interactive multi-select checkbox menu for AI tools.
# Sets globals: _sel_claude, _sel_codex, _sel_copilot, _sel_opencode
# Usage: run_ai_tool_select <cc_installed> <codex_installed> <copilot_installed> <oc_installed>
run_ai_tool_select() {
  local _cc_installed="$1" _codex_installed="$2" _copilot_installed="$3" _oc_installed="$4"

  # Default selections: always pre-check Claude, pre-check others if installed
  _sel_claude=1
  _sel_codex=$_codex_installed
  _sel_copilot=$_copilot_installed
  _sel_opencode=$_oc_installed

  local _selecting=1 _cursor=0 _drawn="" _i _name _sel _tag _key _s1 _s2

  while [ "$_selecting" -eq 1 ]; do
    [ -n "$_drawn" ] && printf '\033[5A'
    _drawn=1

    for _i in 0 1 2 3; do
      case $_i in
        0) _name="Claude Code"; _sel=$_sel_claude; _tag="" ;;
        1) _name="Codex CLI (OpenAI)"; _sel=$_sel_codex; _tag="" ;;
        2) _name="Copilot CLI (GitHub)"; _sel=$_sel_copilot; _tag="" ;;
        3) _name="OpenCode (anomalyco)"; _sel=$_sel_opencode; _tag="" ;;
      esac
      case $_i in
        0) [ "$_cc_installed" -eq 1 ] && _tag=" ${_YELLOW}(installed)${_NC}" ;;
        1) [ "$_codex_installed" -eq 1 ] && _tag=" ${_YELLOW}(installed)${_NC}" ;;
        2) [ "$_copilot_installed" -eq 1 ] && _tag=" ${_YELLOW}(installed)${_NC}" ;;
        3) [ "$_oc_installed" -eq 1 ] && _tag=" ${_YELLOW}(installed)${_NC}" ;;
      esac

      if [ "$_i" -eq "$_cursor" ]; then
        if [ "$_sel" -eq 1 ]; then
          echo -e "  ${_BOLD}❯ [x] ${_name}${_NC}${_tag}"
        else
          echo -e "  ${_BOLD}❯ [ ] ${_name}${_NC}${_tag}"
        fi
      else
        if [ "$_sel" -eq 1 ]; then
          echo -e "    [x] ${_name}${_tag}"
        else
          echo -e "    [ ] ${_name}${_tag}"
        fi
      fi
    done
    echo -e "  ${_BLUE}↑↓${_NC} navigate  ${_BLUE}Space${_NC} toggle  ${_BLUE}Enter${_NC} confirm"

    read -rsn1 _key </dev/tty
    if [[ "$_key" == $'\x1b' ]]; then
      read -rsn1 _s1 </dev/tty
      if [[ "$_s1" == "[" ]]; then
        read -rsn1 _s2 </dev/tty
        case "$_s2" in
          A) _cursor=$(( (_cursor - 1 + 4) % 4 )) ;;
          B) _cursor=$(( (_cursor + 1) % 4 )) ;;
        esac
      fi
    elif [[ "$_key" == " " ]]; then
      case $_cursor in
        0) _sel_claude=$(( 1 - _sel_claude )) ;;
        1) _sel_codex=$(( 1 - _sel_codex )) ;;
        2) _sel_copilot=$(( 1 - _sel_copilot )) ;;
        3) _sel_opencode=$(( 1 - _sel_opencode )) ;;
      esac
    elif [[ "$_key" == "" ]]; then
      if [ $(( _sel_claude + _sel_codex + _sel_copilot + _sel_opencode )) -eq 0 ]; then
        echo -e "  ${_RED}✗${_NC} Select at least one AI tool"
        sleep 0.8
        printf '\033[1A\033[K'
      else
        _selecting=0
      fi
    fi
  done
}
