package searchkit

// SortSpec field -> SQL fragment (safe)
type SortSpec map[string]string

// FilterSpec field -> FieldSpec
type FilterSpec map[string]FieldSpec

// FieldSpec описывает, как поле выглядит в SQL и как нормализовать значение.
type FieldSpec struct {
	SQLExpr string // Например: "u.email", "c.full_name"
	Type    string // "uuid"|"text"|"int"|"bool"|"time"|...
}
