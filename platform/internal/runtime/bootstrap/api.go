package bootstrap

import (
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"

	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/humaerr"
)

func NewAPI(router chi.Router, conf *config.Config) huma.API {
	humaerr.Install()

	makeTag := func(name, displayName, description string) *huma.Tag {
		tag := &huma.Tag{Name: name, Description: description}
		if displayName != "" {
			tag.Extensions = map[string]any{"x-displayName": displayName}
		}
		return tag
	}

	cfg := huma.DefaultConfig(conf.APP.Title, conf.APP.Version)
	cfg.CreateHooks = nil
	cfg.DocsPath = ""
	cfg.DocsRenderer = huma.DocsRendererScalar
	cfg.Info.Description = strings.TrimSpace(strings.Join([]string{
		"CollabSphere API управляет аккаунтами, организациями, каталогом и collaboration-сценариями.",
		"",
		"## Как читать API reference",
		"",
		"- Все маршруты живут под базовым префиксом `/api/v1`.",
		"- Защищенные ручки используют `Authorization: Bearer <token>`.",
		"- Загрузки файлов выполняются прямым `multipart/form-data` запросом в предметные upload endpoint'ы.",
		"- Скачивание файлов работает через предметные download endpoint'ы и возвращает short-lived `downloadUrl`.",
		"",
		"## Основные области",
		"",
		"- `Authentication`: вход, refresh, logout и текущий principal.",
		"- `Accounts`: профиль пользователя и связанные с ним файлы.",
		"- `Organizations`: профиль, доступы, onboarding, каталог и файлы организации.",
		"- `Groups` и `Collaboration`: каналы, сообщения, вложения, конференции и записи.",
		"",
		"## Рекомендованный порядок для новых интеграций",
		"",
		"1. Создать аккаунт и выполнить login.",
		"2. Создать организацию и назначить участников.",
		"3. Заполнить профиль организации и загрузить обязательные документы.",
		"4. Импортировать категории и товары.",
		"5. Использовать collab-раздел для каналов, сообщений, вложений и конференций.",
	}, "\n"))
	if cfg.Components == nil {
		cfg.Components = &huma.Components{}
	}
	cfg.Components.Schemas = huma.NewMapRegistry("#/components/schemas/", newSchemaNamer())
	if cfg.Components.SecuritySchemes == nil {
		cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{}
	}
	cfg.Components.SecuritySchemes["bearerAuth"] = &huma.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
	cfg.Tags = []*huma.Tag{
		makeTag("System", "Health", "Health and platform-level endpoints."),
		makeTag("Auth", "Sessions", "Authentication and session endpoints."),
		makeTag("Accounts", "Profile", "Account profile and identity endpoints."),
		makeTag("Accounts / Files", "Files", "Account-owned files such as avatars and account file listings."),
		makeTag("Organizations", "Profile", "Organization profile endpoints."),
		makeTag("Organizations / Members", "Access", "Organization membership and role management."),
		makeTag("Organizations / Files", "Files", "Organization files such as logos, legal documents, imports, and downloadable assets."),
		makeTag("Organizations / Onboarding", "Onboarding", "Cooperation application and legal document review flows."),
		makeTag("Organizations / Catalog", "Catalog", "Organization product categories, products, and import flows."),
		makeTag("Groups", "Groups", "Group lifecycle and membership endpoints."),
		makeTag("Collab / Channels", "Channels", "Channels, messages, reactions, read cursors, and guest invites."),
		makeTag("Collab / Conferences", "Conferences", "Conference lifecycle and transcript endpoints."),
		makeTag("Collab / Files", "Files", "Chat attachments and conference recordings."),
	}
	if cfg.Extensions == nil {
		cfg.Extensions = map[string]any{}
	}
	cfg.Extensions["x-tagGroups"] = []map[string]any{
		{"name": "Platform", "tags": []string{"System"}},
		{"name": "Authentication", "tags": []string{"Auth"}},
		{"name": "Accounts", "tags": []string{"Accounts", "Accounts / Files"}},
		{"name": "Organizations", "tags": []string{"Organizations", "Organizations / Members", "Organizations / Files", "Organizations / Onboarding", "Organizations / Catalog"}},
		{"name": "Groups", "tags": []string{"Groups"}},
		{"name": "Collaboration", "tags": []string{"Collab / Channels", "Collab / Conferences", "Collab / Files"}},
	}

	// важно: чтобы Swagger/SDK знали, что API живёт под /v1
	cfg.Servers = []*huma.Server{
		{URL: "/api/v1"},
	}

	return humachi.New(router, cfg)
}
