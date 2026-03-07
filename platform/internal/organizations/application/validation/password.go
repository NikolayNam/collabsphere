package validation

import (
	"strings"
	"unicode/utf8"

	apperrors "github.com/NikolayNam/collabsphere/internal/accounts/application/errors"
)

const (
	MinPasswordRunes = 8
	MaxBcryptBytes   = 72 // bcrypt uses first 72 bytes; enforce to avoid truncation surprises
)

// ValidatePassword validates raw password as a business/application rule.
// It MUST NOT mutate the input.
func ValidatePassword(pw string) error {
	trimmed := strings.TrimSpace(pw)
	if trimmed == "" {
		return apperrors.InvalidInput("Password is required")
	}

	// Do not silently "fix" user input. If they send spaces at edges, fail fast.
	if pw != trimmed {
		return apperrors.InvalidInput("Password must not start or end with spaces")
	}

	if utf8.RuneCountInString(pw) < MinPasswordRunes {
		return apperrors.InvalidInput("Password must be at least 8 characters")
	}

	// bcrypt truncation risk: even if rune-count is fine, bytes can exceed 72.
	if len([]byte(pw)) > MaxBcryptBytes {
		return apperrors.InvalidInput("Password is too long")
	}

	return nil
}
