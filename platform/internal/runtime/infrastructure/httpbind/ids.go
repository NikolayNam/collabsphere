package httpbind

import (
	"context"
	"strings"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	groupsdomain "github.com/NikolayNam/collabsphere/internal/groups/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/google/uuid"
)

func Principal(ctx context.Context) authdomain.Principal {
	return authmw.PrincipalFromContext(ctx)
}

func RequireAccountUUID(ctx context.Context, unauthorized error) (uuid.UUID, error) {
	principal := Principal(ctx)
	if !principal.IsAccount() || principal.AccountID == uuid.Nil {
		return uuid.Nil, unauthorized
	}
	return principal.AccountID, nil
}

func RequireAccountID(ctx context.Context, unauthorized error) (accdomain.AccountID, error) {
	accountUUID, err := RequireAccountUUID(ctx, unauthorized)
	if err != nil {
		return accdomain.AccountID{}, err
	}
	accountID, err := accdomain.AccountIDFromUUID(accountUUID)
	if err != nil {
		return accdomain.AccountID{}, unauthorized
	}
	return accountID, nil
}

func ParseUUID(raw string, invalid error) (uuid.UUID, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == uuid.Nil {
		return uuid.Nil, invalid
	}
	return parsed, nil
}

func ParseAccountID(raw string, invalid error) (accdomain.AccountID, error) {
	parsed, err := ParseUUID(raw, invalid)
	if err != nil {
		return accdomain.AccountID{}, err
	}
	return accdomain.AccountIDFromUUID(parsed)
}

func ParseOrganizationID(raw string, invalid error) (orgdomain.OrganizationID, error) {
	parsed, err := ParseUUID(raw, invalid)
	if err != nil {
		return orgdomain.OrganizationID{}, err
	}
	return orgdomain.OrganizationIDFromUUID(parsed)
}

func ParseGroupID(raw string, invalid error) (groupsdomain.GroupID, error) {
	parsed, err := ParseUUID(raw, invalid)
	if err != nil {
		return groupsdomain.GroupID{}, err
	}
	return groupsdomain.GroupIDFromUUID(parsed)
}
