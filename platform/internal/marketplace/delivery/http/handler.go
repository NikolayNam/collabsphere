package http

import (
	"context"
	"net/http"

	marketapp "github.com/NikolayNam/collabsphere/internal/marketplace/application"
	"github.com/NikolayNam/collabsphere/internal/marketplace/delivery/http/dto"
	"github.com/NikolayNam/collabsphere/internal/marketplace/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/google/uuid"
)

type Handler struct {
	svc *marketapp.Service
}

func NewHandler(svc *marketapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListOrders(ctx context.Context, input *dto.ListOrdersInput) (*dto.OrdersListResponse, error) {
	items, err := h.svc.ListOrders(ctx, input.Limit, input.Offset)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.OrdersListResponse{Status: http.StatusOK}
	out.Body.Items = make([]dto.OrderPayload, 0, len(items))
	for _, item := range items {
		out.Body.Items = append(out.Body.Items, toOrderPayload(item, nil, nil))
	}
	return out, nil
}

func (h *Handler) CreateOrder(ctx context.Context, input *dto.CreateOrderInput) (*dto.OrderResponse, error) {
	actorID, err := actorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseUUID(input.Body.OrganizationID, "Invalid organizationId")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items := make([]marketapp.CreateOrderItemInput, 0, len(input.Body.Items))
	for _, item := range input.Body.Items {
		var categoryID *uuid.UUID
		if item.CategoryID != nil {
			id, parseErr := parseUUID(*item.CategoryID, "Invalid categoryId")
			if parseErr != nil {
				return nil, humaerr.From(ctx, parseErr)
			}
			categoryID = &id
		}
		items = append(items, marketapp.CreateOrderItemInput{
			CategoryID:  categoryID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			Note:        item.Note,
		})
	}
	order, err := h.svc.CreateOrder(ctx, marketapp.CreateOrderCmd{
		ActorAccountID: actorID,
		OrganizationID: organizationID,
		Title:          input.Body.Title,
		Description:    input.Body.Description,
		BudgetAmount:   input.Body.BudgetAmount,
		CurrencyCode:   input.Body.CurrencyCode,
		Comment:        input.Body.Comment,
		Items:          items,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.OrderResponse{Status: http.StatusCreated, Body: toOrderPayload(*order, nil, nil)}, nil
}

func (h *Handler) GetOrder(ctx context.Context, input *dto.GetOrderInput) (*dto.OrderResponse, error) {
	orderID, err := parseUUID(input.OrderID, "Invalid order_id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	order, items, comments, err := h.svc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.OrderResponse{Status: http.StatusOK, Body: toOrderPayload(*order, items, comments)}, nil
}

func (h *Handler) AddOrderComment(ctx context.Context, input *dto.AddOrderCommentInput) (*dto.EmptyResponse, error) {
	actorID, err := actorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orderID, err := parseUUID(input.OrderID, "Invalid order_id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseUUID(input.Body.OrganizationID, "Invalid organizationId")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	if err := h.svc.AddOrderComment(ctx, actorID, organizationID, orderID, input.Body.Comment); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: http.StatusCreated}, nil
}

func (h *Handler) CreateOffer(ctx context.Context, input *dto.CreateOfferInput) (*dto.OfferResponse, error) {
	actorID, err := actorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	orderID, err := parseUUID(input.OrderID, "Invalid order_id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseUUID(input.Body.OrganizationID, "Invalid organizationId")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items := make([]marketapp.CreateOfferItemInput, 0, len(input.Body.Items))
	for _, item := range input.Body.Items {
		var categoryID *uuid.UUID
		if item.CategoryID != nil {
			id, parseErr := parseUUID(*item.CategoryID, "Invalid categoryId")
			if parseErr != nil {
				return nil, humaerr.From(ctx, parseErr)
			}
			categoryID = &id
		}
		var productID *uuid.UUID
		if item.ProductID != nil {
			id, parseErr := parseUUID(*item.ProductID, "Invalid productId")
			if parseErr != nil {
				return nil, humaerr.From(ctx, parseErr)
			}
			productID = &id
		}
		items = append(items, marketapp.CreateOfferItemInput{
			CategoryID:   categoryID,
			ProductID:    productID,
			CustomTitle:  item.CustomTitle,
			Quantity:     item.Quantity,
			Unit:         item.Unit,
			PriceAmount:  item.PriceAmount,
			CurrencyCode: item.CurrencyCode,
			Note:         item.Note,
		})
	}
	offer, err := h.svc.CreateOffer(ctx, marketapp.CreateOfferCmd{
		ActorAccountID: actorID,
		OrganizationID: organizationID,
		OrderID:        orderID,
		Comment:        input.Body.Comment,
		Items:          items,
	})
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.OfferResponse{Status: http.StatusCreated, Body: toOfferPayload(*offer, nil, nil)}, nil
}

func (h *Handler) ListOffers(ctx context.Context, input *dto.ListOffersInput) (*dto.OffersListResponse, error) {
	orderID, err := parseUUID(input.OrderID, "Invalid order_id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	items, err := h.svc.ListOffers(ctx, orderID)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	out := &dto.OffersListResponse{Status: http.StatusOK}
	out.Body.Items = make([]dto.OfferPayload, 0, len(items))
	for _, item := range items {
		offerItems, comments, detailErr := h.svc.GetOfferDetails(ctx, item.ID)
		if detailErr != nil {
			return nil, humaerr.From(ctx, detailErr)
		}
		out.Body.Items = append(out.Body.Items, toOfferPayload(item, offerItems, comments))
	}
	return out, nil
}

func (h *Handler) AddOfferComment(ctx context.Context, input *dto.AddOfferCommentInput) (*dto.EmptyResponse, error) {
	actorID, err := actorAccountID(ctx)
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	offerID, err := parseUUID(input.OfferID, "Invalid offer_id")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	organizationID, err := parseUUID(input.Body.OrganizationID, "Invalid organizationId")
	if err != nil {
		return nil, humaerr.From(ctx, err)
	}
	if err := h.svc.AddOfferComment(ctx, actorID, organizationID, offerID, input.Body.Comment); err != nil {
		return nil, humaerr.From(ctx, err)
	}
	return &dto.EmptyResponse{Status: http.StatusCreated}, nil
}

func actorAccountID(ctx context.Context) (uuid.UUID, error) {
	principal := authmw.PrincipalFromContext(ctx)
	if !principal.IsAccount() {
		return uuid.Nil, fault.Unauthorized("Authentication required")
	}
	return principal.AccountID, nil
}

func parseUUID(raw string, message string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fault.Validation(message)
	}
	return id, nil
}

func toOrderPayload(order domain.Order, items []domain.OrderItem, comments []domain.BoardComment) dto.OrderPayload {
	out := dto.OrderPayload{
		ID:             order.ID.String(),
		OrganizationID: order.OrganizationID.String(),
		Number:         order.Number,
		Title:          order.Title,
		Description:    order.Description,
		Status:         order.Status,
		BudgetAmount:   order.BudgetAmount,
		CurrencyCode:   order.CurrencyCode,
		CreatedAt:      dto.FormatTime(order.CreatedAt),
		Items:          make([]dto.OrderItemPayload, 0, len(items)),
		Comments:       make([]dto.BoardCommentPayload, 0, len(comments)),
	}
	for _, item := range items {
		out.Items = append(out.Items, dto.OrderItemPayload{
			ID:          item.ID.String(),
			CategoryID:  toUUIDString(item.CategoryID),
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			Note:        item.Note,
			SortOrder:   item.SortOrder,
		})
	}
	for _, item := range comments {
		out.Comments = append(out.Comments, toCommentPayload(item))
	}
	return out
}

func toOfferPayload(offer domain.Offer, items []domain.OfferItem, comments []domain.BoardComment) dto.OfferPayload {
	out := dto.OfferPayload{
		ID:             offer.ID.String(),
		OrderID:        offer.OrderID.String(),
		OrganizationID: offer.OrganizationID.String(),
		Status:         offer.Status,
		Comment:        offer.Comment,
		CreatedBy:      toUUIDString(offer.CreatedBy),
		CreatedAt:      dto.FormatTime(offer.CreatedAt),
		Items:          make([]dto.OfferItemPayload, 0, len(items)),
		Comments:       make([]dto.BoardCommentPayload, 0, len(comments)),
	}
	for _, item := range items {
		out.Items = append(out.Items, dto.OfferItemPayload{
			ID:           item.ID.String(),
			CategoryID:   toUUIDString(item.CategoryID),
			ProductID:    toUUIDString(item.ProductID),
			CustomTitle:  item.CustomTitle,
			Quantity:     item.Quantity,
			Unit:         item.Unit,
			PriceAmount:  item.PriceAmount,
			CurrencyCode: item.CurrencyCode,
			Note:         item.Note,
			SortOrder:    item.SortOrder,
		})
	}
	for _, item := range comments {
		out.Comments = append(out.Comments, toCommentPayload(item))
	}
	return out
}

func toCommentPayload(item domain.BoardComment) dto.BoardCommentPayload {
	return dto.BoardCommentPayload{
		ID:               item.ID.String(),
		OrderID:          toUUIDString(item.OrderID),
		OfferID:          toUUIDString(item.OfferID),
		OrganizationID:   item.OrganizationID.String(),
		OrganizationName: item.OrganizationName,
		AccountID:        item.AccountID.String(),
		Comment:          item.Comment,
		CreatedAt:        dto.FormatTime(item.CreatedAt),
	}
}

func toUUIDString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	text := value.String()
	return &text
}
