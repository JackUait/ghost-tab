#!/bin/bash
# Claude config helpers — pure, no side effects on source.
# A "config" is a settings JSON launched via `claude --settings <file>`.
# Storage: <root>/claude-configs/<file>.json, named in <root>/claude-configs.list
# (name:file), with active filename in <root>/claude-config.

# load_claude_configs <list_file> — prints valid name:file lines (skips blanks/comments).
load_claude_configs() {
  local file="$1" line
  [ ! -f "$file" ] && return 0
  while IFS= read -r line; do
    [[ -z "$line" || "$line" == \#* ]] && continue
    echo "$line"
  done < "$file"
}

# get_active_claude_config <pointer_file> — prints active filename or empty.
get_active_claude_config() {
  local file="$1" line
  [ -f "$file" ] || return 0
  IFS= read -r line < "$file" || true
  line="${line//[[:space:]]/}"
  [ "$line" = "standard" ] && return 0
  printf '%s\n' "$line"
}

# set_active_claude_config <pointer_file> <filename> — empty/standard removes the file.
set_active_claude_config() {
  local file="$1" filename="$2"
  if [ -z "$filename" ] || [ "$filename" = "standard" ]; then
    rm -f "$file"
    return 0
  fi
  mkdir -p "$(dirname "$file")"
  printf '%s\n' "$filename" > "$file"
}

# resolve_claude_config_path <configs_dir> <pointer_file> — abs path iff active file exists.
resolve_claude_config_path() {
  local configs_dir="$1" pointer_file="$2" active
  active="$(get_active_claude_config "$pointer_file")"
  [ -z "$active" ] && return 0
  local path="$configs_dir/$active"
  [ -f "$path" ] && printf '%s\n' "$path"
}

# slugify <name> — lowercase, non-alnum to single dashes, trimmed.
slugify() {
  local s="$1"
  s="$(printf '%s' "$s" | tr '[:upper:]' '[:lower:]')"
  s="$(printf '%s' "$s" | tr -c 'a-z0-9' '-')"
  s="$(printf '%s' "$s" | tr -s '-')"
  s="${s#-}"
  s="${s%-}"
  printf '%s' "$s"
}

# add_claude_config <list_file> <configs_dir> <name> — creates <slug>.json ({}), appends
# name:file to list, prints filename. Resolves filename collisions with -2, -3, ...
add_claude_config() {
  local list_file="$1" configs_dir="$2" name="$3"
  local slug base file n
  slug="$(slugify "$name")"
  [ -z "$slug" ] && slug="config"
  base="$slug"
  file="$base.json"
  n=2
  while [ -e "$configs_dir/$file" ]; do
    file="$base-$n.json"
    n=$((n + 1))
  done
  mkdir -p "$configs_dir"
  printf '{}\n' > "$configs_dir/$file"
  mkdir -p "$(dirname "$list_file")"
  printf '%s:%s\n' "$name" "$file" >> "$list_file"
  printf '%s' "$file"
}

# rename_claude_config <list_file> <file> <new_name> — rewrites the matching line's name.
# Returns 1 if no line in the list matches <file>.
rename_claude_config() {
  local list_file="$1" file="$2" new_name="$3" line f tmp found
  [ -f "$list_file" ] || return 0
  found=0
  tmp="$(mktemp)"
  while IFS= read -r line; do
    f="${line#*:}"
    if [ "$f" = "$file" ]; then
      found=1
      printf '%s:%s\n' "$new_name" "$file" >> "$tmp"
    else
      printf '%s\n' "$line" >> "$tmp"
    fi
  done < "$list_file"
  if [ "$found" -eq 0 ]; then
    rm -f "$tmp"
    return 1
  fi
  mv "$tmp" "$list_file"
}

# delete_claude_config <list_file> <configs_dir> <pointer_file> <file> — remove file + line;
# clear pointer if it was active.
delete_claude_config() {
  local list_file="$1" configs_dir="$2" pointer_file="$3" file="$4" line f tmp active
  rm -f "$configs_dir/$file"
  if [ -f "$list_file" ]; then
    tmp="$(mktemp)"
    while IFS= read -r line; do
      f="${line#*:}"
      [ "$f" = "$file" ] && continue
      printf '%s\n' "$line" >> "$tmp"
    done < "$list_file"
    if ! mv "$tmp" "$list_file"; then
      rm -f "$tmp"
      return 1
    fi
  fi
  active="$(get_active_claude_config "$pointer_file")"
  [ "$active" = "$file" ] && set_active_claude_config "$pointer_file" ""
}
