package subscriptions

import "time"

// Event represents one calendar event used to render ICS files.
type Event struct {
	UID         string    `json:"uid,omitempty"`
	Summary     string    `json:"summary"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	StartAt     time.Time `json:"startAt"`
	EndAt       time.Time `json:"endAt"`
	AllDay      bool      `json:"allDay,omitempty"`
}

// Calendar is the shared source model used to generate both JSON and ICS outputs.
type Calendar struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	GroupName   string    `json:"groupName,omitempty"`
	Description string    `json:"description,omitempty"`
	Provider    string    `json:"provider"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Events      []Event   `json:"events,omitempty"`
}

// Provider is the plugin contract for producing calendar data.
type Provider interface {
	Name() string
	Generate() ([]Calendar, error)
}
