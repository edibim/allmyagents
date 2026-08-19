package profile

import "time"

type DeveloperProfile struct {
	SchemaVersion int         `json:"schema_version"`
	ProfileType   string      `json:"profile_type"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
	Preferences   Preferences `json:"preferences"`
}

type Preferences map[string]Selection

type Selection struct {
	OptionID    string `json:"option_id"`
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

func NewDeveloperProfile(preferences Preferences, now time.Time) DeveloperProfile {
	timestamp := now.Format(time.RFC3339)
	return DeveloperProfile{
		SchemaVersion: 1,
		ProfileType:   "developer_profile",
		CreatedAt:     timestamp,
		UpdatedAt:     timestamp,
		Preferences:   preferences,
	}
}

// PreferenceLabel returns the human-readable label for a question ID (for
// example "learning_method" -> "Learning method"), the same labels used
// throughout the CLI. Callers outside this package that need to render
// preferences, such as the context renderer, use this instead of
// duplicating the label mapping.
func PreferenceLabel(questionID string) string {
	return preferenceMenuLabel(questionID)
}
