package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	appDirName          = ".allmyagents"
	profileFileName     = "developer-profile.json"
	homeOverrideEnv     = "ALLMYAGENTS_HOME"
	sessionOverrideFile = "session-override.json"
)

func DefaultPath() (string, error) {
	if override := os.Getenv(homeOverrideEnv); override != "" {
		return filepath.Join(override, profileFileName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, appDirName, profileFileName), nil
}

// OverridePath returns the project-scoped location of the temporary session
// override file: <projectDir>/.allmyagents/session-override.json.
func OverridePath(projectDir string) string {
	return filepath.Join(projectDir, appDirName, sessionOverrideFile)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Save(path string, developerProfile DeveloperProfile) error {
	if Exists(path) {
		return fmt.Errorf("developer profile already exists at %s", path)
	}

	return write(path, developerProfile, "developer profile")
}

func Update(path string, developerProfile DeveloperProfile) error {
	if !Exists(path) {
		return fmt.Errorf("developer profile not found at %s; run allmyagents init first", path)
	}
	return write(path, developerProfile, "developer profile")
}

func Load(path string) (DeveloperProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DeveloperProfile{}, fmt.Errorf("read developer profile: %w", err)
	}

	var developerProfile DeveloperProfile
	if err := json.Unmarshal(data, &developerProfile); err != nil {
		return DeveloperProfile{}, fmt.Errorf("decode developer profile: %w", err)
	}
	return developerProfile, nil
}

func write(path string, value any, kind string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create %s directory: %w", kind, err)
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", kind, err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", kind, err)
	}
	return nil
}
