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
	return &APIError{Status: http.StatusBadRequest, Code: "bad_request", Detail: detail}
}

func Unauthorized(detail string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: "unauthorized", Detail: detail}
}

func Forbidden(detail string) *APIError {
	return &APIError{Status: http.StatusForbidden, Code: "forbidden", Detail: detail}
}

func NotFound(detail string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: "not_found", Detail: detail}
}

func Conflict(detail string) *APIError {
	return &APIError{Status: http.StatusConflict, Code: "conflict", Detail: detail}
}

func Internal(detail string) *APIError {
	return &APIError{Status: http.StatusInternalServerError, Code: "internal", Detail: detail}
}
