package apierr

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Status int    `json:"-"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Detail)
}

func BadRequest(detail string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Detail: detail}
}

func Unauthorized(detail string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: "UNAUTHORIZED", Detail: detail}
}

func Forbidden(detail string) *APIError {
	return &APIError{Status: http.StatusForbidden, Code: "FORBIDDEN", Detail: detail}
}

func NotFound(detail string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: "NOT_FOUND", Detail: detail}
}

func Conflict(detail string) *APIError {
	return &APIError{Status: http.StatusConflict, Code: "CONFLICT", Detail: detail}
}

func Internal(detail string) *APIError {
	return &APIError{Status: http.StatusInternalServerError, Code: "INTERNAL", Detail: detail}
}
