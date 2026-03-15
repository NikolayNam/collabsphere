package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	accdomain "github.com/NikolayNam/collabsphere/internal/accounts/domain"
	"github.com/NikolayNam/collabsphere/internal/marketplace/application/ports"
	"github.com/NikolayNam/collabsphere/internal/marketplace/domain"
	memberports "github.com/NikolayNam/collabsphere/internal/memberships/application/ports"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	sharedtx "github.com/NikolayNam/collabsphere/shared/tx"
	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type Service struct {
	repo        ports.Repository
	memberships memberports.MembershipRepository
	tx          sharedtx.Manager
	clock       Clock
}

func New(repo ports.Repository, memberships memberports.MembershipRepository, tx sharedtx.Manager, clock Clock) *Service {
	return &Service{repo: repo, memberships: memberships, tx: tx, clock: clock}
}

type CreateOrderItemInput struct {
	CategoryID  *uuid.UUID
	ProductName *string
	Quantity    *string
	Unit        *string
	Note        *string
}

type CreateOrderCmd struct {
	ActorAccountID uuid.UUID
	OrganizationID uuid.UUID
	Title          string
	Description    *string
	BudgetAmount   *string
	CurrencyCode   *string
	Comment        *string
	Items          []CreateOrderItemInput
}

type CreateOfferItemInput struct {
	CategoryID   *uuid.UUID
	ProductID    *uuid.UUID
	CustomTitle  *string
	Quantity     *string
	Unit         *string
	PriceAmount  *string
	CurrencyCode *string
	Note         *string
}

type CreateOfferCmd struct {
	ActorAccountID uuid.UUID
	OrganizationID uuid.UUID
	OrderID        uuid.UUID
	Comment        *string
	Items          []CreateOfferItemInput
}

func (s *Service) CreateOrder(ctx context.Context, cmd CreateOrderCmd) (*domain.Order, error) {
	if err := s.requireOrgMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID); err != nil {
		return nil, err
	}
	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return nil, fault.Validation("Order title is required")
	}
	now := s.clock.Now()
	order := &domain.Order{
		ID:             uuid.New(),
		OrganizationID: cmd.OrganizationID,
		Number:         fmt.Sprintf("ORD-%s", now.UTC().Format("20060102-150405")),
		Title:          title,
		Description:    normalizeOptional(cmd.Description),
		Status:         "open",
		BudgetAmount:   normalizeOptional(cmd.BudgetAmount),
		CurrencyCode:   normalizeOptional(cmd.CurrencyCode),
		CreatedAt:      now,
	}
	items := make([]domain.OrderItem, 0, len(cmd.Items))
	for idx, item := range cmd.Items {
		items = append(items, domain.OrderItem{
			ID:          uuid.New(),
			OrderID:     order.ID,
			CategoryID:  item.CategoryID,
			ProductName: normalizeOptional(item.ProductName),
			Quantity:    normalizeOptional(item.Quantity),
			Unit:        normalizeOptional(item.Unit),
			Note:        normalizeOptional(item.Note),
			SortOrder:   idx,
		})
	}
	var firstComment *domain.BoardComment
	if value := normalizeOptional(cmd.Comment); value != nil {
		firstComment = &domain.BoardComment{
			ID:             uuid.New(),
			OrderID:        &order.ID,
			OrganizationID: cmd.OrganizationID,
			AccountID:      cmd.ActorAccountID,
			Comment:        *value,
			CreatedAt:      now,
		}
	}
	if err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.repo.CreateOrder(ctx, order, items, firstComment)
	}); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *Service) ListOrders(ctx context.Context, limit, offset int) ([]domain.Order, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListOrders(ctx, limit, offset)
}

func (s *Service) GetOrder(ctx context.Context, orderID uuid.UUID) (*domain.Order, []domain.OrderItem, []domain.BoardComment, error) {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, nil, nil, err
	}
	if order == nil {
		return nil, nil, nil, fault.NotFound("Order not found")
	}
	items, err := s.repo.ListOrderItems(ctx, orderID)
	if err != nil {
		return nil, nil, nil, err
	}
	comments, err := s.repo.ListOrderComments(ctx, orderID)
	if err != nil {
		return nil, nil, nil, err
	}
	return order, items, comments, nil
}

func (s *Service) CreateOffer(ctx context.Context, cmd CreateOfferCmd) (*domain.Offer, error) {
	if err := s.requireOrgMembership(ctx, cmd.OrganizationID, cmd.ActorAccountID); err != nil {
		return nil, err
	}
	if cmd.OrderID == uuid.Nil {
		return nil, fault.Validation("Order id is required")
	}
	now := s.clock.Now()
	offer := &domain.Offer{
		ID:             uuid.New(),
		OrderID:        cmd.OrderID,
		OrganizationID: cmd.OrganizationID,
		Status:         "submitted",
		Comment:        normalizeOptional(cmd.Comment),
		CreatedBy:      &cmd.ActorAccountID,
		CreatedAt:      now,
	}
	items := make([]domain.OfferItem, 0, len(cmd.Items))
	for idx, item := range cmd.Items {
		items = append(items, domain.OfferItem{
			ID:           uuid.New(),
			OfferID:      offer.ID,
			CategoryID:   item.CategoryID,
			ProductID:    item.ProductID,
			CustomTitle:  normalizeOptional(item.CustomTitle),
			Quantity:     normalizeOptional(item.Quantity),
			Unit:         normalizeOptional(item.Unit),
			PriceAmount:  normalizeOptional(item.PriceAmount),
			CurrencyCode: normalizeOptional(item.CurrencyCode),
			Note:         normalizeOptional(item.Note),
			SortOrder:    idx,
		})
	}
	var firstComment *domain.BoardComment
	if value := normalizeOptional(cmd.Comment); value != nil {
		firstComment = &domain.BoardComment{
			ID:             uuid.New(),
			OfferID:        &offer.ID,
			OrganizationID: cmd.OrganizationID,
			AccountID:      cmd.ActorAccountID,
			Comment:        *value,
			CreatedAt:      now,
		}
	}
	if err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		return s.repo.CreateOffer(ctx, offer, items, firstComment)
	}); err != nil {
		return nil, err
	}
	return offer, nil
}

func (s *Service) ListOffers(ctx context.Context, orderID uuid.UUID) ([]domain.Offer, error) {
	return s.repo.ListOffersByOrder(ctx, orderID)
}

func (s *Service) GetOfferDetails(ctx context.Context, offerID uuid.UUID) ([]domain.OfferItem, []domain.BoardComment, error) {
	items, err := s.repo.ListOfferItems(ctx, offerID)
	if err != nil {
		return nil, nil, err
	}
	comments, err := s.repo.ListOfferComments(ctx, offerID)
	if err != nil {
		return nil, nil, err
	}
	return items, comments, nil
}

func (s *Service) AddOrderComment(ctx context.Context, actorAccountID, organizationID, orderID uuid.UUID, comment string) error {
	if err := s.requireOrgMembership(ctx, organizationID, actorAccountID); err != nil {
		return err
	}
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return fault.Validation("Comment is required")
	}
	return s.repo.AddComment(ctx, &domain.BoardComment{
		ID:             uuid.New(),
		OrderID:        &orderID,
		OrganizationID: organizationID,
		AccountID:      actorAccountID,
		Comment:        comment,
		CreatedAt:      s.clock.Now(),
	})
}

func (s *Service) AddOfferComment(ctx context.Context, actorAccountID, organizationID, offerID uuid.UUID, comment string) error {
	if err := s.requireOrgMembership(ctx, organizationID, actorAccountID); err != nil {
		return err
	}
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return fault.Validation("Comment is required")
	}
	return s.repo.AddComment(ctx, &domain.BoardComment{
		ID:             uuid.New(),
		OfferID:        &offerID,
		OrganizationID: organizationID,
		AccountID:      actorAccountID,
		Comment:        comment,
		CreatedAt:      s.clock.Now(),
	})
}

func (s *Service) requireOrgMembership(ctx context.Context, organizationID, actorAccountID uuid.UUID) error {
	if organizationID == uuid.Nil {
		return fault.Validation("Organization id is required")
	}
	if actorAccountID == uuid.Nil {
		return fault.Unauthorized("Authentication required")
	}
	orgID, err := orgdomain.OrganizationIDFromUUID(organizationID)
	if err != nil {
		return fault.Validation("Invalid organization id")
	}
	accID, err := accdomain.AccountIDFromUUID(actorAccountID)
	if err != nil {
		return fault.Unauthorized("Authentication required")
	}
	member, err := s.memberships.GetMemberByAccount(ctx, orgID, accID)
	if err != nil {
		return fault.Internal("Get organization membership failed", fault.WithCause(err))
	}
	if member == nil || !member.IsActive() || member.IsRemoved() {
		return fault.Forbidden("Organization access denied")
	}
	return nil
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
