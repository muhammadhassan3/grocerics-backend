package util

import "time"

// ParseDate accepts an RFC3339 timestamp or a plain YYYY-MM-DD date; returns nil
// for an empty string. The bool is false only when a non-empty value fails both.
func ParseDate(s string) (*time.Time, bool) {
	if s == "" {
		return nil, true
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t, true
		}
	}
	return nil, false
}

// FmtDate renders a nullable timestamp as RFC3339, or "" when nil.
func FmtDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
