package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/NikolayNam/collabsphere/internal/marketplace/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type orderRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id"`
	Number         string     `gorm:"column:number"`
	Title          string     `gorm:"column:title"`
	Description    *string    `gorm:"column:description"`
	Status         string     `gorm:"column:status"`
	BudgetAmount   *string    `gorm:"column:budget_amount"`
	CurrencyCode   *string    `gorm:"column:currency_code"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

type orderItemRow struct {
	ID          uuid.UUID  `gorm:"column:id"`
	OrderID     uuid.UUID  `gorm:"column:order_id"`
	CategoryID  *uuid.UUID `gorm:"column:category_id"`
	ProductName *string    `gorm:"column:product_name"`
	Quantity    *string    `gorm:"column:quantity"`
	Unit        *string    `gorm:"column:unit"`
	Note        *string    `gorm:"column:note"`
	SortOrder   int        `gorm:"column:sort_order"`
}

type offerRow struct {
	ID             uuid.UUID  `gorm:"column:id"`
	OrderID        uuid.UUID  `gorm:"column:order_id"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id"`
	Status         string     `gorm:"column:status"`
	Comment        *string    `gorm:"column:comment"`
	CreatedBy      *uuid.UUID `gorm:"column:created_by"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
}

type offerItemRow struct {
	ID           uuid.UUID  `gorm:"column:id"`
	OfferID      uuid.UUID  `gorm:"column:offer_id"`
	CategoryID   *uuid.UUID `gorm:"column:category_id"`
	ProductID    *uuid.UUID `gorm:"column:product_id"`
	CustomTitle  *string    `gorm:"column:custom_title"`
	Quantity     *string    `gorm:"column:quantity"`
	Unit         *string    `gorm:"column:unit"`
	PriceAmount  *string    `gorm:"column:price_amount"`
	CurrencyCode *string    `gorm:"column:currency_code"`
	Note         *string    `gorm:"column:note"`
	SortOrder    int        `gorm:"column:sort_order"`
}

type commentRow struct {
	ID               uuid.UUID  `gorm:"column:id"`
	OrderID          *uuid.UUID `gorm:"column:order_id"`
	OfferID          *uuid.UUID `gorm:"column:offer_id"`
	OrganizationID   uuid.UUID  `gorm:"column:organization_id"`
	AccountID        uuid.UUID  `gorm:"column:account_id"`
	Comment          string     `gorm:"column:comment"`
	OrganizationName *string    `gorm:"column:organization_name"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
}

func (r *Repo) ListOrders(ctx context.Context, limit, offset int) ([]domain.Order, error) {
	var rows []orderRow
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("sales.orders").
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Order, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Order{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			Number:         row.Number,
			Title:          row.Title,
			Description:    row.Description,
			Status:         row.Status,
			BudgetAmount:   row.BudgetAmount,
			CurrencyCode:   row.CurrencyCode,
			CreatedAt:      row.CreatedAt,
		})
	}
	return out, nil
}

func (r *Repo) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	var row orderRow
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("sales.orders").
		Where("id = ? AND deleted_at IS NULL", orderID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.Order{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		Number:         row.Number,
		Title:          row.Title,
		Description:    row.Description,
		Status:         row.Status,
		BudgetAmount:   row.BudgetAmount,
		CurrencyCode:   row.CurrencyCode,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func (r *Repo) ListOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	var rows []orderItemRow
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("sales.order_items").
		Where("order_id = ? AND deleted_at IS NULL", orderID).
		Order("sort_order ASC, created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.OrderItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.OrderItem{
			ID:          row.ID,
			OrderID:     row.OrderID,
			CategoryID:  row.CategoryID,
			ProductName: row.ProductName,
			Quantity:    row.Quantity,
			Unit:        row.Unit,
			Note:        row.Note,
			SortOrder:   row.SortOrder,
		})
	}
	return out, nil
}

func (r *Repo) ListOrderComments(ctx context.Context, orderID uuid.UUID) ([]domain.BoardComment, error) {
	return r.listComments(ctx, "c.order_id = ?", orderID)
}

func (r *Repo) ListOfferComments(ctx context.Context, offerID uuid.UUID) ([]domain.BoardComment, error) {
	return r.listComments(ctx, "c.offer_id = ?", offerID)
}

func (r *Repo) listComments(ctx context.Context, where string, arg any) ([]domain.BoardComment, error) {
	var rows []commentRow
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("sales.board_comments c").
		Select("c.id, c.order_id, c.offer_id, c.organization_id, c.account_id, c.comment, c.created_at, o.name AS organization_name").
		Joins("LEFT JOIN org.organizations o ON o.id = c.organization_id").
		Where(where, arg).
		Order("c.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.BoardComment, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.BoardComment{
			ID:               row.ID,
			OrderID:          row.OrderID,
			OfferID:          row.OfferID,
			OrganizationID:   row.OrganizationID,
			AccountID:        row.AccountID,
			Comment:          row.Comment,
			OrganizationName: row.OrganizationName,
			CreatedAt:        row.CreatedAt,
		})
	}
	return out, nil
}

func (r *Repo) ListOffersByOrder(ctx context.Context, orderID uuid.UUID) ([]domain.Offer, error) {
	var rows []offerRow
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("sales.order_offers").
		Where("order_id = ? AND deleted_at IS NULL", orderID).
		Order("created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Offer, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Offer{
			ID:             row.ID,
			OrderID:        row.OrderID,
			OrganizationID: row.OrganizationID,
			Status:         row.Status,
			Comment:        row.Comment,
			CreatedBy:      row.CreatedBy,
			CreatedAt:      row.CreatedAt,
		})
	}
	return out, nil
}

func (r *Repo) ListOfferItems(ctx context.Context, offerID uuid.UUID) ([]domain.OfferItem, error) {
	var rows []offerItemRow
	err := r.dbFrom(ctx).WithContext(ctx).
		Table("sales.offer_items").
		Where("offer_id = ? AND deleted_at IS NULL", offerID).
		Order("sort_order ASC, created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.OfferItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.OfferItem{
			ID:           row.ID,
			OfferID:      row.OfferID,
			CategoryID:   row.CategoryID,
			ProductID:    row.ProductID,
			CustomTitle:  row.CustomTitle,
			Quantity:     row.Quantity,
			Unit:         row.Unit,
			PriceAmount:  row.PriceAmount,
			CurrencyCode: row.CurrencyCode,
			Note:         row.Note,
			SortOrder:    row.SortOrder,
		})
	}
	return out, nil
}
