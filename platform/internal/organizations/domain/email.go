package domain

import (
	"strings"
	"unicode/utf8"
)

const (
	maxEmailLength      = 254
	maxEmailLocalLength = 64
	maxDomainLabelLen   = 63
)

type Email string

func NewEmail(raw string) (Email, error) {
	s := normalizeEmail(raw)
	if err := validateEmail(s); err != nil {
		return "", err
	}
	return Email(s), nil
}

func (e Email) String() string {
	return string(e)
}

func (e Email) IsZero() bool {
	return strings.TrimSpace(string(e)) == ""
}

func normalizeEmail(raw string) string {
	s := strings.TrimSpace(raw)

	// Если твой бизнес-контракт реально требует lowercase:
	s = strings.ToLower(s)

	return s
}

func validateEmail(s string) error {
	if s == "" {
		return ErrEmailEmpty
	}

	if utf8.RuneCountInString(s) > maxEmailLength {
		return ErrEmailTooLong
	}

	if strings.ContainsAny(s, " \t\r\n") {
		return ErrEmailInvalid
	}

	if strings.Count(s, "@") != 1 {
		return ErrEmailInvalid
	}

	local, domain, ok := strings.Cut(s, "@")
	if !ok || local == "" || domain == "" {
		return ErrEmailInvalid
	}

	if utf8.RuneCountInString(local) > maxEmailLocalLength {
		return ErrEmailInvalid
	}

	if !validateLocalPart(local) {
		return ErrEmailInvalid
	}

	if !validateDomain(domain) {
		return ErrEmailInvalid
	}

	return nil
}

func validateLocalPart(s string) bool {
	if strings.HasPrefix(s, ".") || strings.HasSuffix(s, ".") {
		return false
	}
	if strings.Contains(s, "..") {
		return false
	}

	for _, r := range s {
		if isASCIIAlphaNum(r) {
			continue
		}

		switch r {
		case '.', '_', '-', '+':
			continue
		default:
			return false
		}
	}

	return true
}

func validateDomain(s string) bool {
	if !strings.Contains(s, ".") {
		return false
	}
	if strings.HasPrefix(s, ".") || strings.HasSuffix(s, ".") {
		return false
	}
	if strings.Contains(s, "..") {
		return false
	}

	labels := strings.Split(s, ".")
	for _, label := range labels {
		if !validateDomainLabel(label) {
			return false
		}
	}

	return true
}

func validateDomainLabel(s string) bool {
	if s == "" {
		return false
	}
	if utf8.RuneCountInString(s) > maxDomainLabelLen {
		return false
	}
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return false
	}

	for _, r := range s {
		if isASCIIAlphaNum(r) || r == '-' {
			continue
		}
		return false
	}

	return true
}

func isASCIIAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}
