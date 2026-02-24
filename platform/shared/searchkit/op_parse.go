package searchkit

import (
	"fmt"
	"strings"
)

// ParseOpStrict — принимает только enum значения. Никаких "=".
func ParseOpStrict(op string) (Op, error) {
	s := strings.ToLower(strings.TrimSpace(op))
	switch Op(s) {
	case OpEQ, OpNE, OpGT, OpGTE, OpLT, OpLTE, OpLike, OpIn, OpBetween, OpIsNull, OpNotNull:
		return Op(s), nil
	default:
		return "", fmt.Errorf("invalid op: %s", op)
	}
}
