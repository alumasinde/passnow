package database

import "strings"

// IsDuplicateKeyErr reports whether err is a MySQL duplicate-entry error
// (1062), for repositories that want to translate it into a domain-specific
// "already exists" error. Centralized here so every repository checks it
// the same way instead of each re-implementing the string match.
func IsDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Duplicate entry")
}
