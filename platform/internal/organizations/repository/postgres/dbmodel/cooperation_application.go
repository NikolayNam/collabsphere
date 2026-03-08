package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type CooperationApplication struct {
	ID                    uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID        uuid.UUID  `gorm:"column:organization_id;type:uuid;not null"`
	Status                string     `gorm:"column:status;type:varchar(32);not null"`
	ConfirmationEmail     *string    `gorm:"column:confirmation_email;type:varchar(320)"`
	CompanyName           *string    `gorm:"column:company_name;type:varchar(255)"`
	RepresentedCategories *string    `gorm:"column:represented_categories;type:text"`
	MinimumOrderAmount    *string    `gorm:"column:minimum_order_amount;type:varchar(128)"`
	DeliveryGeography     *string    `gorm:"column:delivery_geography;type:text"`
	SalesChannels         []byte     `gorm:"column:sales_channels;type:jsonb;not null"`
	StorefrontURL         *string    `gorm:"column:storefront_url;type:varchar(512)"`
	ContactFirstName      *string    `gorm:"column:contact_first_name;type:varchar(128)"`
	ContactLastName       *string    `gorm:"column:contact_last_name;type:varchar(128)"`
	ContactJobTitle       *string    `gorm:"column:contact_job_title;type:varchar(128)"`
	PriceListObjectID     *uuid.UUID `gorm:"column:price_list_object_id;type:uuid"`
	ContactEmail          *string    `gorm:"column:contact_email;type:varchar(320)"`
	ContactPhone          *string    `gorm:"column:contact_phone;type:varchar(32)"`
	PartnerCode           *string    `gorm:"column:partner_code;type:varchar(128)"`
	ReviewNote            *string    `gorm:"column:review_note;type:text"`
	ReviewerAccountID     *uuid.UUID `gorm:"column:reviewer_account_id;type:uuid"`
	SubmittedAt           *time.Time `gorm:"column:submitted_at;type:timestamptz"`
	ReviewedAt            *time.Time `gorm:"column:reviewed_at;type:timestamptz"`
	CreatedAt             time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt             *time.Time `gorm:"column:updated_at;type:timestamptz"`
}

func (CooperationApplication) TableName() string { return "org.cooperation_applications" }
