package postgres

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/marketplace/domain"
)

func (r *Repo) CreateOrder(ctx context.Context, order *domain.Order, items []domain.OrderItem, firstComment *domain.BoardComment) error {
	db := r.dbFrom(ctx).WithContext(ctx)
	if err := db.Table("sales.orders").Create(map[string]any{
		"id":              order.ID,
		"organization_id": order.OrganizationID,
		"number":          order.Number,
		"title":           order.Title,
		"description":     order.Description,
		"status":          order.Status,
		"budget_amount":   order.BudgetAmount,
		"currency_code":   order.CurrencyCode,
		"created_at":      order.CreatedAt,
		"updated_at":      order.CreatedAt,
	}).Error; err != nil {
		return err
	}
	for _, item := range items {
		if err := db.Table("sales.order_items").Create(map[string]any{
			"id":              item.ID,
			"organization_id": order.OrganizationID,
			"order_id":        order.ID,
			"category_id":     item.CategoryID,
			"product_name":    item.ProductName,
			"quantity":        item.Quantity,
			"unit":            item.Unit,
			"note":            item.Note,
			"sort_order":      item.SortOrder,
			"created_at":      order.CreatedAt,
			"updated_at":      order.CreatedAt,
		}).Error; err != nil {
			return err
		}
	}
	if firstComment != nil {
		return r.AddComment(ctx, firstComment)
	}
	return nil
}

func (r *Repo) CreateOffer(ctx context.Context, offer *domain.Offer, items []domain.OfferItem, firstComment *domain.BoardComment) error {
	db := r.dbFrom(ctx).WithContext(ctx)
	if err := db.Table("sales.order_offers").Create(map[string]any{
		"id":              offer.ID,
		"order_id":        offer.OrderID,
		"organization_id": offer.OrganizationID,
		"status":          offer.Status,
		"comment":         offer.Comment,
		"created_by":      offer.CreatedBy,
		"created_at":      offer.CreatedAt,
		"updated_at":      offer.CreatedAt,
	}).Error; err != nil {
		return err
	}
	for _, item := range items {
		if err := db.Table("sales.offer_items").Create(map[string]any{
			"id":              item.ID,
			"offer_id":        offer.ID,
			"organization_id": offer.OrganizationID,
			"category_id":     item.CategoryID,
			"product_id":      item.ProductID,
			"custom_title":    item.CustomTitle,
			"quantity":        item.Quantity,
			"unit":            item.Unit,
			"price_amount":    item.PriceAmount,
			"currency_code":   item.CurrencyCode,
			"note":            item.Note,
			"sort_order":      item.SortOrder,
			"created_at":      offer.CreatedAt,
			"updated_at":      offer.CreatedAt,
		}).Error; err != nil {
			return err
		}
	}
	if firstComment != nil {
		return r.AddComment(ctx, firstComment)
	}
	return nil
}

func (r *Repo) AddComment(ctx context.Context, comment *domain.BoardComment) error {
	return r.dbFrom(ctx).WithContext(ctx).Table("sales.board_comments").Create(map[string]any{
		"id":              comment.ID,
		"order_id":        comment.OrderID,
		"offer_id":        comment.OfferID,
		"organization_id": comment.OrganizationID,
		"account_id":      comment.AccountID,
		"comment":         comment.Comment,
		"created_at":      comment.CreatedAt,
	}).Error
}
