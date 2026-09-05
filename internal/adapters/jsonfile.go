package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// readJSONObject reads a JSON object file as a generic map so that any keys
// AllMyAgents doesn't know about are preserved untouched when the file is
// written back. A missing file yields an empty object.
func readJSONObject(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if settings == nil {
		settings = map[string]any{}
	}
	return settings, nil
}

func writeJSONObject(path string, settings map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create %s directory: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
