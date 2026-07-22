package dto

// converts bool to string
func StatusLabel(active bool) string {
	if active {
		return "active"
	}
	return "disabled"
}
