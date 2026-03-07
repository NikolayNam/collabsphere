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

	updateCategory := updateProductCategoryOp
	updateCategory.Middlewares = secured
	huma.Register(api, updateCategory, h.UpdateProductCategory)

	deleteCategory := deleteProductCategoryOp
	deleteCategory.Middlewares = secured
	huma.Register(api, deleteCategory, h.DeleteProductCategory)

	listCategories := listProductCategoriesOp
	listCategories.Middlewares = secured
	huma.Register(api, listCategories, h.ListProductCategories)

	createProduct := createProductOp
	createProduct.Middlewares = secured
	huma.Register(api, createProduct, h.CreateProduct)

	updateProduct := updateProductOp
	updateProduct.Middlewares = secured
	huma.Register(api, updateProduct, h.UpdateProduct)

	deleteProduct := deleteProductOp
	deleteProduct.Middlewares = secured
	huma.Register(api, deleteProduct, h.DeleteProduct)

	listProducts := listProductsOp
	listProducts.Middlewares = secured
	huma.Register(api, listProducts, h.ListProducts)

	getProduct := getProductByIDOp
	getProduct.Middlewares = secured
	huma.Register(api, getProduct, h.GetProductByID)

	createImportUpload := createProductImportUploadOp
	createImportUpload.Middlewares = secured
	huma.Register(api, createImportUpload, h.CreateProductImportUpload)

	runImport := runProductImportOp
	runImport.Middlewares = secured
	huma.Register(api, runImport, h.RunProductImport)

	getImport := getProductImportOp
	getImport.Middlewares = secured
	huma.Register(api, getImport, h.GetProductImport)
}
