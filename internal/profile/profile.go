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
