package mapper

import "strings"

func stringToNilPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return new(s)
}
