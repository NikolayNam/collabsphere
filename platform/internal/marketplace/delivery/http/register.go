package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	secured := huma.Middlewares{authmw.HumaAuthOptional(verifier)}

	list := listOrdersOp
	list.Middlewares = secured
	huma.Register(api, list, h.ListOrders)

	create := createOrderOp
	create.Middlewares = secured
	huma.Register(api, create, h.CreateOrder)

	get := getOrderOp
	get.Middlewares = secured
	huma.Register(api, get, h.GetOrder)

	orderComment := addOrderCommentOp
	orderComment.Middlewares = secured
	huma.Register(api, orderComment, h.AddOrderComment)

	createOffer := createOfferOp
	createOffer.Middlewares = secured
	huma.Register(api, createOffer, h.CreateOffer)

	listOffers := listOffersOp
	listOffers.Middlewares = secured
	huma.Register(api, listOffers, h.ListOffers)

	offerComment := addOfferCommentOp
	offerComment.Middlewares = secured
	huma.Register(api, offerComment, h.AddOfferComment)
}
