package app

import (
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func registerAllModules(api huma.API, apiV1 chi.Router, router chi.Router, db *gorm.DB, conf *config.Config) {
	registerCoreModules(api, apiV1, db, conf)
	registerDomainModules(api, router, db, conf)
}

func registerCoreModules(api huma.API, apiV1 chi.Router, db *gorm.DB, conf *config.Config) {
	registerAuthModule(apiV1, api, db, conf)
	registerPlatformOpsModule(api, db, conf)
}

func registerDomainModules(api huma.API, router chi.Router, db *gorm.DB, conf *config.Config) {
	registerAccountsModule(api, db, conf)
	registerOrganzationsModule(api, db, conf)
	registerTenantsModule(api, db, conf)
	registerMembershipsModule(api, db, conf)
	registerCatalogModule(api, db, conf)
	registerMarketplaceModule(api, db, conf)
	registerGroupsModule(api, db, conf)
	registerCollabModule(api, router, db, conf)
	registerStorageModule(api, db, conf)
}
