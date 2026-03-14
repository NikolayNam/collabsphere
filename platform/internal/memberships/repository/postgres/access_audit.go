package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	memberdomain "github.com/NikolayNam/collabsphere/internal/memberships/domain"
	"github.com/NikolayNam/collabsphere/internal/memberships/repository/postgres/dbmodel"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/tx"
)

type AccessAuditRepo struct {
	db *gorm.DB
}

func NewAccessAuditRepo(db *gorm.DB) *AccessAuditRepo {
	return &AccessAuditRepo{db: db}
}

func (r *AccessAuditRepo) Append(ctx context.Context, event memberdomain.AccessAuditEvent) error {
	previousState, err := marshalJSONMap(event.PreviousState)
	if err != nil {
		return err
	}
	nextState, err := marshalJSONMap(event.NextState)
	if err != nil {
		return err
	}
	model := &dbmodel.OrganizationAccessAuditEvent{
		ID:               nonZeroUUID(event.ID),
		OrganizationID:   event.OrganizationID,
		ActorSubjectType: event.ActorSubjectType,
		ActorSubjectID:   cloneUUID(event.ActorSubjectID),
		ActorAccountID:   cloneUUID(event.ActorAccountID),
		Action:           event.Action,
		TargetType:       event.TargetType,
		TargetID:         cloneUUID(event.TargetID),
		RequestID:        cloneString(event.RequestID),
		PreviousState:    previousState,
		NextState:        nextState,
		CreatedAt:        event.CreatedAt,
	}
	return r.dbFrom(ctx).WithContext(ctx).Create(model).Error
}

func (r *AccessAuditRepo) dbFrom(ctx context.Context) *gorm.DB {
	if gormTx := tx.TxFromContext(ctx); gormTx != nil {
		return gormTx
	}
	return r.db
}

func marshalJSONMap(value map[string]any) ([]byte, error) {
	if len(value) == 0 {
		return []byte(`{}`), nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal access audit state: %w", err)
	}
	return b, nil
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}

func nonZeroUUID(value uuid.UUID) uuid.UUID {
	if value == uuid.Nil {
		return uuid.New()
	}
	return value
}
