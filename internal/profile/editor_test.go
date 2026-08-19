package profile

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunEditorChangesOnePreferenceAndPreservesOthers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "developer-profile.json")
	original := testProfile(t)
	if err := Save(path, original); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	input := strings.NewReader(editorKeys(1, 2, 8))
	var out bytes.Buffer

	if err := RunEditor(input, &out, path); err != nil {
		t.Fatalf("RunEditor returned error: %v", err)
	}

	updated, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if updated.Preferences["learning_method"].Value != "explain_execute" {
		t.Fatalf("learning method = %#v, want explain_execute", updated.Preferences["learning_method"])
	}
	for id, originalSelection := range original.Preferences {
		if id == "learning_method" {
			continue
		}
		if updated.Preferences[id] != originalSelection {
			t.Fatalf("preference %s changed: got %#v, want %#v", id, updated.Preferences[id], originalSelection)
		}
	}
	if updated.CreatedAt != original.CreatedAt {
		t.Fatalf("created_at changed: got %s, want %s", updated.CreatedAt, original.CreatedAt)
	}
	if updated.UpdatedAt == original.UpdatedAt {
		t.Fatalf("updated_at was not changed")
	}
	if !strings.Contains(out.String(), "Developer Profile updated at") {
		t.Fatalf("expected update confirmation, got %q", out.String())
	}
}

func TestRunEditorCanBackOutWithoutChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "developer-profile.json")
	original := testProfile(t)
	if err := Save(path, original); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	input := strings.NewReader(editorKeys(1, 4, 8))
	var out bytes.Buffer

	if err := RunEditor(input, &out, path); err != nil {
		t.Fatalf("RunEditor returned error: %v", err)
	}

	updated, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if updated.UpdatedAt != original.UpdatedAt {
		t.Fatalf("updated_at changed: got %s, want %s", updated.UpdatedAt, original.UpdatedAt)
	}
	if updated.Preferences["learning_method"] != original.Preferences["learning_method"] {
		t.Fatalf("learning method changed: got %#v, want %#v", updated.Preferences["learning_method"], original.Preferences["learning_method"])
	}
	if !strings.Contains(out.String(), "No profile changes made.") {
		t.Fatalf("expected no-change confirmation, got %q", out.String())
	}
}

func TestRunEditorMissingProfileReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "developer-profile.json")
	var out bytes.Buffer

	err := RunEditor(strings.NewReader("\n"), &out, path)
	if err == nil {
		t.Fatal("expected missing profile error")
	}
	if !strings.Contains(err.Error(), "read developer profile") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPreferenceListUsesExpectedLabels(t *testing.T) {
	items := preferenceItems(testProfile(t))
	want := []string{
		"Experience",
		"Learning method",
		"Autonomy",
		"Explanation depth",
		"Work priority",
		"Code review",
		"Git",
		"Uncertainty policy",
		"Done",
	}

	if len(items) != len(want) {
		t.Fatalf("items = %d, want %d", len(items), len(want))
	}
	for index, label := range want {
		if items[index].Label != label {
			t.Fatalf("item %d = %q, want %q", index, items[index].Label, label)
		}
	}
}

func testProfile(t *testing.T) DeveloperProfile {
	t.Helper()

	selections := make(Preferences, len(Questions))
	for _, question := range Questions {
		selections[question.ID] = selectionFromOption(question.Options[0])
	}

	return NewDeveloperProfile(selections, time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC))
}

func editorKeys(moves ...int) string {
	var builder strings.Builder
	for _, move := range moves {
		for range move {
			builder.WriteString("\x1b[B")
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}
