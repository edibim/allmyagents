package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	appDirName      = ".allmyagents"
	profileFileName = "developer-profile.json"
	homeOverrideEnv = "ALLMYAGENTS_HOME"
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

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Save(path string, developerProfile DeveloperProfile) error {
	if Exists(path) {
		return fmt.Errorf("developer profile already exists at %s", path)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create developer profile directory: %w", err)
	}

	data, err := json.MarshalIndent(developerProfile, "", "  ")
	if err != nil {
		return fmt.Errorf("encode developer profile: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write developer profile: %w", err)
	}
	return nil
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
