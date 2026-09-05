package adapters

import "testing"

func TestAllReturnsFiveAdaptersInStableOrder(t *testing.T) {
	names := make([]string, 0, 5)
	for _, a := range All() {
		names = append(names, a.Name())
	}
	want := []string{"claude-code", "codex", "copilot-vscode", "gemini-cli", "antigravity"}
	if len(names) != len(want) {
		t.Fatalf("All() returned %d adapters, want %d: %v", len(names), len(want), names)
	}
	for i, n := range want {
		if names[i] != n {
			t.Fatalf("adapter %d = %q, want %q (full order: %v)", i, names[i], n, names)
		}
	}
}

type panickingAdapter struct{}

func (panickingAdapter) Name() string { return "panicking" }
func (panickingAdapter) Configure(string, string) Result {
	panic("boom")
}

func TestSafeConfigureConvertsPanicToError(t *testing.T) {
	result := safeConfigure(panickingAdapter{}, t.TempDir(), "")
	if result.Status != StatusError {
		t.Fatalf("Status = %q, want %q", result.Status, StatusError)
	}
	if result.Agent != "panicking" {
		t.Fatalf("Agent = %q, want panicking", result.Agent)
	}
}

func TestRunAllIsolatesFailures(t *testing.T) {
	projectDir := newGitRepoWithTrackedFile(t, "AGENTS.md", "team instructions\n")

	results := RunAll(projectDir, "- A: B\n")
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	statusByAgent := make(map[string]Status, len(results))
	for _, r := range results {
		statusByAgent[r.Agent] = r.Status
	}

	if statusByAgent["codex"] != StatusSkipped {
		t.Fatalf("codex Status = %q, want %q (tracked AGENTS.md must be skipped, not blocking others)", statusByAgent["codex"], StatusSkipped)
	}
	for _, agent := range []string{"claude-code", "copilot-vscode", "gemini-cli", "antigravity"} {
		if statusByAgent[agent] != StatusConfigured {
			t.Fatalf("%s Status = %q, want %q — one adapter being skipped must not affect the others", agent, statusByAgent[agent], StatusConfigured)
		}
	}
}
