package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/profile"
)

func TestRunInitCreatesProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)

	input := strings.NewReader(onboardingKeys(0, 1, 2, 3, 0, 1, 2, 3))
	var out bytes.Buffer
	var errOut bytes.Buffer

	if err := Run([]string{"init"}, input, &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}

	if !strings.Contains(out.String(), "Developer Profile created") {
		t.Fatalf("expected success output, got %q", out.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errOut.String())
	}

	profilePath := filepath.Join(home, "developer-profile.json")
	if !fileExists(profilePath) {
		t.Fatalf("expected profile at %s", profilePath)
	}
}

func onboardingKeys(selections ...int) string {
	var builder strings.Builder
	for _, selection := range selections {
		for range selection {
			builder.WriteString("\x1b[B")
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}

func TestRunUnknownCommandPrintsUsage(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := Run([]string{"unknown"}, strings.NewReader(""), &out, &errOut)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut.String(), "allmyagents profile") {
		t.Fatalf("expected usage, got %q", errOut.String())
	}
}

func TestRunProfileReturnsMissingProfileError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)

	var out bytes.Buffer
	var errOut bytes.Buffer

	err := Run([]string{"profile"}, strings.NewReader("\n"), &out, &errOut)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut.String(), "read developer profile") {
		t.Fatalf("expected missing profile error, got %q", errOut.String())
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func testDeveloperProfile(t *testing.T) profile.DeveloperProfile {
	t.Helper()

	selections := make(profile.Preferences, len(profile.Questions))
	for _, question := range profile.Questions {
		option := question.Options[0]
		selections[question.ID] = profile.Selection{OptionID: option.ID, Value: option.Value, Label: option.Label, Description: option.Description}
	}
	return profile.NewDeveloperProfile(selections, time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC))
}

func TestRunOverrideSetChangesOverrideNotProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	profilePath := filepath.Join(home, "developer-profile.json")
	original := testDeveloperProfile(t)
	if err := profile.Save(profilePath, original); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	input := strings.NewReader(onboardingKeys(1, 2, 8))
	var out bytes.Buffer
	var errOut bytes.Buffer

	if err := Run([]string{"override"}, input, &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}
	if !strings.Contains(out.String(), "Session override updated at") {
		t.Fatalf("expected update confirmation, got %q", out.String())
	}

	updatedProfile, err := profile.Load(profilePath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if updatedProfile.UpdatedAt != original.UpdatedAt {
		t.Fatalf("developer profile was modified: updated_at got %s, want %s", updatedProfile.UpdatedAt, original.UpdatedAt)
	}
	for id, selection := range original.Preferences {
		if updatedProfile.Preferences[id] != selection {
			t.Fatalf("developer profile preference %s changed: got %#v, want %#v", id, updatedProfile.Preferences[id], selection)
		}
	}

	overridePath := profile.OverridePath(projectDir)
	override, err := profile.LoadOverride(overridePath)
	if err != nil {
		t.Fatalf("LoadOverride returned error: %v", err)
	}
	if len(override.Preferences) != 1 {
		t.Fatalf("expected exactly one override, got %#v", override.Preferences)
	}
	if override.Preferences["learning_method"].Value != "explain_execute" {
		t.Fatalf("unexpected override: %#v", override.Preferences["learning_method"])
	}
}

func TestRunOverrideShowReportsEffectivePreferences(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	profilePath := filepath.Join(home, "developer-profile.json")
	if err := profile.Save(profilePath, testDeveloperProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"override", "show"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}
	if !strings.Contains(out.String(), "No active overrides") {
		t.Fatalf("expected no active overrides, got %q", out.String())
	}
	if !strings.Contains(out.String(), "(profile)") {
		t.Fatalf("expected preferences sourced from profile, got %q", out.String())
	}
}

func TestRunOverrideClearRemovesOnePreferenceThenAll(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ALLMYAGENTS_HOME", home)
	projectDir := t.TempDir()
	t.Chdir(projectDir)

	profilePath := filepath.Join(home, "developer-profile.json")
	if err := profile.Save(profilePath, testDeveloperProfile(t)); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	overridePath := profile.OverridePath(projectDir)
	override := profile.SessionOverride{
		SchemaVersion: 1,
		Preferences: profile.Preferences{
			"learning_method": {OptionID: "C", Value: "explain_execute", Label: "Explain & Execute"},
			"work_priority":   {OptionID: "C", Value: "speed", Label: "Speed"},
		},
	}
	if err := profile.SaveOverride(overridePath, override); err != nil {
		t.Fatalf("SaveOverride returned error: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := Run([]string{"override", "clear", "learning_method"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}
	remaining, err := profile.LoadOverride(overridePath)
	if err != nil {
		t.Fatalf("LoadOverride returned error: %v", err)
	}
	if _, ok := remaining.Preferences["learning_method"]; ok {
		t.Fatalf("expected learning_method cleared, got %#v", remaining.Preferences)
	}
	if _, ok := remaining.Preferences["work_priority"]; !ok {
		t.Fatalf("expected work_priority to remain, got %#v", remaining.Preferences)
	}

	if err := Run([]string{"override", "clear"}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("Run returned error: %v\nstderr: %s", err, errOut.String())
	}
	if profile.Exists(overridePath) {
		t.Fatal("expected override file removed after clearing all preferences")
	}

	updatedProfile, err := profile.Load(profilePath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if updatedProfile.Preferences["learning_method"].Value != "socratic" {
		t.Fatalf("developer profile must remain untouched by override clearing, got %#v", updatedProfile.Preferences["learning_method"])
	}
}
