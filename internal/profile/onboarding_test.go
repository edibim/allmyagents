package profile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuestionsMatchV0OnboardingShape(t *testing.T) {
	if len(Questions) != 8 {
		t.Fatalf("expected 8 questions, got %d", len(Questions))
	}

	for _, question := range Questions {
		if question.ID == "" {
			t.Fatal("question ID must not be empty")
		}
		if question.Prompt == "" {
			t.Fatalf("question %s prompt must not be empty", question.ID)
		}
		if len(question.Options) != 4 {
			t.Fatalf("question %s expected 4 options, got %d", question.ID, len(question.Options))
		}

		for index, option := range question.Options {
			wantID := string(rune('A' + index))
			if option.ID != wantID {
				t.Fatalf("question %s option %d ID = %q, want %q", question.ID, index, option.ID, wantID)
			}
			if option.Value == "" || option.Label == "" {
				t.Fatalf("question %s option %s must have value and label", question.ID, option.ID)
			}
		}
	}
}

func TestAskQuestionsReturnsStructuredSelections(t *testing.T) {
	input := strings.NewReader(onboardingKeys(0, 1, 2, 3, 0, 1, 2, 3))
	var out bytes.Buffer

	selections, err := AskQuestions(input, &out, Questions)
	if err != nil {
		t.Fatalf("AskQuestions returned error: %v", err)
	}

	if selections["experience_level"].Value != "learning_guided" {
		t.Fatalf("unexpected experience level: %#v", selections["experience_level"])
	}
	if selections["learning_method"].Value != "guided" {
		t.Fatalf("unexpected learning method: %#v", selections["learning_method"])
	}
	if selections["uncertainty_policy"].Value != "keep_moving" {
		t.Fatalf("unexpected uncertainty policy: %#v", selections["uncertainty_policy"])
	}
	if !strings.Contains(out.String(), "Question 1 of 8") {
		t.Fatalf("expected progress text, got %q", out.String())
	}
	if !strings.Contains(out.String(), "Use Up/Down arrows to move. Press Enter to select.") {
		t.Fatalf("expected navigation instructions, got %q", out.String())
	}
	if strings.Contains(out.String(), "\033[") {
		t.Fatalf("non-interactive rendering should not include ANSI escapes, got %q", out.String())
	}
}

func TestAskQuestionsIgnoresUnknownKeys(t *testing.T) {
	input := strings.NewReader("x\x1b[B\n" + onboardingKeys(0, 0, 0, 0, 0, 0, 0))
	var out bytes.Buffer

	selections, err := AskQuestions(input, &out, Questions)
	if err != nil {
		t.Fatalf("AskQuestions returned error: %v", err)
	}
	if selections["experience_level"].Value != "builder_needs_help" {
		t.Fatalf("unexpected experience level after unknown key: %#v", selections["experience_level"])
	}
}

func TestRunOnboardingStoresDeveloperProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "developer-profile.json")
	input := strings.NewReader(onboardingKeys(0, 1, 2, 3, 0, 1, 2, 3))
	var out bytes.Buffer

	if err := RunOnboarding(input, &out, path); err != nil {
		t.Fatalf("RunOnboarding returned error: %v", err)
	}

	saved, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if saved.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", saved.SchemaVersion)
	}
	if len(saved.Preferences) != 8 {
		t.Fatalf("preferences = %d, want 8", len(saved.Preferences))
	}
	if saved.Preferences["git_autonomy"].Value != "autonomous_with_safeguards" {
		t.Fatalf("unexpected git autonomy: %#v", saved.Preferences["git_autonomy"])
	}

	content, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if !strings.Contains(string(content), "developer_profile") {
		t.Fatalf("expected profile type in JSON, got %s", string(content))
	}
	if !strings.Contains(out.String(), "Developer Profile created") {
		t.Fatalf("expected completion summary, got %q", out.String())
	}
	if !strings.Contains(out.String(), "- Learning: Guided") {
		t.Fatalf("expected learning summary, got %q", out.String())
	}
}

func TestRunOnboardingDoesNotOverwriteExistingProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "developer-profile.json")
	original := []byte(`{"schema_version":1,"profile_type":"developer_profile","preferences":{"sentinel":{"option_id":"A","value":"original","label":"Original"}}}`)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var out bytes.Buffer
	if err := RunOnboarding(strings.NewReader("A\n"), &out, path); err != nil {
		t.Fatalf("RunOnboarding returned error: %v", err)
	}

	saved, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if saved.Preferences["sentinel"].Value != "original" {
		t.Fatalf("existing profile was overwritten: %#v", saved.Preferences)
	}
}

func TestInteractiveRenderUsesColorAndStableScreenReset(t *testing.T) {
	var out bytes.Buffer
	renderQuestion(&out, Questions[1], 1, len(Questions), 1, newStyle(true))

	rendered := out.String()
	if !strings.Contains(rendered, "\033[H\033[J") {
		t.Fatalf("expected stable screen reset, got %q", rendered)
	}
	if !strings.Contains(rendered, "\033[34mHow should AI help you learn?\033[0m") {
		t.Fatalf("expected blue question text, got %q", rendered)
	}
	if !strings.Contains(rendered, "\033[33m> Guided\033[0m") {
		t.Fatalf("expected yellow selected answer, got %q", rendered)
	}
	if !strings.Contains(rendered, "\033[2m    Explain the reasoning, then let me try or implement it.\033[0m") {
		t.Fatalf("expected dim answer description, got %q", rendered)
	}
	if !strings.Contains(rendered, "\033[2mUse Up/Down arrows to move. Press Enter to select.\033[0m") {
		t.Fatalf("expected dim navigation instructions, got %q", rendered)
	}
}

func TestNonInteractiveRenderDoesNotResetScreen(t *testing.T) {
	var out bytes.Buffer
	renderQuestion(&out, Questions[0], 0, len(Questions), 0, newStyle(false))

	rendered := out.String()
	if strings.Contains(rendered, "\033[") {
		t.Fatalf("non-interactive render should not include terminal control sequences, got %q", rendered)
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
