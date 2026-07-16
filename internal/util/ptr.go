package util

// Deref returns the pointed-to string, or "" if the pointer is nil.
func Deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// PtrIfSet returns nil for an empty string so optional *string columns stay NULL.
func PtrIfSet(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
