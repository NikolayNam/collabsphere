package postgres

import "gorm.io/gorm"

type GroupRepo struct {
	db *gorm.DB
}

func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db: db}
}
