package domain

import (
	"time"

	"github.com/google/uuid"
)

// Organization — доменная сущность.
type Organization struct {
	UID uuid.UUID

	LegalName   string
	DisplayName *string

	// Регистрация/юрисдикция (1:1)
	CountryOfRegistration string // ISO2: "DE","US","RU"
	LegalEntityTypeID     *int
	TypeID                *int

	PrimaryEmail   string
	PrimaryAddress *string
	PrimaryPhone   *string
	PrimarySite    *string

	// Статус/верификация
	Status     string // Draft/Active/Suspended/Archived
	VerifiedAt *time.Time
	VerifiedBy *string

	// Аудит
	CreatedBy string
	CreatedAt time.Time

	UpdatedBy *string
	UpdatedAt *time.Time

	DeletedAt *time.Time

	Version int
}

// NewOrganization — фабрика: гарантирует корректные дефолты (UID, CreatedAt).
// Валидации минимальные; усиливай по необходимости.
func NewOrganization(
	legalName string,
	primaryEmail string,
	createdBy string,
) (*Organization, error) {
	if legalName == "" {
		return nil, NewInvalidLegalName()
	}
	if primaryEmail == "" {
		return nil, NewInvalidEmail() // <-- вместо “просто ошибки”
	}
	if createdBy == "" {
		return nil, NewCreatorLoginRequired()
	}

	now := time.Now().UTC()

	return &Organization{
		UID:          uuid.New(),
		LegalName:    legalName,
		PrimaryEmail: primaryEmail,
		CreatedBy:    createdBy,
		CreatedAt:    now,
	}, nil
}

func (o *Organization) ChangeLegalName(newLegalName, user string) error {
	if newLegalName == "" {
		return NewInvalidLegalName() // <-- доменная ошибка
	}
	o.LegalName = newLegalName
	o.touch(user)
	return nil
}

func (o *Organization) touch(user string) {
	o.UpdatedBy = new(user)
	o.UpdatedAt = new(time.Now().UTC())
}
