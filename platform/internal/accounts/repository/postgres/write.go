package postgres

import (
    "context"
    "errors"
    "fmt"

    apperrors "github.com/NikolayNam/collabsphere/internal/accounts/application/errors"
    "github.com/NikolayNam/collabsphere/internal/accounts/domain"
    "github.com/NikolayNam/collabsphere/internal/accounts/repository/postgres/mapper"
    "gorm.io/gorm"
)

func (r *AccountRepo) Create(ctx context.Context, account *domain.Account) error {
    if account == nil {
        return errors.New("account is nil")
    }

    accountModel := mapper.ToDBAccountForCreate(account)
    credentialModel := mapper.ToDBPasswordCredentialForCreate(account)
    if accountModel == nil || credentialModel == nil {
        return errors.New("db account model is nil")
    }

    db := r.dbFrom(ctx).WithContext(ctx)

    return db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(accountModel).Error; err != nil {
            if isUniqueViolation(err) {
                return fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
            }
            return err
        }

        if err := tx.Create(credentialModel).Error; err != nil {
            if isUniqueViolation(err) {
                return fmt.Errorf("%w: %w", apperrors.ErrConflict, err)
            }
            return err
        }

        return nil
    })
}
