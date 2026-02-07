package common

import "time"

// DocProperties holds document metadata (core properties).
type DocProperties struct {
	Title          string
	Subject        string
	Creator        string
	Keywords       string
	Description    string
	Category       string
	LastModifiedBy string
	Company        string
	Manager        string
	Created        time.Time
	Modified       time.Time
	Revision       int
}

// NewDocProperties returns DocProperties with sensible defaults.
func NewDocProperties() DocProperties {
	now := time.Now()
	return DocProperties{
		Creator:        "GoWord",
		LastModifiedBy: "GoWord",
		Created:        now,
		Modified:       now,
		Revision:       1,
	}
}
