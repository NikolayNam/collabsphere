package mapper

import (
	authdomain "github.com/NikolayNam/collabsphere/internal/auth/domain"
	"github.com/NikolayNam/collabsphere/internal/auth/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/persistence/gorm/model"
)

func ToDBRefreshSessionForCreate(s *authdomain.RefreshSession) *dbmodel.RefreshSession {
	if s == nil {
		return nil
	}

	return &dbmodel.RefreshSession{
		UUIDPK: model.UUIDPK{
			ID: s.ID(),
		},
		Timestamps: model.Timestamps{
			CreatedAt: s.CreatedAt(),
			UpdatedAt: s.UpdatedAt(),
		},
		AccountID: s.AccountID(),
		TokenHash: s.TokenHash(),
		UserAgent: s.UserAgent(),
		IP:        s.IP(),
		ExpiresAt: s.ExpiresAt(),
		RevokedAt: s.RevokedAt(),
	}
}
