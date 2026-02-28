package mapper

import (
	"strings"
	"time"
)

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stringToNilPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return new(s)
}

func nonZeroOrNow(t time.Time, now time.Time) time.Time {
	if t.IsZero() {
		return now
	}
	return t
}
