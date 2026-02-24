package searchkit

// Payload — внутренний (domain-level) контракт для поиска.
// Он НЕ зависит от ogen/sqlx/sqlc и используется между transport -> bootstrap -> repo.
type Payload struct {
	Page      int      `json:"page"`
	Size      int      `json:"size"`
	OrderBy   []string `json:"orderBy"`
	OrderDesc bool     `json:"orderDesc"`
	Filters   []Filter `json:"filters"`
}

// Filter — канонический фильтр. ВАЖНО: Op типизированный (enum), не string.
type Filter struct {
	Field string `json:"field"` // приходит как camelCase; нормализуется (camel_to_snake) перед сборкой SQL
	Op    Op     `json:"op"`    // только eq/ne/gt/... (строго)
	Value any    `json:"value"` // для is_null/not_null может быть nil/ignored
}

// Op — перечисление допустимых операций фильтрации.
// Снаружи в API ты разрешаешь только эти значения (enum в OpenAPI).
type Op string

const (
	OpEQ      Op = "eq"
	OpNE      Op = "ne"
	OpGT      Op = "gt"
	OpGTE     Op = "gte"
	OpLT      Op = "lt"
	OpLTE     Op = "lte"
	OpLike    Op = "like"
	OpIn      Op = "in"
	OpBetween Op = "between"
	OpIsNull  Op = "is_null"
	OpNotNull Op = "not_null"
)
