package run_product_import

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	catalogaccess "github.com/NikolayNam/collabsphere/internal/catalog/application/access"
	catalogerrors "github.com/NikolayNam/collabsphere/internal/catalog/application/errors"
	"github.com/NikolayNam/collabsphere/internal/catalog/application/ports"
	productimport "github.com/NikolayNam/collabsphere/internal/catalog/application/product_import"
	catalogdomain "github.com/NikolayNam/collabsphere/internal/catalog/domain"
	orgdomain "github.com/NikolayNam/collabsphere/internal/organizations/domain"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/fault"
	"github.com/google/uuid"
)

type Handler struct {
	repo          ports.CatalogRepository
	organizations ports.OrganizationReader
	memberships   ports.MembershipReader
	clock         ports.Clock
	storage       ports.ObjectStorage
}

type importRow struct {
	RowNo              int
	CategoryCode       string
	CategoryName       *string
	CategoryParentCode *string
	CategorySortOrder  *int64
	ProductName        *string
	Description        *string
	SKU                *string
	PriceAmount        *string
	CurrencyCode       *string
	IsActive           *bool
}

type categorySpec struct {
	Code       string
	Name       *string
	ParentCode *string
	SortOrder  *int64
	RowNos     []int
}

type importIssue struct {
	RowNo   *int
	Code    string
	Message string
	Details map[string]any
}

type importStats struct {
	CategoriesCreated int
	CategoriesUpdated int
	ProductsCreated   int
	ProductsUpdated   int
}

type headerIndexes struct {
	categoryCode       int
	categoryName       int
	categoryParentCode int
	categorySortOrder  int
	productName        int
	description        int
	sku                int
	priceAmount        int
	currencyCode       int
	isActive           int
}

func NewHandler(repo ports.CatalogRepository, organizations ports.OrganizationReader, memberships ports.MembershipReader, clock ports.Clock, storage ports.ObjectStorage) *Handler {
	return &Handler{repo: repo, organizations: organizations, memberships: memberships, clock: clock, storage: storage}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*productimport.View, error) {
	if err := catalogaccess.RequireOrganizationAccess(ctx, h.organizations, h.memberships, cmd.OrganizationID, cmd.ActorAccountID, true); err != nil {
		return nil, err
	}
	if h.storage == nil {
		return nil, catalogerrors.ProductImportUnavailable()
	}

	mode := normalizeMode(cmd.Mode)
	if mode != "upsert" {
		return nil, catalogerrors.ProductImportFileInvalid("Only import mode 'upsert' is supported")
	}

	source, err := h.repo.GetStorageObjectByID(ctx, cmd.OrganizationID, cmd.SourceObjectID)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, catalogerrors.ProductImportSourceObjectNotFound()
	}
	if !isCSVSource(source.FileName, source.ContentType) {
		return nil, catalogerrors.ProductImportFileInvalid("Only CSV import files are supported")
	}

	now := h.clock.Now()
	startedBy := "api"
	modeCopy := mode
	batch := &ports.ProductImportBatch{
		ID:                 uuid.New(),
		OrganizationID:     cmd.OrganizationID,
		SourceObjectID:     source.ID,
		CreatedByAccountID: cmd.ActorAccountID,
		Status:             ports.ProductImportStatusProcessing,
		ProcessedRows:      0,
		SuccessRows:        0,
		ErrorRows:          0,
		StartedBy:          &startedBy,
		StartedAt:          now,
		CreatedAt:          now,
		UpdatedAt:          timePtr(now),
		Mode:               &modeCopy,
		ResultSummary:      map[string]any{},
	}
	if err := h.repo.CreateProductImportBatch(ctx, batch); err != nil {
		return nil, err
	}

	return h.run(ctx, batch, source)
}

func (h *Handler) run(ctx context.Context, batch *ports.ProductImportBatch, source *ports.StorageObject) (*productimport.View, error) {
	reader, err := h.storage.ReadObject(ctx, source.Bucket, source.ObjectKey)
	if err != nil {
		return h.finalizeBatch(ctx, batch, nil, nil, importIssue{
			Code:    "download_failed",
			Message: "Failed to download import source object",
			Details: map[string]any{"cause": err.Error()},
		}, true, importStats{})
	}
	defer reader.Close()

	rows, parseIssues, fatalIssue := parseImportCSV(reader)
	if fatalIssue != nil {
		return h.finalizeBatch(ctx, batch, rows, parseIssues, *fatalIssue, true, importStats{})
	}

	stats, rowIssues, fatalIssue := h.applyRows(ctx, batch.OrganizationID, rows, parseIssues)
	if fatalIssue != nil {
		return h.finalizeBatch(ctx, batch, rows, rowIssues, *fatalIssue, true, stats)
	}
	return h.finalizeBatch(ctx, batch, rows, rowIssues, importIssue{}, false, stats)
}

func (h *Handler) applyRows(ctx context.Context, organizationID orgdomain.OrganizationID, rows []importRow, initial []importIssue) (importStats, []importIssue, *importIssue) {
	stats := importStats{}
	issues := append([]importIssue(nil), initial...)
	issueRows := map[int]struct{}{}
	for _, issue := range issues {
		if issue.RowNo != nil {
			issueRows[*issue.RowNo] = struct{}{}
		}
	}

	addRowIssue := func(rowNo int, code, message string, details map[string]any) {
		if _, ok := issueRows[rowNo]; ok {
			return
		}
		issueRows[rowNo] = struct{}{}
		issues = append(issues, importIssue{
			RowNo:   intPtr(rowNo),
			Code:    code,
			Message: message,
			Details: copyMap(details),
		})
	}

	categories, err := h.repo.ListProductCategories(ctx, organizationID)
	if err != nil {
		fatal := importIssue{Code: "load_categories_failed", Message: "Failed to load organization product categories", Details: map[string]any{"cause": err.Error()}}
		return stats, issues, &fatal
	}
	categoriesByCode := make(map[string]*catalogdomain.ProductCategory, len(categories))
	for i := range categories {
		category := categories[i]
		copyCategory := category
		categoriesByCode[category.Code()] = &copyCategory
	}

	specs := map[string]*categorySpec{}
	for _, row := range rows {
		if _, ok := issueRows[row.RowNo]; ok {
			continue
		}
		if row.CategoryCode == "" {
			continue
		}
		if row.CategoryParentCode != nil && *row.CategoryParentCode == row.CategoryCode {
			addRowIssue(row.RowNo, "invalid_category_parent", "categoryParentCode cannot reference the category itself", nil)
			continue
		}

		spec, ok := specs[row.CategoryCode]
		if !ok {
			spec = &categorySpec{Code: row.CategoryCode}
			specs[row.CategoryCode] = spec
		}
		spec.RowNos = appendUniqueRowNo(spec.RowNos, row.RowNo)

		if row.CategoryName != nil {
			if spec.Name != nil && *spec.Name != *row.CategoryName {
				addRowIssue(row.RowNo, "conflicting_category_name", "Conflicting categoryName values for the same categoryCode", map[string]any{"categoryCode": row.CategoryCode})
			} else if spec.Name == nil {
				spec.Name = cloneStringPtr(row.CategoryName)
			}
		}
		if row.CategoryParentCode != nil {
			if spec.ParentCode != nil && *spec.ParentCode != *row.CategoryParentCode {
				addRowIssue(row.RowNo, "conflicting_category_parent", "Conflicting categoryParentCode values for the same categoryCode", map[string]any{"categoryCode": row.CategoryCode})
			} else if spec.ParentCode == nil {
				spec.ParentCode = cloneStringPtr(row.CategoryParentCode)
			}
		}
		if row.CategorySortOrder != nil {
			if spec.SortOrder != nil && *spec.SortOrder != *row.CategorySortOrder {
				addRowIssue(row.RowNo, "conflicting_category_sort_order", "Conflicting categorySortOrder values for the same categoryCode", map[string]any{"categoryCode": row.CategoryCode})
			} else if spec.SortOrder == nil {
				value := *row.CategorySortOrder
				spec.SortOrder = &value
			}
		}
	}

	pending := make(map[string]*categorySpec, len(specs))
	for code, spec := range specs {
		pending[code] = spec
	}

	now := h.clock.Now()
	for len(pending) > 0 {
		progressed := false
		for code, spec := range pending {
			validRows := filterValidRows(spec.RowNos, issueRows)
			if len(validRows) == 0 {
				delete(pending, code)
				progressed = true
				continue
			}

			var parentID *catalogdomain.ProductCategoryID
			if spec.ParentCode != nil {
				parent := categoriesByCode[*spec.ParentCode]
				if parent == nil {
					if _, waiting := pending[*spec.ParentCode]; waiting {
						continue
					}
					for _, rowNo := range validRows {
						addRowIssue(rowNo, "parent_category_not_found", "categoryParentCode references a category that does not exist", map[string]any{"categoryCode": code, "parentCode": *spec.ParentCode})
					}
					delete(pending, code)
					progressed = true
					continue
				}
				parentCopy := parent.ID()
				parentID = &parentCopy
			}

			existing := categoriesByCode[code]
			if existing == nil {
				if spec.Name == nil {
					for _, rowNo := range validRows {
						addRowIssue(rowNo, "missing_category_name", "categoryName is required for a new category", map[string]any{"categoryCode": code})
					}
					delete(pending, code)
					progressed = true
					continue
				}

				sortOrder := int64(0)
				if spec.SortOrder != nil {
					sortOrder = *spec.SortOrder
				}
				category, err := catalogdomain.NewProductCategory(catalogdomain.NewProductCategoryParams{
					ID:             catalogdomain.NewProductCategoryID(),
					OrganizationID: organizationID,
					ParentID:       parentID,
					Code:           code,
					Name:           *spec.Name,
					SortOrder:      sortOrder,
					Now:            now,
				})
				if err != nil {
					for _, rowNo := range validRows {
						addRowIssue(rowNo, "invalid_category", "Invalid product category data", map[string]any{"categoryCode": code})
					}
					delete(pending, code)
					progressed = true
					continue
				}
				if err := h.repo.CreateProductCategory(ctx, category); err != nil {
					if issue, ok := rowIssueFromError(validRows[0], err, "category_create_failed"); ok {
						for _, rowNo := range validRows {
							addRowIssue(rowNo, issue.Code, issue.Message, issue.Details)
						}
						delete(pending, code)
						progressed = true
						continue
					}
					fatal := importIssue{Code: "category_create_failed", Message: "Failed to create product category", Details: map[string]any{"categoryCode": code, "cause": err.Error()}}
					return stats, issues, &fatal
				}

				categoriesByCode[code] = category
				stats.CategoriesCreated++
				delete(pending, code)
				progressed = true
				continue
			}

			nextName := existing.Name()
			if spec.Name != nil {
				nextName = *spec.Name
			}
			nextSortOrder := existing.SortOrder()
			if spec.SortOrder != nil {
				nextSortOrder = *spec.SortOrder
			}
			nextParentID := existing.ParentID()
			if parentID != nil {
				nextParentID = parentID
			}
			if sameCategoryID(existing.ParentID(), nextParentID) && nextName == existing.Name() && nextSortOrder == existing.SortOrder() {
				delete(pending, code)
				progressed = true
				continue
			}

			updated, err := catalogdomain.RehydrateProductCategory(catalogdomain.RehydrateProductCategoryParams{
				ID:             existing.ID(),
				OrganizationID: organizationID,
				ParentID:       nextParentID,
				TemplateID:     existing.TemplateID(),
				Code:           code,
				Name:           nextName,
				SortOrder:      nextSortOrder,
				CreatedAt:      existing.CreatedAt(),
				UpdatedAt:      now,
			})
			if err != nil {
				for _, rowNo := range validRows {
					addRowIssue(rowNo, "invalid_category", "Invalid product category data", map[string]any{"categoryCode": code})
				}
				delete(pending, code)
				progressed = true
				continue
			}
			if err := h.repo.UpdateProductCategory(ctx, updated); err != nil {
				if issue, ok := rowIssueFromError(validRows[0], err, "category_update_failed"); ok {
					for _, rowNo := range validRows {
						addRowIssue(rowNo, issue.Code, issue.Message, issue.Details)
					}
					delete(pending, code)
					progressed = true
					continue
				}
				fatal := importIssue{Code: "category_update_failed", Message: "Failed to update product category", Details: map[string]any{"categoryCode": code, "cause": err.Error()}}
				return stats, issues, &fatal
			}

			categoriesByCode[code] = updated
			stats.CategoriesUpdated++
			delete(pending, code)
			progressed = true
		}

		if !progressed {
			for code, spec := range pending {
				for _, rowNo := range filterValidRows(spec.RowNos, issueRows) {
					addRowIssue(rowNo, "unresolved_category_dependency", "categoryCode cannot be resolved because of missing or cyclic parent references", map[string]any{"categoryCode": code})
				}
			}
			break
		}
	}

	products, err := h.repo.ListProducts(ctx, organizationID)
	if err != nil {
		fatal := importIssue{Code: "load_products_failed", Message: "Failed to load organization products", Details: map[string]any{"cause": err.Error()}}
		return stats, issues, &fatal
	}
	productsBySKU := map[string]*catalogdomain.Product{}
	ambiguousSKUs := map[string]struct{}{}
	for i := range products {
		product := products[i]
		sku := product.SKU()
		if sku == nil {
			continue
		}
		key := *sku
		if _, exists := productsBySKU[key]; exists {
			ambiguousSKUs[key] = struct{}{}
			delete(productsBySKU, key)
			continue
		}
		copyProduct := product
		productsBySKU[key] = &copyProduct
	}

	now = h.clock.Now()
	for _, row := range rows {
		if _, ok := issueRows[row.RowNo]; ok {
			continue
		}
		if row.ProductName == nil {
			continue
		}

		var categoryID *catalogdomain.ProductCategoryID
		if row.CategoryCode != "" {
			category := categoriesByCode[row.CategoryCode]
			if category == nil {
				addRowIssue(row.RowNo, "product_category_not_found", "categoryCode references a category that was not imported", map[string]any{"categoryCode": row.CategoryCode})
				continue
			}
			categoryCopy := category.ID()
			categoryID = &categoryCopy
		}

		if row.SKU != nil {
			if _, ambiguous := ambiguousSKUs[*row.SKU]; ambiguous {
				addRowIssue(row.RowNo, "ambiguous_sku", "Multiple existing products share the same sku", map[string]any{"sku": *row.SKU})
				continue
			}
			existing := productsBySKU[*row.SKU]
			if existing != nil {
				isActive := existing.IsActive()
				if row.IsActive != nil {
					isActive = *row.IsActive
				}
				updated, err := catalogdomain.RehydrateProduct(catalogdomain.RehydrateProductParams{
					ID:             existing.ID(),
					OrganizationID: organizationID,
					CategoryID:     categoryID,
					Name:           *row.ProductName,
					Description:    cloneStringPtr(row.Description),
					SKU:            cloneStringPtr(row.SKU),
					PriceAmount:    cloneStringPtr(row.PriceAmount),
					CurrencyCode:   cloneStringPtr(row.CurrencyCode),
					IsActive:       isActive,
					CreatedAt:      existing.CreatedAt(),
					UpdatedAt:      now,
				})
				if err != nil {
					addRowIssue(row.RowNo, "invalid_product", "Invalid product data", map[string]any{"sku": *row.SKU})
					continue
				}
				if err := h.repo.UpdateProduct(ctx, updated); err != nil {
					if issue, ok := rowIssueFromError(row.RowNo, err, "product_update_failed"); ok {
						addRowIssue(row.RowNo, issue.Code, issue.Message, issue.Details)
						continue
					}
					fatal := importIssue{Code: "product_update_failed", Message: "Failed to update product", Details: map[string]any{"rowNo": row.RowNo, "cause": err.Error()}}
					return stats, issues, &fatal
				}
				productsBySKU[*row.SKU] = updated
				stats.ProductsUpdated++
				continue
			}
		}

		isActive := true
		if row.IsActive != nil {
			isActive = *row.IsActive
		}
		product, err := catalogdomain.NewProduct(catalogdomain.NewProductParams{
			ID:             catalogdomain.NewProductID(),
			OrganizationID: organizationID,
			CategoryID:     categoryID,
			Name:           *row.ProductName,
			Description:    cloneStringPtr(row.Description),
			SKU:            cloneStringPtr(row.SKU),
			PriceAmount:    cloneStringPtr(row.PriceAmount),
			CurrencyCode:   cloneStringPtr(row.CurrencyCode),
			IsActive:       boolPtr(isActive),
			Now:            now,
		})
		if err != nil {
			addRowIssue(row.RowNo, "invalid_product", "Invalid product data", map[string]any{"rowNo": row.RowNo})
			continue
		}
		if err := h.repo.CreateProduct(ctx, product); err != nil {
			if issue, ok := rowIssueFromError(row.RowNo, err, "product_create_failed"); ok {
				addRowIssue(row.RowNo, issue.Code, issue.Message, issue.Details)
				continue
			}
			fatal := importIssue{Code: "product_create_failed", Message: "Failed to create product", Details: map[string]any{"rowNo": row.RowNo, "cause": err.Error()}}
			return stats, issues, &fatal
		}
		if row.SKU != nil {
			productsBySKU[*row.SKU] = product
		}
		stats.ProductsCreated++
	}

	return stats, issues, nil
}

func (h *Handler) finalizeBatch(ctx context.Context, batch *ports.ProductImportBatch, rows []importRow, rowIssues []importIssue, fatalIssue importIssue, failed bool, stats importStats) (*productimport.View, error) {
	totalRows := len(rows)
	batch.TotalRows = intPtr(totalRows)
	batch.ProcessedRows = totalRows
	batch.ErrorRows = countRowIssues(rowIssues)
	if batch.ErrorRows > batch.ProcessedRows {
		batch.ErrorRows = batch.ProcessedRows
	}
	batch.SuccessRows = batch.ProcessedRows - batch.ErrorRows
	if batch.SuccessRows < 0 {
		batch.SuccessRows = 0
	}

	now := h.clock.Now()
	batch.FinishedAt = &now
	batch.UpdatedAt = &now
	batch.ResultSummary = map[string]any{
		"categoriesCreated": stats.CategoriesCreated,
		"categoriesUpdated": stats.CategoriesUpdated,
		"productsCreated":   stats.ProductsCreated,
		"productsUpdated":   stats.ProductsUpdated,
	}
	if failed {
		batch.Status = ports.ProductImportStatusFailed
		batch.ResultSummary["failed"] = true
	} else {
		batch.Status = ports.ProductImportStatusCompleted
	}

	errorsToStore := make([]ports.ProductImportErrorRecord, 0, len(rowIssues)+1)
	createdAt := h.clock.Now()
	for _, issue := range dedupeIssues(rowIssues) {
		errorsToStore = append(errorsToStore, toErrorRecord(batch.ID, issue, createdAt))
	}
	if fatalIssue.Message != "" {
		errorsToStore = append(errorsToStore, toErrorRecord(batch.ID, fatalIssue, createdAt))
	}

	if err := h.repo.UpdateProductImportBatch(ctx, batch); err != nil {
		return nil, err
	}
	if err := h.repo.AddProductImportErrors(ctx, batch.ID, errorsToStore); err != nil {
		return nil, err
	}

	persistedErrors, err := h.repo.ListProductImportErrors(ctx, batch.ID)
	if err != nil {
		return nil, err
	}
	return &productimport.View{Batch: batch, Errors: persistedErrors}, nil
}

func parseImportCSV(r io.Reader) ([]importRow, []importIssue, *importIssue) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err == io.EOF {
		issue := importIssue{Code: "empty_file", Message: "CSV file is empty"}
		return nil, nil, &issue
	}
	if err != nil {
		issue := importIssue{Code: "invalid_csv", Message: "Failed to read CSV header", Details: map[string]any{"cause": err.Error()}}
		return nil, nil, &issue
	}

	indexes, ok := resolveHeaderIndexes(header)
	if !ok {
		issue := importIssue{Code: "invalid_header", Message: "CSV header must include at least categoryCode or productName columns"}
		return nil, nil, &issue
	}

	rows := make([]importRow, 0)
	issues := make([]importIssue, 0)
	for rowNo := 2; ; rowNo++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			issue := importIssue{Code: "invalid_csv", Message: fmt.Sprintf("Failed to parse CSV row %d", rowNo), Details: map[string]any{"cause": err.Error()}}
			return rows, issues, &issue
		}
		if recordIsEmpty(record) {
			continue
		}

		row, issue := buildImportRow(indexes, record, rowNo)
		rows = append(rows, row)
		if issue != nil {
			issues = append(issues, *issue)
		}
	}

	return rows, issues, nil
}

func buildImportRow(indexes headerIndexes, record []string, rowNo int) (importRow, *importIssue) {
	row := importRow{
		RowNo:              rowNo,
		CategoryCode:       normalizeCell(getCell(record, indexes.categoryCode)),
		CategoryName:       stringPtrFromCell(getCell(record, indexes.categoryName)),
		CategoryParentCode: stringPtrFromCell(getCell(record, indexes.categoryParentCode)),
		ProductName:        stringPtrFromCell(getCell(record, indexes.productName)),
		Description:        stringPtrFromCell(getCell(record, indexes.description)),
		SKU:                stringPtrFromCell(getCell(record, indexes.sku)),
		PriceAmount:        stringPtrFromCell(getCell(record, indexes.priceAmount)),
		CurrencyCode:       stringPtrFromCell(getCell(record, indexes.currencyCode)),
	}

	if value := stringPtrFromCell(getCell(record, indexes.categorySortOrder)); value != nil {
		parsed, err := strconv.ParseInt(*value, 10, 64)
		if err != nil || parsed < 0 {
			return row, &importIssue{RowNo: intPtr(rowNo), Code: "invalid_category_sort_order", Message: "categorySortOrder must be a non-negative integer"}
		}
		row.CategorySortOrder = &parsed
	}
	if value := stringPtrFromCell(getCell(record, indexes.isActive)); value != nil {
		parsed, ok := parseBool(*value)
		if !ok {
			return row, &importIssue{RowNo: intPtr(rowNo), Code: "invalid_is_active", Message: "isActive must be true/false/1/0/yes/no"}
		}
		row.IsActive = &parsed
	}
	if row.CategoryCode == "" && row.ProductName == nil && row.CategoryName != nil {
		return row, &importIssue{RowNo: intPtr(rowNo), Code: "missing_category_code", Message: "categoryCode is required when categoryName is provided"}
	}
	if row.ProductName == nil && row.SKU != nil {
		return row, &importIssue{RowNo: intPtr(rowNo), Code: "missing_product_name", Message: "productName is required when sku is provided"}
	}

	return row, nil
}

func resolveHeaderIndexes(header []string) (headerIndexes, bool) {
	indexes := headerIndexes{categoryCode: -1, categoryName: -1, categoryParentCode: -1, categorySortOrder: -1, productName: -1, description: -1, sku: -1, priceAmount: -1, currencyCode: -1, isActive: -1}

	for idx, raw := range header {
		normalized := normalizeHeader(raw)
		switch normalized {
		case "categorycode", "producttypecode":
			indexes.categoryCode = idx
		case "categoryname", "producttypename":
			indexes.categoryName = idx
		case "categoryparentcode", "producttypeparentcode":
			indexes.categoryParentCode = idx
		case "categorysortorder", "producttypesortorder":
			indexes.categorySortOrder = idx
		case "productname", "name":
			indexes.productName = idx
		case "description":
			indexes.description = idx
		case "sku":
			indexes.sku = idx
		case "priceamount", "price":
			indexes.priceAmount = idx
		case "currencycode", "currency":
			indexes.currencyCode = idx
		case "isactive", "active":
			indexes.isActive = idx
		}
	}

	return indexes, indexes.categoryCode >= 0 || indexes.productName >= 0
}

func normalizeHeader(value string) string {
	value = strings.TrimPrefix(value, "\ufeff")
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer("_", "", "-", "", " ", "")
	return replacer.Replace(value)
}

func recordIsEmpty(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func getCell(record []string, idx int) string {
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return record[idx]
}

func stringPtrFromCell(value string) *string {
	trimmed := normalizeCell(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeCell(value string) string {
	return strings.TrimSpace(value)
}

func parseBool(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y":
		return true, true
	case "0", "false", "no", "n":
		return false, true
	default:
		return false, false
	}
}

func isCSVSource(fileName string, contentType *string) bool {
	if strings.HasSuffix(strings.ToLower(strings.TrimSpace(fileName)), ".csv") {
		return true
	}
	if contentType == nil {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(*contentType))
	return strings.Contains(lower, "csv") || strings.Contains(lower, "text/plain")
}

func normalizeMode(mode *string) string {
	if mode == nil {
		return "upsert"
	}
	trimmed := strings.ToLower(strings.TrimSpace(*mode))
	if trimmed == "" {
		return "upsert"
	}
	return trimmed
}

func dedupeIssues(items []importIssue) []importIssue {
	if len(items) == 0 {
		return nil
	}

	rowSeen := map[int]struct{}{}
	out := make([]importIssue, 0, len(items))
	for _, item := range items {
		if item.RowNo == nil {
			out = append(out, item)
			continue
		}
		if _, ok := rowSeen[*item.RowNo]; ok {
			continue
		}
		rowSeen[*item.RowNo] = struct{}{}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RowNo == nil {
			return true
		}
		if out[j].RowNo == nil {
			return false
		}
		return *out[i].RowNo < *out[j].RowNo
	})
	return out
}

func countRowIssues(items []importIssue) int {
	seen := map[int]struct{}{}
	for _, item := range items {
		if item.RowNo == nil {
			continue
		}
		seen[*item.RowNo] = struct{}{}
	}
	return len(seen)
}

func rowIssueFromError(rowNo int, err error, fallbackCode string) (importIssue, bool) {
	appErr, ok := fault.As(err)
	if !ok {
		return importIssue{}, false
	}
	switch appErr.Kind {
	case fault.KindValidation, fault.KindConflict, fault.KindNotFound:
		code := fallbackCode
		if strings.TrimSpace(appErr.Code) != "" {
			code = strings.ToLower(appErr.Code)
		}
		return importIssue{RowNo: intPtr(rowNo), Code: code, Message: appErr.Message}, true
	default:
		return importIssue{}, false
	}
}

func toErrorRecord(batchID uuid.UUID, issue importIssue, now time.Time) ports.ProductImportErrorRecord {
	return ports.ProductImportErrorRecord{
		ID:        uuid.New(),
		BatchID:   batchID,
		RowNo:     issue.RowNo,
		Code:      stringPtr(issue.Code),
		Message:   issue.Message,
		Details:   copyMap(issue.Details),
		CreatedAt: now,
	}
}

func filterValidRows(rowNos []int, issueRows map[int]struct{}) []int {
	out := make([]int, 0, len(rowNos))
	for _, rowNo := range rowNos {
		if _, ok := issueRows[rowNo]; ok {
			continue
		}
		out = append(out, rowNo)
	}
	return out
}

func appendUniqueRowNo(items []int, value int) []int {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func sameCategoryID(left, right *catalogdomain.ProductCategoryID) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return left.UUID() == right.UUID()
	}
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func boolPtr(value bool) *bool {
	return &value
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func copyMap(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = item
	}
	return out
}

func intPtr(value int) *int {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
