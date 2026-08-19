package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureProjectStateExcludedAddsEntry(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}

	if err := EnsureProjectStateExcluded(repoRoot); err != nil {
		t.Fatalf("EnsureProjectStateExcluded returned error: %v", err)
	}

	excludePath := filepath.Join(repoRoot, ".git", "info", "exclude")
	data, err := os.ReadFile(excludePath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(data) != ".allmyagents/\n" {
		t.Fatalf("exclude content = %q, want %q", string(data), ".allmyagents/\n")
	}
}

func TestEnsureProjectStateExcludedFindsRepoRootFromSubdirectory(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}
	nested := filepath.Join(repoRoot, "src", "pkg")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	if err := EnsureProjectStateExcluded(nested); err != nil {
		t.Fatalf("EnsureProjectStateExcluded returned error: %v", err)
	}

	excludePath := filepath.Join(repoRoot, ".git", "info", "exclude")
	if !Exists(excludePath) {
		t.Fatalf("expected exclude file at repo root %s", excludePath)
	}
	if Exists(filepath.Join(nested, ".git")) {
		t.Fatal("must not create .git in the subdirectory")
	}
}

func TestEnsureProjectStateExcludedIsIdempotent(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.Mkdir(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}

	for i := range 3 {
		if err := EnsureProjectStateExcluded(repoRoot); err != nil {
			t.Fatalf("EnsureProjectStateExcluded run %d returned error: %v", i, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(repoRoot, ".git", "info", "exclude"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(data) != ".allmyagents/\n" {
		t.Fatalf("exclude content = %q, want exactly one entry", string(data))
	}
}

func TestEnsureProjectStateExcludedPreservesExistingContent(t *testing.T) {
	repoRoot := t.TempDir()
	gitInfoDir := filepath.Join(repoRoot, ".git", "info")
	if err := os.MkdirAll(gitInfoDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	existing := "# local excludes\n*.log\nbuild/\n"
	if err := os.WriteFile(filepath.Join(gitInfoDir, "exclude"), []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if err := EnsureProjectStateExcluded(repoRoot); err != nil {
		t.Fatalf("EnsureProjectStateExcluded returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(gitInfoDir, "exclude"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	want := existing + ".allmyagents/\n"
	if string(data) != want {
		t.Fatalf("exclude content = %q, want %q", string(data), want)
	}
}

func TestEnsureProjectStateExcludedNoOpWhenAlreadyPresent(t *testing.T) {
	repoRoot := t.TempDir()
	gitInfoDir := filepath.Join(repoRoot, ".git", "info")
	if err := os.MkdirAll(gitInfoDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	excludePath := filepath.Join(gitInfoDir, "exclude")
	existing := "*.log\n.allmyagents/\nbuild/\n"
	if err := os.WriteFile(excludePath, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	before, err := os.ReadFile(excludePath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if err := EnsureProjectStateExcluded(repoRoot); err != nil {
		t.Fatalf("EnsureProjectStateExcluded returned error: %v", err)
	}

	after, err := os.ReadFile(excludePath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("exclude content changed: got %q, want unchanged %q", string(after), string(before))
	}
}

func TestEnsureProjectStateExcludedOutsideGitRepoIsNoOp(t *testing.T) {
	dir := t.TempDir()

	if err := EnsureProjectStateExcluded(dir); err != nil {
		t.Fatalf("EnsureProjectStateExcluded returned error: %v", err)
	}
	if Exists(filepath.Join(dir, ".git")) {
		t.Fatal("must not create a .git directory outside a repository")
	}
}
