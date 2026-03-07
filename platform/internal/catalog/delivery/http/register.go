package http

import (
	authmw "github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/middleware"
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler, verifier authmw.AccessTokenVerifier) {
	secured := huma.Middlewares{authmw.HumaAuthOptional(verifier)}

	createCategory := createProductCategoryOp
	createCategory.Middlewares = secured
	huma.Register(api, createCategory, h.CreateProductCategory)

	listCategories := listProductCategoriesOp
	listCategories.Middlewares = secured
	huma.Register(api, listCategories, h.ListProductCategories)

	createProduct := createProductOp
	createProduct.Middlewares = secured
	huma.Register(api, createProduct, h.CreateProduct)

	listProducts := listProductsOp
	listProducts.Middlewares = secured
	huma.Register(api, listProducts, h.ListProducts)

	getProduct := getProductByIDOp
	getProduct.Middlewares = secured
	huma.Register(api, getProduct, h.GetProductByID)
}
