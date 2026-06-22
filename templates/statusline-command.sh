#!/bin/bash
input=$(cat)
cwd=$(echo "$input" | sed -n 's/.*"current_dir":"\([^"]*\)".*/\1/p')

printf '\033[01;36m%s\033[00m' "$(basename "$cwd")"
