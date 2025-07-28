package utils

import "regexp"

// ValidateEmail performs basic email validation
func ValidateEmail(email string) bool {
	// Simple regex for basic email validation
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}
