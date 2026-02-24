package strcase

import (
	"fmt"
	"regexp"
	"strings"
)

var camelRx = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func CamelToSnake(s string) string {
	if s == "" {
		return s
	}
	var out = camelRx.ReplaceAllString(s, `${1}_${2}`)
	out = strings.ReplaceAll(out, "-", "_")
	return strings.ToLower(out)
}

func CamelToSnakeStrict(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("value is required")
	}
	return CamelToSnake(s), nil
}
