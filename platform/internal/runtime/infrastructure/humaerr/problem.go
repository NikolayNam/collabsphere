package humaerr

import "net/http"

// Problem — публичная ошибка API.
// Это то, что клиент увидит в JSON.
//
// Мы делаем её совместимой с huma.StatusError:
//   - Error() string  (стандартный интерфейс error)
//   - GetStatus() int (интерфейс huma.StatusError)
type Problem struct {
	Status int    `json:"status" example:"400" doc:"HTTP status code"`
	Title  string `json:"title" example:"Bad Request" doc:"Short, human-readable summary"`
	Detail string `json:"detail" example:"Invalid input" doc:"Human-readable details safe for clients"`
	Code   string `json:"code" example:"accounts.invalid_input" doc:"Machine-readable stable error code"`

	// Errors — опциональные детали, обычно для валидации.
	Errors []Item `json:"errors,omitempty"`
}

type Item struct {
	Field   string `json:"field,omitempty" example:"email" doc:"Field name, if applicable"`
	Message string `json:"message" example:"must be a valid email"`
}

func (p *Problem) Error() string {
	if p == nil {
		return "<nil>"
	}
	if p.Detail != "" {
		return p.Detail
	}
	if p.Title != "" {
		return p.Title
	}
	if p.Status != 0 {
		return http.StatusText(p.Status)
	}
	return "error"
}

// GetStatus делает Problem huma.StatusError.
// Huma использует этот статус как HTTP status code ответа.
func (p *Problem) GetStatus() int { return p.Status }
