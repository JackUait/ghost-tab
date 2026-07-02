# Multi-select & batch discard in the changeset ledger

**Date:** 2026-07-02
**Component:** `lib/compact-view.sh` (the compact-view "changeset ledger")

## Goal

Let the user select several files in the file list at once and discard their
working-tree changes together, instead of opening each file's diff popup and
discarding it one at a time.

## Background

`compact_view` renders a scrollable ledger of tracked working-tree changes
(staged + modified groups). Today:

- Mouse **hover** highlights the file row under the cursor (`hover_line`).
- **Left-click** opens that file's full-window diff popup.
- Discard exists only *inside* that popup: a single-file `[ Discard ]` button →
  Yes/No → the bash caller runs `git restore -- <file>` (`discard_worktree_file`).

There is no way to act on more than one file at a time.

## Decisions

- **Selection input:** keyboard toggle on the hovered row (user's choice).
  Hover a row with the mouse, press **`x`** to toggle it selected/unselected.
  (`Space` is already page-down; `x` is free and reads as "mark".)
- **Discard trigger:** press **`d`** to *arm* an inline confirmation; the footer
  shows `Discard N file(s)? [y/n]`. **`y`** discards, **`n`**/**`Esc`**/**`d`**
  cancels. This mirrors the diff popup's existing armed-confirm pattern — the op
  is irreversible, so it stays guarded. (Chosen as best-judgment default while
  the user was away; matches the recommended option.)
- **Fallback:** if nothing is selected, `d` acts on the currently-hovered file
  (single-file discard from the ledger). If nothing is selected and nothing is
  hovered, `d` is a no-op.

## Selection is tracked by PATH, not line index

The ledger rebuilds every ~2s and scrolls, so body-line indices are unstable.
Selection is stored as a newline-delimited set of **file paths** (`SELECTED`),
which survives rebuilds and scrolling. After each rebuild the set is pruned to
paths that still exist in the changeset (a file that was discarded, committed,
or reverted externally drops out of the selection).

## New pure functions (each unit-tested first, TDD)

All are pure string helpers so they test without a PTY:

1. `toggle_selection <selected> <path>` — echo the new set with `path` removed if
   present, else appended. Membership is whole-line exact match.
2. `selection_contains <selected> <path>` — exit 0 if `path` is a member.
3. `selection_count <selected>` — echo the number of non-empty members.
4. `prune_selection <selected> <valid_newline_paths>` — echo `selected` minus any
   path not present in the valid set.
5. `apply_selection_markers <body> <body_map> <selected>` — for every body line
   whose mapped path is in `selected`, replace the row's 3-space indent with a
   colored ` ✓ ` marker (exactly 3 visible columns, so column alignment and the
   hover-highlight width math are unaffected). Non-file rows and unselected rows
   pass through unchanged.
6. `discard_prompt <count>` — echo the footer confirm string
   `Discard N file(s)? [y/n]` (singular "file" when count == 1).
7. `discard_worktree_files <dir> <selected>` — `git restore --` each member path;
   reuses `discard_worktree_file`. Returns non-zero if any restore fails.

## Loop wiring (`compact_view`)

New loop-scope state: `SELECTED=""`, `discard_armed=0`.

- On rebuild (`need_build`): `SELECTED=$(prune_selection "$SELECTED" "<paths from body_map>")`.
- Render order: build `body` → `apply_selection_markers` → `highlight_body_line`
  (hover). When `discard_armed`, the bottom line shows `discard_prompt` (in place
  of / alongside the scroll status).
- Key `x`: if the hovered row maps to a path, `SELECTED=$(toggle_selection ...)`,
  request redraw.
- Key `d`:
  - not armed: compute the discard set (SELECTED, else the hovered path). If
    non-empty, `discard_armed=1`; else no-op. Redraw.
  - (armed handled below)
- When `discard_armed`:
  - `y`: `discard_worktree_files "$project_dir" "$discard_set"`, clear `SELECTED`,
    `discard_armed=0`, `need_build=1`.
  - `n` / `Esc` / anything else: `discard_armed=0`, redraw.

The `x`/`d`/`y`/`n` keys are added to `handle_key`; none collide with existing
bindings (`j k b Space g G` + arrows/mouse). The hover hot path (mouse motion)
is untouched, so the fork-free tripwire (`body_line_for_click`, `nth_line`,
`set_hover_from_row`) still holds.

## Testing

- **Unit (TDD):** one Go bash test per new pure function — toggle add/remove,
  contains true/false, count, prune drops-missing/keeps-present, markers on
  selected file rows only (and idempotent visible width), prompt singular/plural,
  `discard_worktree_files` reverts multiple files and reports failure.
- **Integration (PTY):** drive the real loop over a pty — hover row A (`x`),
  hover row B (`x`), `d`, `y` — then assert both files reverted to HEAD and the
  unselected file is untouched.
- `shellcheck lib/compact-view.sh` clean.
- Full `./run-tests.sh` green.

## Out of scope

- Mouse-checkbox selection, a keyboard navigation cursor independent of hover,
  select-all, discarding staged-only copies, and touching the diff popup's
  single-file discard (left as-is).
