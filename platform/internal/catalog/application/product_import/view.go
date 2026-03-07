package product_import

import "github.com/NikolayNam/collabsphere/internal/catalog/application/ports"

type View struct {
	Batch  *ports.ProductImportBatch
	Errors []ports.ProductImportErrorRecord
}
