# Plain Terminal Menu Option

Adds a "Plain terminal" option as the last item in the new-tab project selection menu.

## Behavior

When selected, the wrapper script exits entirely and replaces itself with a bare bash shell via `exec bash`. No tmux session, no project selection, no tools. The tab becomes a normal terminal.

## Menu Position

Last item, after "Open once":

```
  · Add new project
  · Delete a project
  · Open once
  · Plain terminal          ← dim style
```

## Implementation

Two changes in `setup.sh`:

1. Append menu item arrays with label `"Plain terminal"`, type `"plain"`, dim highlight style.
2. Add `plain)` case in selection handler that restores terminal state and runs `exec bash`.
