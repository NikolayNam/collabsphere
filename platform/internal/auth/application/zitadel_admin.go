package application

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"

	autherrors "github.com/NikolayNam/collabsphere/internal/auth/application/errors"
	"github.com/NikolayNam/collabsphere/internal/auth/application/ports"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
)

type ForceVerifyZitadelUserEmailCmd struct {
	UserID string
}

type ForceVerifyZitadelUserEmailResult struct {
	UserID          string
	Email           string
	Verified        bool
	AlreadyVerified bool
}

type zitadelAdmin struct {
	client ports.ZitadelAdminClient
}

func newZitadelAdmin(client ports.ZitadelAdminClient) *zitadelAdmin {
	return &zitadelAdmin{client: client}
}

func (z *zitadelAdmin) ForceVerifyUserEmail(ctx context.Context, cmd ForceVerifyZitadelUserEmailCmd) (*ForceVerifyZitadelUserEmailResult, error) {
	if z == nil || z.client == nil {
		return nil, autherrors.Unavailable("ZITADEL admin email verification is unavailable")
	}
	userID := strings.TrimSpace(cmd.UserID)
	if userID == "" {
		return nil, autherrors.InvalidInput("ZITADEL user id is required")
	}

	res, err := z.client.ForceVerifyUserEmail(ctx, userID)
	if err != nil {
		return nil, mapZitadelAdminError(err)
	}
	return &ForceVerifyZitadelUserEmailResult{
		UserID:          res.UserID,
		Email:           res.Email,
		Verified:        true,
		AlreadyVerified: res.AlreadyVerified,
	}, nil
}

func mapZitadelAdminError(err error) error {
	var apiErr *ports.ZitadelAdminAPIError
	if stderrors.As(err, &apiErr) && apiErr != nil {
		switch apiErr.StatusCode {
		case http.StatusBadRequest, http.StatusPreconditionFailed, http.StatusUnprocessableEntity:
			return autherrors.InvalidInput(nonEmpty(apiErr.Message, "ZITADEL request is invalid"))
		case http.StatusNotFound:
			return fault.NotFound("ZITADEL user not found", fault.Code("AUTH_ZITADEL_USER_NOT_FOUND"))
		case http.StatusConflict:
			return fault.Conflict(nonEmpty(apiErr.Message, "ZITADEL request conflicted"), fault.Code("AUTH_ZITADEL_CONFLICT"))
		case http.StatusTooManyRequests:
			return fault.TooManyRequests("ZITADEL admin API rate limit exceeded", fault.Code("AUTH_ZITADEL_RATE_LIMIT"))
		case http.StatusUnauthorized, http.StatusForbidden:
			return autherrors.Unavailable("ZITADEL admin token is invalid or missing required permissions")
		default:
			if apiErr.StatusCode >= 500 {
				return autherrors.Unavailable("ZITADEL admin API is unavailable")
			}
		}
	}
	return autherrors.Internal("force verify zitadel email failed", err)
}

func nonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
