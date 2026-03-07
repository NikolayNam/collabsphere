package mapper

import (
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/dbmodel"
)

func ToDomainRefreshSession(m *dbmodel.RefreshSession) (*authdomain.RefreshSession, error) {
	if m == nil {
		return nil, nil
	}

	return authdomain.RehydrateRefreshSession(authdomain.RehydrateRefreshSessionParams{
		ID:        m.ID,
		AccountID: m.AccountID,
		TokenHash: m.TokenHash,
		UserAgent: m.UserAgent,
		IP:        m.IP,
		ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	})
}
