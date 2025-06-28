package shared

// Sanitize takes a string and returns a sanitized version containing only ASCII characters.
// It converts non-ASCII characters to underscores and keeps only alphanumeric characters
// and select punctuation marks (., !, -, ', ", _) matching the API regex: ^[a-zA-Z0-9\.!\-'"_]+$
func Sanitize(s string) string {
	// Convert to ASCII
	ascii := make([]byte, len(s))
	for i, r := range s {
		if r < 128 {
			ascii[i] = byte(r)
		} else {
			ascii[i] = '_'
		}
	}

	// Filter allowed characters - matching API regex ^[a-zA-Z0-9\.!\-'"_]+$
	allowed := func(r byte) bool {
		return (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '.' || r == '!' || r == '-' ||
			r == '\'' || r == '"' || r == '_'
	}

	// Remove special characters like newline, carriage return, spaces, commas
	result := make([]byte, 0, len(ascii))
	for _, c := range ascii {
		if allowed(c) {
			result = append(result, c)
		}
	}

	return string(result)
}

// SanitizeWithSpaces takes a string and returns a sanitized version containing only ASCII characters.
// Similar to Sanitize but preserves spaces. Used for OS version strings which need to be human-readable.
func SanitizeWithSpaces(s string) string {
	// Convert to ASCII
	ascii := make([]byte, len(s))
	for i, r := range s {
		if r < 128 {
			ascii[i] = byte(r)
		} else {
			ascii[i] = '_'
		}
	}

	// Filter allowed characters - including spaces
	allowed := func(r byte) bool {
		return (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '.' || r == '!' || r == '-' ||
			r == '\'' || r == '"' || r == '_' ||
			r == ' '
	}

	// Remove special characters like newline, carriage return but keep spaces
	result := make([]byte, 0, len(ascii))
	for _, c := range ascii {
		if allowed(c) {
			result = append(result, c)
		}
	}

	return string(result)
}
