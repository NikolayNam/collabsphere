package postgres

import (
	platformdomain "github.com/NikolayNam/collabsphere/internal/platformops/domain"
	"gorm.io/gorm"
)

type Repo struct {
	db                      *gorm.DB
	bootstrapAutoGrantRules []platformdomain.AutoGrantRule
}

type Option func(*Repo)

func WithBootstrapAutoGrantRules(rules []platformdomain.AutoGrantRule) Option {
	cloned := append([]platformdomain.AutoGrantRule{}, rules...)
	return func(r *Repo) {
		r.bootstrapAutoGrantRules = cloned
	}
}

func NewRepo(db *gorm.DB, opts ...Option) *Repo {
	repo := &Repo{db: db}
	for _, opt := range opts {
		if opt != nil {
			opt(repo)
		}
	}
	return repo
}
