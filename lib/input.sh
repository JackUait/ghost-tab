#!/bin/bash
# Input parsing helpers — no side effects on source.

# Parses an escape sequence from stdin (bytes AFTER the initial \x1b).
# Outputs to stdout:
#   "A"/"B"/"C"/"D" for arrow keys
#   "click:ROW" for SGR mouse left-click press
#   "" (empty) for ignored events (release, non-left-click)
parse_esc_sequence() {
  local _b1 _b2 _mc _mouse_data _mouse_btn _mouse_rest _mouse_col _mouse_row

  read -rsn1 _b1
  if [[ "$_b1" == "[" ]]; then
    read -rsn1 _b2
    if [[ "$_b2" == "<" ]]; then
      # SGR mouse: read until M (press) or m (release)
      _mouse_data=""
      while true; do
        read -rsn1 _mc
        if [[ "$_mc" == "M" || "$_mc" == "m" ]]; then
          break
        fi
        _mouse_data="${_mouse_data}${_mc}"
      done
      # Only handle press (M), ignore release (m)
      if [[ "$_mc" == "M" ]]; then
        _mouse_btn="${_mouse_data%%;*}"
        _mouse_rest="${_mouse_data#*;}"
        _mouse_col="${_mouse_rest%%;*}"
        _mouse_row="${_mouse_rest##*;}"
        if [[ "$_mouse_btn" == "0" ]]; then
          echo "click:${_mouse_row}"
        fi
      fi
    else
      echo "$_b2"
    fi
  fi
}
