package searchkit

import (
	"fmt"
	"strings"

	"github.com/NikolayNam/collabsphere-go/shared/strcase"
)

func NormalizeFilters(filters []Filter, allowed FilterSpec) ([]Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	// cache: raw field -> normalized snake
	fieldCache := make(map[string]string, minVal(len(filters), 32))
	// cache: normalized field -> allowed?
	allowedCache := make(map[string]bool, minVal(len(filters), 32))

	out := make([]Filter, 0, len(filters))
	for i, f := range filters {
		raw := strings.TrimSpace(f.Field)
		if raw == "" {
			return nil, fmt.Errorf("filters[%d].field is required", i)
		}

		// 1) normalize field (cached)
		field, ok := fieldCache[raw]
		if !ok {
			field = strcase.CamelToSnake(raw)
			fieldCache[raw] = field
		}
		if field == "" {
			return nil, fmt.Errorf("filters[%d].field is required", i)
		}

		// 2) strict op validation (no aliases)
		if !isValidOp(f.Op) {
			return nil, fmt.Errorf("filters[%d].op invalid: %s", i, f.Op)
		}

		// 3) whitelist check (cached)
		allowedOK, ok := allowedCache[field]
		if !ok {
			_, exists := allowed[field]
			allowedOK = exists
			allowedCache[field] = allowedOK
		}
		if !allowedOK {
			return nil, fmt.Errorf("filters[%d].field invalid: %s", i, f.Field)
		}

		// 4) basic structural validation (op/value compatibility)
		canon := Filter{Field: field, Op: f.Op, Value: f.Value}
		if err := ValidateBasic(canon); err != nil {
			return nil, fmt.Errorf("filters[%d]: %w", i, err)
		}

		out = append(out, canon)
	}

	return out, nil
}

func isValidOp(op Op) bool {
	switch op {
	case OpEQ, OpNE, OpGT, OpGTE, OpLT, OpLTE, OpLike, OpIn, OpBetween, OpIsNull, OpNotNull:
		return true
	default:
		return false
	}
}

func minVal(a, b int) int {
	if a < b {
		return a
	}
	return b
}
