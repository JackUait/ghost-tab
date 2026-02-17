package models

import (
	"os/exec"
	"strings"
)

// Worktree represents a git worktree entry.
type Worktree struct {
	Path   string
	Branch string
}

// ParseWorktreeListPorcelain parses the output of `git worktree list --porcelain`
// and returns only non-main worktrees.
func ParseWorktreeListPorcelain(output string) []Worktree {
	if output == "" {
		return nil
	}

	var all []Worktree
	blocks := strings.Split(strings.TrimRight(output, "\n"), "\n\n")

	for _, block := range blocks {
		if block == "" {
			continue
		}
		var wt Worktree
		for _, line := range strings.Split(block, "\n") {
			if strings.HasPrefix(line, "worktree ") {
				wt.Path = strings.TrimPrefix(line, "worktree ")
			} else if strings.HasPrefix(line, "branch ") {
				ref := strings.TrimPrefix(line, "branch ")
				wt.Branch = strings.TrimPrefix(ref, "refs/heads/")
			} else if line == "detached" {
				wt.Branch = "(detached)"
			}
		}
		all = append(all, wt)
	}

	// First entry is always the main worktree
	if len(all) <= 1 {
		return nil
	}

	// First is main worktree, return the rest
	return all[1:]
}

// PopulateWorktrees runs DetectWorktrees for each project and attaches
// the results to the Worktrees field.
func PopulateWorktrees(projects []Project) {
	for i := range projects {
		projects[i].Worktrees = DetectWorktrees(projects[i].Path)
	}
}

// DetectWorktrees runs `git worktree list --porcelain` for the given path
// and returns non-main worktrees. Returns nil on any error.
func DetectWorktrees(projectPath string) []Worktree {
	cmd := exec.Command("git", "-C", projectPath, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return ParseWorktreeListPorcelain(string(out))
}

// ParseBranchList parses the output of `git branch -a --format='%(refname:short)'`
// and returns a deduplicated list. When a branch exists both locally and as
// origin/<name>, only the local name is kept. HEAD refs are removed.
func ParseBranchList(output string) []string {
	if output == "" {
		return nil
	}

	localSet := make(map[string]bool)
	var remoteOnly []string

	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		branch := strings.TrimSpace(line)
		if branch == "" {
			continue
		}
		if branch == "origin/HEAD" || strings.HasPrefix(branch, "origin/HEAD ") {
			continue
		}
		if strings.HasPrefix(branch, "origin/") {
			localName := strings.TrimPrefix(branch, "origin/")
			if !localSet[localName] {
				remoteOnly = append(remoteOnly, branch)
			}
		} else {
			localSet[branch] = true
		}
	}

	var result []string
	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		branch := strings.TrimSpace(line)
		if localSet[branch] {
			result = append(result, branch)
			delete(localSet, branch)
		}
	}
	for _, remote := range remoteOnly {
		localName := strings.TrimPrefix(remote, "origin/")
		found := false
		for _, r := range result {
			if r == localName {
				found = true
				break
			}
		}
		if !found {
			result = append(result, remote)
		}
	}

	return result
}

// ListBranches runs `git branch -a --format='%(refname:short)'` and returns
// a deduplicated branch list. Returns nil on error.
func ListBranches(projectPath string) []string {
	cmd := exec.Command("git", "-C", projectPath, "branch", "-a", "--format=%(refname:short)")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return ParseBranchList(string(out))
}
