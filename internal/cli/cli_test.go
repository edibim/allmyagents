package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
