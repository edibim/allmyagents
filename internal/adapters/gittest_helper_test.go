package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// newGitRepo creates a real, empty git repository in a fresh temp dir.
func newGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	return dir
}

// newGitRepoWithTrackedFile creates a real git repository with relPath
// already committed, so profile.IsTracked(dir, relPath) reports true.
func newGitRepoWithTrackedFile(t *testing.T, relPath, content string) string {
	t.Helper()
	dir := newGitRepo(t)
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	runGit(t, dir, "add", relPath)
	runGit(t, dir, "commit", "-q", "-m", "add "+relPath)
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}
