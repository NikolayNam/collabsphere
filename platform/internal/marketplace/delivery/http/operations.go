package http

import "github.com/danielgtaylor/huma/v2"

var listOrdersOp = huma.Operation{
	OperationID: "sales-board-list-orders",
	Method:      "GET",
	Path:        "/sales/orders",
	Tags:        []string{"Sales Board"},
	Summary:     "List order board",
	Description: "Returns the marketplace order board with open and recent orders.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createOrderOp = huma.Operation{
	OperationID: "sales-board-create-order",
	Method:      "POST",
	Path:        "/sales/orders",
	Tags:        []string{"Sales Board"},
	Summary:     "Create order",
	Description: "Creates order with custom required list items and optional first comment.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var getOrderOp = huma.Operation{
	OperationID: "sales-board-get-order",
	Method:      "GET",
	Path:        "/sales/orders/{order_id}",
	Tags:        []string{"Sales Board"},
	Summary:     "Get order details",
	Description: "Returns order details with requirement items and discussion comments.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var addOrderCommentOp = huma.Operation{
	OperationID: "sales-board-add-order-comment",
	Method:      "POST",
	Path:        "/sales/orders/{order_id}/comments",
	Tags:        []string{"Sales Board"},
	Summary:     "Add order comment",
	Description: "Adds a comment to order discussion from selected organization context.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var createOfferOp = huma.Operation{
	OperationID: "sales-board-create-offer",
	Method:      "POST",
	Path:        "/sales/orders/{order_id}/offers",
	Tags:        []string{"Sales Board"},
	Summary:     "Create offer for order",
	Description: "Creates offer with items from needed product categories or custom offer list.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var listOffersOp = huma.Operation{
	OperationID: "sales-board-list-offers",
	Method:      "GET",
	Path:        "/sales/orders/{order_id}/offers",
	Tags:        []string{"Sales Board"},
	Summary:     "List offers for order",
	Description: "Returns offers posted against the order with their items and comments.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}

var addOfferCommentOp = huma.Operation{
	OperationID: "sales-board-add-offer-comment",
	Method:      "POST",
	Path:        "/sales/offers/{offer_id}/comments",
	Tags:        []string{"Sales Board"},
	Summary:     "Add offer comment",
	Description: "Adds a comment inside an offer thread.",
	Security:    []map[string][]string{{"bearerAuth": {}}},
}
