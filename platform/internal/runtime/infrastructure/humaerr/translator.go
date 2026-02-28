package humaerr

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
)

func From(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	var head http.Header
	if he, ok := errors.AsType[huma.HeadersError](err); ok {
		head = he.GetHeaders()
	}

	if se, ok := errors.AsType[huma.StatusError](err); ok {
		if head != nil {
			return huma.ErrorWithHeaders(se.(error), head)
		}
		return se.(error)
	}

	if ae, ok := fault.As(err); ok {
		out := fromAppError(ae)
		if head != nil {
			return huma.ErrorWithHeaders(out, head)
		}
		return out
	}

	if errors.Is(err, context.DeadlineExceeded) {
		out := &Problem{
			Status: http.StatusGatewayTimeout,
			Title:  http.StatusText(http.StatusGatewayTimeout),
			Detail: "Request timeout",
			Code:   "timeout",
		}
		if head != nil {
			return huma.ErrorWithHeaders(out, head)
		}
		return out
	}
	if errors.Is(err, context.Canceled) {
		out := &Problem{
			Status: http.StatusRequestTimeout,
			Title:  http.StatusText(http.StatusRequestTimeout),
			Detail: "Request canceled",
			Code:   "canceled",
		}
		if head != nil {
			return huma.ErrorWithHeaders(out, head)
		}
		return out
	}

	slog.Default().ErrorContext(ctx, "request failed",
		"event", "http.request.error",
		"request_id", chimw.GetReqID(ctx),
		"error", err.Error(),
	)

	out := &Problem{
		Status: http.StatusInternalServerError,
		Title:  http.StatusText(http.StatusInternalServerError),
		Detail: "Internal error",
		Code:   "internal",
	}
	if head != nil {
		return huma.ErrorWithHeaders(out, head)
	}
	return out
}

func fromAppError(fa *fault.Error) *Problem {
	status := statusFromKind(fa.Kind)

	detail := fa.Message
	if fa.Kind == fault.KindInternal && strings.TrimSpace(detail) == "" {
		detail = "Internal error"
	}
	if fa.Kind == fault.KindInternal {
		detail = "Internal error"
	}

	code := fa.Code
	if strings.TrimSpace(code) == "" {
		code = defaultCodeFromStatus(status)
	}

	p := &Problem{
		Status: status,
		Title:  http.StatusText(status),
		Detail: detail,
		Code:   code,
	}
	for _, fe := range fa.Fields {
		p.Errors = append(p.Errors, Item{
			Field:   fe.Field,
			Message: fe.Message,
		})
	}
	if strings.TrimSpace(p.Detail) == "" {
		p.Detail = p.Title
	}
	return p
}

func statusFromKind(k fault.Kind) int {
	switch k {
	case fault.KindValidation:
		return http.StatusBadRequest
	case fault.KindConflict:
		return http.StatusConflict
	case fault.KindNotFound:
		return http.StatusNotFound
	case fault.KindUnauthorized:
		return http.StatusUnauthorized
	case fault.KindForbidden:
		return http.StatusForbidden
	case fault.KindTooManyRequests:
		return http.StatusTooManyRequests
	case fault.KindUnavailable:
		return http.StatusServiceUnavailable
	case fault.KindInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func defaultCodeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case http.StatusServiceUnavailable:
		return "UNAVAILABLE"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	default:
		return "INTERNAL"
	}
}
