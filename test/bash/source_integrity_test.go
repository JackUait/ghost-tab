package bash_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// This file guards against a whole class of installer/runtime breakage:
// a shell script that `source`s a lib file (or names a lib in wrapper's
// _gt_libs array) that no longer exists in the repo. Under `set -e`, sourcing
// a missing file aborts with "No such file or directory" — exactly the bug
// that broke `npx wisp-deck` when project-actions-tui.sh was deleted but its
// `source` line lingered in bin/wisp-deck.
//
// Any future deletion/rename/typo of a statically-sourced lib file is caught
// here at test time instead of by users at install time.

// form1 matches `source "$VAR/lib/<rel>.sh"` — a variable prefix followed by a
// literal `lib/....sh` tail. Captures the repo-relative `lib/....sh` path.
var form1SourceRe = regexp.MustCompile(`source "\$[A-Za-z_][A-Za-z0-9_]*/(lib/[^"$]+\.sh)"`)

// form2 matches `source "$VAR/<name>.sh"` — a variable prefix followed by a
// single bare filename (no slash). Used by lib files that source siblings via
// their own directory variable (e.g. config-tui.sh). Resolved relative to the
// scanned file's directory. Captures `<name>.sh`.
var form2SourceRe = regexp.MustCompile(`source "\$[A-Za-z_][A-Za-z0-9_]*/([^"/$]+\.sh)"`)

// gtLibsRe extracts the contents of wrapper.sh's `_gt_libs=(...)` array.
var gtLibsRe = regexp.MustCompile(`_gt_libs=\(([^)]*)\)`)

// collectShellScripts returns repo-relative paths of every shell script whose
// `source` statements we can statically resolve: the two bin entry points,
// wrapper.sh, and every lib/**/*.sh.
func collectShellScripts(t *testing.T, root string) []string {
	t.Helper()
	scripts := []string{"bin/wisp-deck", "bin/wisp-deck-config", "wrapper.sh"}

	libRoot := filepath.Join(root, "lib")
	err := filepath.Walk(libRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".sh") {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			scripts = append(scripts, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk lib/: %v", err)
	}
	return scripts
}

func TestSourceIntegrity_all_statically_sourced_lib_files_exist(t *testing.T) {
	root := projectRoot(t)

	for _, script := range collectShellScripts(t, root) {
		script := script
		t.Run(script, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, script))
			if err != nil {
				t.Fatalf("failed to read %s: %v", script, err)
			}
			scriptDir := filepath.Dir(script) // repo-relative dir of the sourcing file

			for i, line := range strings.Split(string(data), "\n") {
				lineNo := i + 1

				// Form 1: `$VAR/lib/<rel>.sh` → repo-relative target is the capture.
				for _, m := range form1SourceRe.FindAllStringSubmatch(line, -1) {
					assertSourcedFileExists(t, root, script, lineNo, m[1])
				}

				// Form 2: `$VAR/<name>.sh` → resolve relative to the scanned
				// file's own directory (sibling source).
				for _, m := range form2SourceRe.FindAllStringSubmatch(line, -1) {
					target := filepath.Join(scriptDir, m[1])
					assertSourcedFileExists(t, root, script, lineNo, target)
				}
			}
		})
	}
}

// TestSourceIntegrity_wrapper_gt_libs_all_exist checks that every entry in
// wrapper.sh's `_gt_libs` array resolves to a real lib/<name>.sh file. These
// are sourced dynamically (`source "$_WRAPPER_DIR/lib/${_gt_lib}.sh"`), so the
// generic source-line scan cannot see them.
func TestSourceIntegrity_wrapper_gt_libs_all_exist(t *testing.T) {
	root := projectRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "wrapper.sh"))
	if err != nil {
		t.Fatalf("failed to read wrapper.sh: %v", err)
	}

	m := gtLibsRe.FindStringSubmatch(string(data))
	if m == nil {
		t.Fatal("could not find _gt_libs=(...) array in wrapper.sh")
	}

	libs := strings.Fields(m[1])
	if len(libs) == 0 {
		t.Fatal("_gt_libs array is empty — expected the list of runtime libs")
	}
	for _, lib := range libs {
		lib := lib
		t.Run(lib, func(t *testing.T) {
			target := filepath.Join("lib", lib+".sh")
			assertSourcedFileExists(t, root, "wrapper.sh", 0, target)
		})
	}
}

// assertSourcedFileExists fails the test if repoRelTarget does not exist,
// reporting where the source reference lives.
func assertSourcedFileExists(t *testing.T, root, fromScript string, lineNo int, repoRelTarget string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, repoRelTarget)); os.IsNotExist(err) {
		if lineNo > 0 {
			t.Errorf("%s:%d sources %q which does not exist — sourcing a missing file aborts the script under set -e", fromScript, lineNo, repoRelTarget)
		} else {
			t.Errorf("%s references lib %q which does not exist", fromScript, repoRelTarget)
		}
	}
}
