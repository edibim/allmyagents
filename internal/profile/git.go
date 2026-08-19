package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const gitExcludeEntry = appDirName + "/"

// EnsureProjectStateExcluded makes sure the project-local .allmyagents/
// directory is excluded from Git without touching any tracked file. It
// walks up from projectDir looking for a .git directory and, if one is
// found, appends an entry to <repo-root>/.git/info/exclude — Git's
// local-only, never-committed ignore mechanism — unless that entry is
// already present.
//
// If projectDir is not inside a Git repository, this is a safe no-op:
// AllMyAgents only manages Git-local exclusions where a Git repository
// actually exists, and never creates one.
//
// A .git that is a file rather than a directory (Git worktrees and
// submodules use this) is treated the same as "no repository found" for
// V0; supporting that layout is left for a future version.
func EnsureProjectStateExcluded(projectDir string) error {
	gitDir, err := findGitDir(projectDir)
	if err != nil {
		return err
	}
	if gitDir == "" {
		return nil
	}

	excludePath := filepath.Join(gitDir, "info", "exclude")
	return ensureExcludeEntry(excludePath, gitExcludeEntry)
}

func findGitDir(startDir string) (string, error) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, ".git")
		info, err := os.Stat(candidate)
		switch {
		case err == nil:
			if info.IsDir() {
				return candidate, nil
			}
			return "", nil
		case !os.IsNotExist(err):
			return "", fmt.Errorf("check for git directory: %w", err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func ensureExcludeEntry(excludePath string, entry string) error {
	data, err := os.ReadFile(excludePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read git exclude file: %w", err)
	}

	content := string(data)
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(excludePath), 0o755); err != nil {
		return fmt.Errorf("create git info directory: %w", err)
	}

	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += entry + "\n"

	if err := os.WriteFile(excludePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write git exclude file: %w", err)
	}
	return nil
}
