package humaerr

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// Install переопределяет глобальные фабрики ошибок Huma так,
// чтобы любые huma.ErrorXXX(...) и встроенные ошибки (валидация, парсинг)
// возвращались в твоём формате (Problem).
//
// Документация Huma прямо говорит, что huma.NewError можно заменять.
func Install() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		return buildProblem(status, msg, errs...)
	}

	huma.NewErrorWithContext = func(_ huma.Context, status int, msg string, errs ...error) huma.StatusError {
		return buildProblem(status, msg, errs...)
	}
}

func buildProblem(status int, msg string, errs ...error) *Problem {
	p := &Problem{
		Status: status,
		Title:  http.StatusText(status),
		Detail: msg,
		Code:   defaultCodeFromStatus(status),
	}

	for _, e := range errs {
		if e == nil {
			continue
		}

		// Если это huma.ErrorDetailer — вытащим location/message аккуратно.
		if d, ok := e.(huma.ErrorDetailer); ok && d.ErrorDetail() != nil {
			ed := d.ErrorDetail()
			p.Errors = append(p.Errors, Item{
				Field:   simplifyLocation(ed.Location),
				Message: strings.TrimSpace(ed.Message),
			})
			continue
		}

		// иначе просто строка
		p.Errors = append(p.Errors, Item{Message: e.Error()})
	}

	if strings.TrimSpace(p.Detail) == "" {
		p.Detail = p.Title
	}
	return p
}

// Превращает "body.email" / "path.id" / "query.limit" -> "email" / "id" / "limit"
func simplifyLocation(loc string) string {
	loc = strings.TrimSpace(loc)
	for _, prefix := range []string{"body.", "path.", "query.", "header.", "cookie."} {
		if strings.HasPrefix(loc, prefix) {
			return strings.TrimPrefix(loc, prefix)
		}
	}
	return loc
}
