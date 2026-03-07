package postgres

import "gorm.io/gorm"

type ProductCategoryRepo struct {
	db *gorm.DB
}

func NewProductCategoryRepo(db *gorm.DB) *ProductCategoryRepo {
	return &ProductCategoryRepo{db: db}
}
