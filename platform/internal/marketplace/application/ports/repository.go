package ports

import (
	"context"

	"github.com/NikolayNam/collabsphere/internal/marketplace/domain"
	"github.com/google/uuid"
)

type Repository interface {
	CreateOrder(ctx context.Context, order *domain.Order, items []domain.OrderItem, firstComment *domain.BoardComment) error
	ListOrders(ctx context.Context, limit, offset int) ([]domain.Order, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error)
	ListOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error)
	ListOrderComments(ctx context.Context, orderID uuid.UUID) ([]domain.BoardComment, error)

	CreateOffer(ctx context.Context, offer *domain.Offer, items []domain.OfferItem, firstComment *domain.BoardComment) error
	ListOffersByOrder(ctx context.Context, orderID uuid.UUID) ([]domain.Offer, error)
	ListOfferItems(ctx context.Context, offerID uuid.UUID) ([]domain.OfferItem, error)
	ListOfferComments(ctx context.Context, offerID uuid.UUID) ([]domain.BoardComment, error)

	AddComment(ctx context.Context, comment *domain.BoardComment) error
}
