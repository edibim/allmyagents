package profile

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveOverlaysOverrideOnProfile(t *testing.T) {
	base := testProfile(t)
	override := SessionOverride{
		Preferences: Preferences{
			"learning_method": {OptionID: "D", Value: "execute", Label: "Execute"},
		},
	}

	effective := Resolve(base, override)

	if effective["learning_method"].Value != "execute" {
		t.Fatalf("learning_method = %#v, want execute", effective["learning_method"])
	}
	for id, selection := range base.Preferences {
		if id == "learning_method" {
			continue
		}
		if effective[id] != selection {
			t.Fatalf("preference %s = %#v, want %#v (untouched by override)", id, effective[id], selection)
		}
	}
	if base.Preferences["learning_method"].Value == "execute" {
		t.Fatal("base profile must not be mutated by Resolve")
	}
}

func TestLoadOverrideMissingFileReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".allmyagents", "session-override.json")

	override, err := LoadOverride(path)
	if err != nil {
		t.Fatalf("LoadOverride returned error: %v", err)
	}
	if len(override.Preferences) != 0 {
		t.Fatalf("expected empty preferences, got %#v", override.Preferences)
	}
}

func TestRunOverrideEditorSetsOneOverrideWithoutTouchingProfile(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "developer-profile.json")
	overridePath := filepath.Join(t.TempDir(), ".allmyagents", "session-override.json")
	original := testProfile(t)
	if err := Save(profilePath, original); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	input := strings.NewReader(editorKeys(1, 2, 8))
	var out bytes.Buffer

	if err := RunOverrideEditor(input, &out, profilePath, overridePath); err != nil {
		t.Fatalf("RunOverrideEditor returned error: %v", err)
	}
	if !strings.Contains(out.String(), "Session override updated at") {
		t.Fatalf("expected update confirmation, got %q", out.String())
	}

	override, err := LoadOverride(overridePath)
	if err != nil {
		t.Fatalf("LoadOverride returned error: %v", err)
	}
	if len(override.Preferences) != 1 {
		t.Fatalf("expected exactly one override, got %#v", override.Preferences)
	}
	if override.Preferences["learning_method"].Value != "explain_execute" {
		t.Fatalf("learning_method override = %#v, want explain_execute", override.Preferences["learning_method"])
	}

	updatedProfile, err := Load(profilePath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if updatedProfile.UpdatedAt != original.UpdatedAt {
		t.Fatalf("developer profile updated_at changed: got %s, want %s", updatedProfile.UpdatedAt, original.UpdatedAt)
	}
	for id, selection := range original.Preferences {
		if updatedProfile.Preferences[id] != selection {
			t.Fatalf("developer profile preference %s changed: got %#v, want %#v", id, updatedProfile.Preferences[id], selection)
		}
	}
}

func TestRunOverrideEditorCanBackOutWithoutChanges(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "developer-profile.json")
	overridePath := filepath.Join(t.TempDir(), ".allmyagents", "session-override.json")
	if err := Save(profilePath, testProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	input := strings.NewReader(editorKeys(1, 4, 8))
	var out bytes.Buffer

	if err := RunOverrideEditor(input, &out, profilePath, overridePath); err != nil {
		t.Fatalf("RunOverrideEditor returned error: %v", err)
	}
	if !strings.Contains(out.String(), "No session override changes made.") {
		t.Fatalf("expected no-change confirmation, got %q", out.String())
	}
	if Exists(overridePath) {
		t.Fatal("expected no session override file to be created")
	}
}

func TestClearOverrideRemovesOnePreferenceThenWholeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".allmyagents", "session-override.json")
	override := SessionOverride{
		SchemaVersion: 1,
		Preferences: Preferences{
			"learning_method": {OptionID: "C", Value: "explain_execute", Label: "Explain & Execute"},
			"work_priority":   {OptionID: "C", Value: "speed", Label: "Speed"},
		},
	}
	if err := SaveOverride(path, override); err != nil {
		t.Fatalf("SaveOverride returned error: %v", err)
	}

	if err := ClearOverride(path, "learning_method"); err != nil {
		t.Fatalf("ClearOverride returned error: %v", err)
	}
	remaining, err := LoadOverride(path)
	if err != nil {
		t.Fatalf("LoadOverride returned error: %v", err)
	}
	if _, ok := remaining.Preferences["learning_method"]; ok {
		t.Fatalf("expected learning_method cleared, got %#v", remaining.Preferences)
	}
	if _, ok := remaining.Preferences["work_priority"]; !ok {
		t.Fatalf("expected work_priority to remain, got %#v", remaining.Preferences)
	}
	if !Exists(path) {
		t.Fatal("expected override file to remain while a preference is still set")
	}

	if err := ClearOverride(path, ""); err != nil {
		t.Fatalf("ClearOverride returned error: %v", err)
	}
	if Exists(path) {
		t.Fatal("expected override file removed after clearing all preferences")
	}
}

func TestClearOverrideMissingFileIsNotAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".allmyagents", "session-override.json")

	if err := ClearOverride(path, ""); err != nil {
		t.Fatalf("ClearOverride returned error: %v", err)
	}
	if err := ClearOverride(path, "learning_method"); err != nil {
		t.Fatalf("ClearOverride returned error: %v", err)
	}
}

func TestShowOverrideReportsSourcePerPreference(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "developer-profile.json")
	overridePath := filepath.Join(t.TempDir(), ".allmyagents", "session-override.json")
	if err := Save(profilePath, testProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	override := SessionOverride{
		SchemaVersion: 1,
		Preferences: Preferences{
			"learning_method": {OptionID: "D", Value: "execute", Label: "Execute"},
		},
	}
	if err := SaveOverride(overridePath, override); err != nil {
		t.Fatalf("SaveOverride returned error: %v", err)
	}

	var out bytes.Buffer
	if err := ShowOverride(&out, profilePath, overridePath); err != nil {
		t.Fatalf("ShowOverride returned error: %v", err)
	}

	rendered := out.String()
	if !strings.Contains(rendered, "Learning method: Execute (override)") {
		t.Fatalf("expected overridden learning method, got %q", rendered)
	}
	if !strings.Contains(rendered, "(profile)") {
		t.Fatalf("expected other preferences sourced from profile, got %q", rendered)
	}
}
