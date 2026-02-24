package searchkit

import (
	"fmt"
	"strings"
)

// ValidateBasic — проверяет форму фильтра и совместимость Op/Value.
func ValidateBasic(f Filter) error {
	if strings.TrimSpace(f.Field) == "" {
		return fmt.Errorf("field is required")
	}

	switch f.Op {
	case OpEQ, OpNE, OpGT, OpGTE, OpLT, OpLTE, OpLike:
		if f.Value == nil {
			return fmt.Errorf("value is required for op=%s", f.Op)
		}
	case OpIn:
		if f.Value == nil {
			return fmt.Errorf("value is required for op=in")
		}
	case OpBetween:
		if f.Value == nil {
			return fmt.Errorf("value is required for op=between")
		}
	case OpIsNull, OpNotNull:
		// value ignored
	default:
		return fmt.Errorf("unsupported op: %s", f.Op)
	}

	return nil
}
