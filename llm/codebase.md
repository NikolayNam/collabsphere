# Tree View:
```
.
├── deploy
│   ├── docker-compose.infrastructure.yaml
│   ├── docker-compose.migrate.yaml
│   ├── docker-compose.platform.yaml
│   └── docker-compose.proxy.yaml
├── llm
│   └── codebase.md
├── platform
│   ├── backend-context.md
│   ├── cmd
│   │   ├── api
│   │   │   └── main.go
│   │   ├── app
│   │   │   ├── api.go
│   │   │   ├── app.go
│   │   │   ├── db.go
│   │   │   └── router.go
│   │   ├── httpserver
│   │   │   └── server.go
│   │   └── migrate
│   │       └── main.go
│   ├── internal
│   │   ├── organizations
│   │   │   └── domain
│   │   │       ├── exceptions.go
│   │   │       └── organizations.go
│   │   ├── platform
│   │   │   ├── actorctx
│   │   │   │   └── actorctx.go
│   │   │   ├── config
│   │   │   │   └── config.go
│   │   │   ├── db
│   │   │   │   └── migrations
│   │   │   │       └── 0001_init.sql
│   │   │   ├── logger
│   │   │   │   ├── context.go
│   │   │   │   └── logger.go
│   │   │   └── middleware
│   │   │       ├── access_log.go
│   │   │       ├── logger_context.go
│   │   │       ├── middleware.go
│   │   │       ├── org.go
│   │   │       ├── ratelimit.go
│   │   │       └── security.go
│   │   ├── system
│   │   │   ├── dto.go
│   │   │   └── handlers.go
│   │   └── users
│   │       ├── application
│   │       │   ├── errors.go
│   │       │   └── service.go
│   │       ├── cache
│   │       │   ├── interface.go
│   │       │   └── redis
│   │       │       ├── keys.go
│   │       │       └── user_cache.go
│   │       ├── delivery
│   │       │   └── http
│   │       │       ├── dto.go
│   │       │       ├── handler.go
│   │       │       ├── operations.go
│   │       │       └── register.go
│   │       ├── domain
│   │       │   ├── email.go
│   │       │   ├── ids.go
│   │       │   ├── password.go
│   │       │   ├── role.go
│   │       │   ├── user.go
│   │       │   └── user_repository.go
│   │       ├── repository
│   │       │   ├── interface.go
│   │       │   └── postgres
│   │       │       ├── dbmodel
│   │       │       │   ├── membership.go
│   │       │       │   └── user.go
│   │       │       ├── mapper
│   │       │       │   └── user_mapper.go
│   │       │       └── user_repository.go
│   │       └── storage
│   │           └── interface.go
│   └── shared
│       ├── contracts
│       │   └── persistence
│       │       ├── dbmodel
│       │       │   └── base_model.go
│       │       └── gormblame
│       │           └── gormblame.go
│       ├── errors
│       │   └── errors.go
│       ├── pagination
│       │   └── pagination.go
│       ├── searchkit
│       │   ├── normalize.go
│       │   ├── op_parse.go
│       │   ├── payload.go
│       │   ├── spec.go
│       │   └── validate.go
│       └── strcase
│           └── camel_to_case.go
└── README.md

```

# Content:

## README.md

```md
# collabsphere-go

Запуск для windows:
# Создание сети для внешней работы
- docker network create web.network || true
# Запуска server api + local бд postgres
- docker compose -p collabsphere -f docker-compose.infrastructure.yaml -f docker-compose.platform.yaml --profile local up -d --build --force-recreate
# Миграция данных 
- docker compose -p collabsphere-migrate -f docker-compose.migrate.yaml up --abort-on-container-exit --exit-code-from migrate migrate
# Получить дерево файлов + содержимое 
codeweaver -input=. -output=backend-context.md -include='\.go$,\.md$,\.sql$,\.yaml$' -clipboard
```


## deploy/docker-compose.infrastructure.yaml

```yaml
name: collabsphere

services:
  postgres:
    profiles: [ "local" ]
    image: postgres:18.2-bookworm
    restart: unless-stopped
    hostname: "${POSTGRES_HOST}"
    networks:
      - internal.network
    env_file:
      - .env
    secrets:
      - dev_password
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d collabsphere"]
      interval: 3s
      timeout: 3s
      retries: 20
    ports:
      - "${POSTGRES_PORT}:${POSTGRES_PORT}"
    volumes:
      - postgres.data:/var/lib/postgresql


#volumes
volumes:
  postgres.data:
    name: "postgres.data"

#network settings
networks:
  internal.network:
    name: internal.network
    driver: bridge
    external: false
  web.network:
    name: web.network
    external: true

#secret keys
secrets:
  dev_password:
    file: secrets/postgres/dev/db_password

```


## deploy/docker-compose.migrate.yaml

```yaml
name: collabsphere

services:
  migrate:
    profiles: [ "migrate" ]
    image: colabsphere-api:${IMAGE_TAG}
    env_file:
      - .env
    volumes:
      - ../platform/internal/platform/db/migrations:/app/migrations:ro
    environment:
      # принимает up, status, down
      MIGRATE_CMD: up
    secrets:
      - dev_password
      - neon_password
    networks:
      - internal.network
    command: ["/migrate"]
    restart: "no"

networks:
  internal.network:
    name: internal.network
    driver: bridge

secrets:
  dev_password:
    file: secrets/postgres/dev/db_password
  neon_password:
    file: secrets/postgres/prod/db_password
```


## deploy/docker-compose.platform.yaml

```yaml
name: collabsphere

services:
  api:
    profiles: [ "local", "prod" ]
    image: colabsphere-api:${IMAGE_TAG}
    restart: unless-stopped
    build:
      context: ../platform                  # parent directory (collabsphere-go)
      dockerfile: ../platform/Dockerfile    # path to your Dockerfile relative to context
    env_file:
      - .env
    secrets:
      - dev_password
      - neon_password
    networks:
      - internal.network
      - web.network
    ports:
      - "${APPLICATION_PORT}:${APPLICATION_PORT}"
    healthcheck:
      test: ["CMD-SHELL", "curl -fsSL http://localhost:8080/docs"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
    command: ["/api"]
    depends_on:
      postgres:
        condition: service_healthy

networks:
  internal.network:
    name: internal.network
    driver: bridge
    external: false
  web.network:
    name: web.network
    external: true

secrets:
  dev_password:
    file: secrets/postgres/dev/db_password
  neon_password:
    file: secrets/postgres/prod/db_password
```


## deploy/docker-compose.proxy.yaml

```yaml
services:
  proxy-caddy:
    profiles: ["local", "cloud"]
    image: proxy-caddy:latest
    restart: unless-stopped
    build:
      context: ..//proxy/caddy
    env_file:
      - .env
    ports:
      - "80:80"
      - "443:443/tcp"
      - "443:443/udp"
    networks:
      - web.network
      - internal.network
    volumes:
      - ../proxy/caddy/Caddyfile.local:/etc/caddy/Caddyfile.local:ro
      - ../proxy/caddy/Caddyfile.prod:/etc/caddy/Caddyfile.prod:ro
      - ../proxy/caddy/ssl:/etc/caddy/ssl:ro

networks:
  internal.network:
    name: internal.network
    driver: bridge
    external: false
  web.network:
    name: web.network
    external: true
```


## llm/codebase.md

````md
# Tree View:
```
./platform

```

# Content:

````


## platform/backend-context.md

````md
# Tree View:
```
.
└── internal
    └── platform
        └── db
            └── migrations
                └── 0001_init.sql

```

# Content:

## internal/platform/db/migrations/0001_init.sql

```sql
-- +goose Up
-- Включаем uuid генератор
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Таблица users
CREATE TABLE IF NOT EXISTS users (
                                     id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

                                     email varchar NOT NULL,
                                     password_hash varchar NOT NULL,

                                     first_name varchar NOT NULL,
                                     last_name  varchar NOT NULL,
                                     phone      varchar,

                                     role varchar(50) NOT NULL,
                                     is_active boolean NOT NULL DEFAULT true,

                                     created_at timestamptz NOT NULL DEFAULT now(),
                                     updated_at timestamptz NOT NULL DEFAULT now(),
                                     created_by varchar NULL,
                                     updated_by varchar NULL
);

-- Уникальность email (глобально).
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users (email);

-- Индексы для аудита
CREATE INDEX IF NOT EXISTS ix_users_created_by ON users (created_by);
CREATE INDEX IF NOT EXISTS ix_users_updated_by ON users (updated_by);

-- Часто полезно:
CREATE INDEX IF NOT EXISTS ix_users_created_at ON users (created_at);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
CREATE TRIGGER trg_users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


-- +goose Down
DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS users;
```


````


## platform/cmd/api/main.go

```go
package main

import (
	"log"

	"github.com/NikolayNam/collabsphere-go/cmd/app"
	"github.com/NikolayNam/collabsphere-go/cmd/httpserver"
	"github.com/NikolayNam/collabsphere-go/internal/platform/config"
)

func main() {
	// 1) config (env + secrets + TZ)
	conf := config.New()

	// 2) build app (router + huma + module registration)
	application := app.New(conf)

	// 3) run http server (timeouts + graceful shutdown)
	if err := httpserver.Run(application.Router, conf.APP.Address,
		httpserver.Options{
			ReadTimeout:       conf.APP.TimeoutRead,
			WriteTimeout:      conf.APP.TimeoutWrite,
			IdleTimeout:       conf.APP.TimeoutIdle,
			ReadHeaderTimeout: 5, // seconds
			ShutdownTimeout:   5, // seconds
		}); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

```


## platform/cmd/app/api.go

```go
package app

import (
	"github.com/NikolayNam/collabsphere-go/internal/platform/config"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func newAPI(router chi.Router, conf *config.Config) huma.API {
	cfg := huma.DefaultConfig(conf.APP.Title, conf.APP.Version)
	cfg.CreateHooks = nil

	return humachi.New(router, cfg)
}

```


## platform/cmd/app/app.go

```go
package app

import (
	"github.com/NikolayNam/collabsphere-go/internal/platform/config"
	appLogger "github.com/NikolayNam/collabsphere-go/internal/platform/logger"
	"gorm.io/gorm"

	systemapp "github.com/NikolayNam/collabsphere-go/internal/system"
	usersapp "github.com/NikolayNam/collabsphere-go/internal/users/application"
	usershttp "github.com/NikolayNam/collabsphere-go/internal/users/delivery/http"
	usersrepo "github.com/NikolayNam/collabsphere-go/internal/users/repository/postgres"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
)

type App struct {
	Router chi.Router
	API    huma.API
}

func New(conf *config.Config) *App {
	log := appLogger.New()

	router := newRouter(log)
	api := newAPI(router, conf)

	db := mustOpenGormDB(conf, log)
	registerDBHooks(db)

	registerPlatform(api)
	registerUsersModule(api, db)

	return &App{
		Router: router,
		API:    api,
	}
}

func registerPlatform(api huma.API) {
	systemapp.Register(api)
}

func registerUsersModule(api huma.API, db *gorm.DB) {
	userRepo := usersrepo.NewUserRepo(db)
	userApp := usersapp.New(userRepo)
	userHandler := usershttp.NewHandler(userApp)

	usershttp.Register(api, userHandler)
}

```


## platform/cmd/app/db.go

```go
package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/NikolayNam/collabsphere-go/internal/platform/config"
	"github.com/NikolayNam/collabsphere-go/shared/contracts/persistence/gormblame"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func registerDBHooks(db *gorm.DB) {
	if err := gormblame.Register(db); err != nil {
		panic(err)
	}
}

func mustOpenGormDB(conf *config.Config, log *slog.Logger) *gorm.DB {
	dsn, err := conf.DB.DSN()
	if err != nil {
		panic(err)
	}

	log.Info(
		"db connecting",
		"dsn", sanitizeDSN(dsn),
		"host", conf.DB.Host,
		"port", conf.DB.Port,
		"db", conf.DB.DBName,
		"user", conf.DB.Username,
		"schema", conf.DB.DBSchema,
	)

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newGormLogger(log, conf.DB.Debug),
	})
	if err != nil {
		log.Error("db gorm open failed", "err", err, "dsn", sanitizeDSN(dsn))
		panic(err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		panic(err)
	}

	configureSQLPool(sqlDB)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Error("db ping failed", "err", err, "dsn", sanitizeDSN(dsn))
		panic(err)
	}

	log.Info("db connected")
	return gdb
}

func configureSQLPool(sqlDB *sql.DB) {
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
}

func sanitizeDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return `<invalid dsn>`
	}
	if u.User != nil {
		u.User = url.User(u.User.Username())
	}
	return u.String()
}

func newGormLogger(log *slog.Logger, debug bool) gormlogger.Interface {
	level := gormlogger.Error
	if debug {
		level = gormlogger.Info
	}

	return gormlogger.New(
		slogWriter{log: log},
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

type slogWriter struct {
	log *slog.Logger
}

func (w slogWriter) Printf(format string, args ...any) {
	w.log.Info("gorm", "msg", fmt.Sprintf(format, args...))
}

```


## platform/cmd/app/router.go

```go
package app

import (
	"log/slog"

	"github.com/NikolayNam/collabsphere-go/internal/platform/middleware"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func newRouter(log *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestSize(1 << 20))

	r.Use(middleware.LoggerContext(log))
	r.Use(middleware.AccessLog())

	return r
}

```


## platform/cmd/httpserver/server.go

```go
package httpserver

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Options struct {
	ReadHeaderTimeout int // seconds
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   int // seconds

	// optional:
	Logger *slog.Logger
}

func Run(handler http.Handler, addr string, opt Options) error {
	logger := opt.Logger
	if logger == nil {
		// прод: JSON; локально можно TextHandler
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(opt.ReadHeaderTimeout) * time.Second,
		ReadTimeout:       opt.ReadTimeout,
		WriteTimeout:      opt.WriteTimeout,
		IdleTimeout:       opt.IdleTimeout,

		// важно: ошибки самого сервера в тот же логгер
		ErrorLog: slogToStdLogger(logger, slog.LevelError),
	}

	serverErr := make(chan error, 1)

	go func() {
		logger.Info("http server starting",
			"addr", addr,
			"read_header_timeout_s", opt.ReadHeaderTimeout,
			"read_timeout", opt.ReadTimeout.String(),
			"write_timeout", opt.WriteTimeout.String(),
			"idle_timeout", opt.IdleTimeout.String(),
		)

		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case sig := <-quit:
		logger.Warn("shutdown signal received", "signal", sig.String())
	case err := <-serverErr:
		if err != nil {
			logger.Error("http server failed", "error", err.Error())
		}
		return err
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(opt.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	logger.Info("http server shutting down", "timeout_s", opt.ShutdownTimeout)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("http server shutdown failed",
			"error", err.Error(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}

	logger.Info("http server stopped",
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return nil
}

// slogToStdLogger адаптирует slog -> *log.Logger для http.Server.ErrorLog
func slogToStdLogger(l *slog.Logger, level slog.Level) *log.Logger {
	return log.New(&slogWriter{l: l, level: level}, "", 0)
}

type slogWriter struct {
	l     *slog.Logger
	level slog.Level
}

func (w *slogWriter) Write(p []byte) (n int, err error) {
	// p обычно уже с \n
	msg := string(p)
	w.l.Log(context.Background(), w.level, "http server error", "message", msg)
	return len(p), nil
}

```


## platform/cmd/migrate/main.go

```go
package main

import (
	"database/sql"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/NikolayNam/collabsphere-go/internal/platform/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	conf := config.New()

	dsn, err := conf.DB.DSN()
	if err != nil {
		log.Fatal(err)
	}

	dir := envOrDefault("MIGRATIONS_DIR", "/app/migrations")
	cmd := strings.ToLower(envOrDefault("MIGRATE_CMD", "up"))

	log.Printf("migrate: connecting dsn=%s schema=%s dir=%s cmd=%s",
		sanitizeDSN(dsn), conf.DB.DBSchema, dir, cmd,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("migrate: close db: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	switch cmd {
	case "up":
		err = goose.Up(db, dir)
	case "down":
		err = goose.Down(db, dir)
	case "status":
		err = goose.Status(db, dir)
	case "version":
		var v int64
		v, err = goose.GetDBVersion(db)
		if err == nil {
			log.Printf("migrate: current version=%d", v)
		}
	default:
		log.Fatalf("migrate: unsupported MIGRATE_CMD=%q", cmd)
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("migrate: done")
}

func envOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func sanitizeDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return "<invalid dsn>"
	}
	if u.User != nil {
		u.User = url.User(u.User.Username())
	}
	return u.String()
}

```


## platform/internal/organizations/domain/exceptions.go

```go
package domain

import "errors"

const (
	CodeInvalidEmail         ErrorCode = "INVALID_EMAIL"
	CodeInvalidLegalName     ErrorCode = "INVALID_LEGAL_NAME"
	CodeAlreadyVerified      ErrorCode = "ALREADY_VERIFIED"
	CodeAlreadyDeactivated   ErrorCode = "ALREADY_DEACTIVATED"
	CodeCreatorLoginRequired ErrorCode = "CREATOR_LOGIN_REQUIRED"
	CodeDomain               ErrorCode = "DOMAIN_ERROR"
)

type (
	// ErrorCode — стабильные коды доменных ошибок.
	ErrorCode string
)

// Error — единый тип доменной ошибки.
type Error struct {
	Code    ErrorCode
	Message string
}

func (e Error) Error() string {
	return e.Message
}

// Is позволяет использовать errors.Is(...)
func (e Error) Is(target error) bool {
	var t Error
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func NewInvalidEmail() error {
	return Error{
		Code:    CodeInvalidEmail,
		Message: "Неверный формат email",
	}
}

func NewInvalidLegalName() error {
	return Error{
		Code:    CodeInvalidLegalName,
		Message: "Некорректное юридическое наименование",
	}
}

func NewAlreadyVerified() error {
	return Error{
		Code:    CodeAlreadyVerified,
		Message: "Компания уже верифицирована",
	}
}

func NewAlreadyDeactivated() error {
	return Error{
		Code:    CodeAlreadyDeactivated,
		Message: "Компания деактивирована",
	}
}

func NewCreatorLoginRequired() error {
	return Error{
		Code:    CodeCreatorLoginRequired,
		Message: "Логин создателя обязателен",
	}
}

func New(msg string) error {
	return Error{
		Code:    CodeDomain,
		Message: msg,
	}
}

```


## platform/internal/organizations/domain/organizations.go

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Organization — доменная сущность.
type Organization struct {
	UID uuid.UUID

	LegalName   string
	DisplayName *string

	// Регистрация/юрисдикция (1:1)
	CountryOfRegistration string // ISO2: "DE","US","RU"
	LegalEntityTypeID     *int
	TypeID                *int

	PrimaryEmail   string
	PrimaryAddress *string
	PrimaryPhone   *string
	PrimarySite    *string

	// Статус/верификация
	Status     string // Draft/Active/Suspended/Archived
	VerifiedAt *time.Time
	VerifiedBy *string

	// Аудит
	CreatedBy string
	CreatedAt time.Time

	UpdatedBy *string
	UpdatedAt *time.Time

	DeletedAt *time.Time

	Version int
}

// NewOrganization — фабрика: гарантирует корректные дефолты (UID, CreatedAt).
// Валидации минимальные; усиливай по необходимости.
func NewOrganization(
	legalName string,
	primaryEmail string,
	createdBy string,
) (*Organization, error) {
	if legalName == "" {
		return nil, NewInvalidLegalName()
	}
	if primaryEmail == "" {
		return nil, NewInvalidEmail() // <-- вместо “просто ошибки”
	}
	if createdBy == "" {
		return nil, NewCreatorLoginRequired()
	}

	now := time.Now().UTC()

	return &Organization{
		UID:          uuid.New(),
		LegalName:    legalName,
		PrimaryEmail: primaryEmail,
		CreatedBy:    createdBy,
		CreatedAt:    now,
	}, nil
}

func (o *Organization) ChangeLegalName(newLegalName, user string) error {
	if newLegalName == "" {
		return NewInvalidLegalName() // <-- доменная ошибка
	}
	o.LegalName = newLegalName
	o.touch(user)
	return nil
}

func (o *Organization) touch(user string) {
	o.UpdatedBy = new(user)
	o.UpdatedAt = new(time.Now().UTC())
}

```


## platform/internal/platform/actorctx/actorctx.go

```go
package actorctx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

var key ctxKey

func WithActorID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, key, id)
}

func ActorID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(key)
	if v == nil {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func MustActorID(ctx context.Context) uuid.UUID {
	id, ok := ActorID(ctx)
	if !ok {
		panic("actor id not found in context")
	}
	return id
}

```


## platform/internal/platform/config/config.go

```go
package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	TZ  string `env:"TZ" envDefault:"UTC"`
	APP App
	DB  DB
}

type App struct {
	Title        string        `env:"APPLICATION_TITLE,required"`
	Version      string        `env:"APPLICATION_VERSION,required"`
	Address      string        `env:"APPLICATION_ADDRESS" envDefault:"0.0.0.0:8080"`
	TimeoutRead  time.Duration `env:"APPLICATION_TIMEOUT_READ" envDefault:"15s"`
	TimeoutWrite time.Duration `env:"APPLICATION_TIMEOUT_WRITE" envDefault:"15s"`
	TimeoutIdle  time.Duration `env:"APPLICATION_TIMEOUT_IDLE" envDefault:"60s"`
	Debug        bool          `env:"APPLICATION_DEBUG" envDefault:"false"`
}

type DB struct {
	Host string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port int    `env:"POSTGRES_PORT" envDefault:"5432"`

	DBName   string `env:"POSTGRES_DB" envDefault:"postgres"`
	DBSchema string `env:"POSTGRES_SCHEMA,required"`

	Username string `env:"POSTGRES_USER" envDefault:"postgres"`

	Password     string `env:"POSTGRES_PASSWORD"`
	PasswordFile string `env:"POSTGRES_PASSWORD_FILE"`

	Debug bool `env:"POSTGRES_DEBUG" envDefault:"false"`
}

func New() *Config {
	var c Config

	if err := env.Parse(&c); err != nil {
		log.Fatalf("failed to parse env: %s", err)
	}

	if err := applyTZ(c.TZ); err != nil {
		log.Fatalf("invalid TZ: %s", err)
	}

	return &c
}

func (d DB) PasswordValue() (string, error) {
	if strings.TrimSpace(d.Password) != "" {
		return d.Password, nil
	}
	if strings.TrimSpace(d.PasswordFile) == "" {
		return "", errors.New("postgres password is empty (set POSTGRES_PASSWORD or POSTGRES_PASSWORD_FILE)")
	}

	b, err := os.ReadFile(d.PasswordFile)
	if err != nil {
		return "", err
	}
	pw := strings.TrimSpace(string(b))
	if pw == "" {
		return "", errors.New("postgres password file is empty")
	}
	return pw, nil
}

// --- Secret Loader ---

func applyTZ(tz string) error {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
}

// --- DSN Builder (не логируй его целиком) ---

func (d DB) DSN() (string, error) {
	pw, err := d.PasswordValue()
	if err != nil {
		return "", err
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.Username, pw),
		Host:   fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:   d.DBName,
	}

	q := u.Query()
	q.Set("sslmode", "disable")
	// schema first, then public fallback
	if strings.TrimSpace(d.DBSchema) != "" {
		q.Set("search_path", fmt.Sprintf("%s,public", d.DBSchema))
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

```


## platform/internal/platform/db/migrations/0001_init.sql

```sql
-- +goose Up
-- Включаем uuid генератор
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Таблица users
CREATE TABLE IF NOT EXISTS users (
                                     id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

                                     email varchar NOT NULL,
                                     password_hash varchar NOT NULL,

                                     first_name varchar NOT NULL,
                                     last_name  varchar NOT NULL,
                                     phone      varchar,

                                     role varchar(50) NOT NULL,
                                     is_active boolean NOT NULL DEFAULT true,

                                     created_at timestamptz NOT NULL DEFAULT now(),
                                     updated_at timestamptz NOT NULL DEFAULT now(),
                                     created_by varchar NULL,
                                     updated_by varchar NULL
);

-- Уникальность email (глобально).
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users (email);

-- Индексы для аудита
CREATE INDEX IF NOT EXISTS ix_users_created_by ON users (created_by);
CREATE INDEX IF NOT EXISTS ix_users_updated_by ON users (updated_by);

-- Часто полезно:
CREATE INDEX IF NOT EXISTS ix_users_created_at ON users (created_at);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
CREATE TRIGGER trg_users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();


-- +goose Down
DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS users;
```


## platform/internal/platform/logger/context.go

```go
package logger

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

func With(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func From(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(ctxKey{}).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return l
}

```


## platform/internal/platform/logger/logger.go

```go
package logger

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
}

```


## platform/internal/platform/middleware/access_log.go

```go
package middleware

import (
	"net/http"
	"time"

	"log/slog"

	appLogger "github.com/NikolayNam/collabsphere-go/internal/platform/logger"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func AccessLog() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w}
			start := time.Now()

			next.ServeHTTP(sw, r)

			dur := time.Since(start)

			l := appLogger.From(r.Context())

			level := slog.LevelInfo
			if sw.status >= 500 {
				level = slog.LevelError
			} else if sw.status >= 400 {
				level = slog.LevelWarn
			}

			reqID := chimw.GetReqID(r.Context())

			l.Log(r.Context(), level, "http request",
				"request_id", reqID,
				"status", sw.status,
				"bytes", sw.bytes,
				"duration_ms", dur.Milliseconds(),
				"remote_ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

```


## platform/internal/platform/middleware/logger_context.go

```go
package middleware

import (
	"net/http"

	"log/slog"

	appLogger "github.com/NikolayNam/collabsphere-go/internal/platform/logger"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func LoggerContext(base *slog.Logger) func(http.Handler) http.Handler {
	if base == nil {
		panic("base logger is nil")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := chimw.GetReqID(r.Context())

			l := base.With(
				"request_id", reqID,
				"method", r.Method,
				"path", r.URL.Path,
			)

			ctx := appLogger.With(r.Context(), l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

```


## platform/internal/platform/middleware/middleware.go

```go
package middleware

import (
	_ "context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
)

type ctxKey string

const OrganizationIDKey ctxKey = "organization_id"

var limiter = rate.NewLimiter(5, 10) // 5 rps, burst 10

func Stack(next http.Handler) http.Handler {
	return chi.Chain(
		chimw.RequestID,
		chimw.RealIP,
		chimw.Recoverer,
		chimw.Timeout(30*time.Second),
		SecurityHeaders,
		OrgContext,
		RateLimit,
		chimw.RequestSize(1<<20), // 1MB
	).Handler(next)
}

```


## platform/internal/platform/middleware/org.go

```go
package middleware

import (
	"context"
	"net/http"
)

func OrgContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		orgID := r.Header.Get("X-Organization-ID")
		if orgID != "" {
			ctx := context.WithValue(r.Context(), OrganizationIDKey, orgID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

```


## platform/internal/platform/middleware/ratelimit.go

```go
package middleware

import "net/http"

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

```


## platform/internal/platform/middleware/security.go

```go
package middleware

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}

```


## platform/internal/system/dto.go

```go
package system

type HealthOutput struct {
	Body struct {
		Status string `json:"status" example:"ok"`
	}
}

type ReadinessOutput struct {
	Body struct {
		Status string `json:"status" example:"ready"`
	}
}

```


## platform/internal/system/handlers.go

```go
package system

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API) {
	huma.Get(api, "/health", healthHandler)
	huma.Get(api, "/ready", readyHandler)
}

func healthHandler(ctx context.Context, input *struct{}) (*HealthOutput, error) {
	resp := &HealthOutput{}
	resp.Body.Status = "ok"
	return resp, nil
}

func readyHandler(ctx context.Context, input *struct{}) (*ReadinessOutput, error) {
	resp := &ReadinessOutput{}
	resp.Body.Status = "ready"
	return resp, nil
}

```


## platform/internal/users/application/errors.go

```go
package application

import "errors"

var (
	ErrValidation = errors.New("validation")
	ErrConflict   = errors.New("conflict")
	ErrInternal   = errors.New("internal")
)

```


## platform/internal/users/application/service.go

```go
package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/NikolayNam/collabsphere-go/internal/users/domain"
)

type Repository interface {
	Create(ctx context.Context, u *domain.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateUserCmd struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
	Role      string
}

func (s *Service) CreateUser(ctx context.Context, cmd CreateUserCmd) (*domain.User, error) {
	// 1) normalize/validate email
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return nil, ErrValidation
	}

	// 2) role
	roleRaw := strings.TrimSpace(cmd.Role)
	if roleRaw == "" {
		roleRaw = "user"
	}
	role, err := domain.NewRole(roleRaw)
	if err != nil {
		return nil, ErrValidation
	}

	if len(strings.TrimSpace(cmd.Password)) == 0 {
		return nil, ErrValidation
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrInternal
	}
	ph, err := domain.NewPasswordHash(string(hash))
	if err != nil {
		return nil, ErrInternal
	}

	// 4) optional: уникальность email (если repo ожидает org)
	exists, err := s.repo.ExistsByEmail(ctx, email.String())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}
	if exists {
		return nil, ErrConflict
	}

	// 5) create domain entity
	u, err := domain.NewUser(domain.NewUserParams{
		ID:           domain.NewUserID(),
		Email:        email,
		PasswordHash: ph,
		FirstName:    cmd.FirstName,
		LastName:     cmd.LastName,
		Phone:        cmd.Phone,
		Role:         role,
		Now:          time.Now(),
	})
	if err != nil {
		return nil, ErrValidation
	}

	// 6) persist
	if err := s.repo.Create(ctx, u); err != nil {
		// если ты маппишь уникальность в repo и он возвращает ErrConflict — ок
		if errors.Is(err, ErrConflict) {
			return nil, ErrConflict
		}
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return u, nil
}

```


## platform/internal/users/cache/interface.go

```go
package cache

```


## platform/internal/users/cache/redis/keys.go

```go
package redis

```


## platform/internal/users/cache/redis/user_cache.go

```go
package redis

```


## platform/internal/users/delivery/http/dto.go

```go
package http

// CreateUserInput POST /users (в рамках X-Organization-ID)
type CreateUserInput struct {
	Body struct {
		Email     string `json:"email" required:"true" format:"email"`
		Password  string `json:"password" required:"true" minLength:"6"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Phone     string `json:"phone,omitempty"`
		Role      string `json:"role,omitempty"`
	}
}

// UpdateUserInput PUT /users/{user_id} (в рамках X-Organization-ID)
type UpdateUserInput struct {
	UserID uint `path:"user_id" doc:"User ID"`

	Body struct {
		FirstName *string `json:"first_name,omitempty"`
		LastName  *string `json:"last_name,omitempty"`
		Phone     *string `json:"phone,omitempty"`
		Role      *string `json:"role,omitempty" doc:"Update role in this organization"`
		IsActive  *bool   `json:"is_active,omitempty" doc:"Deactivate membership in this organization"`
	}
}

type UserResponse struct {
	Body struct {
		ID        uint   `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`

		OrganizationID string `json:"organization_id"`
		Role           string `json:"role"`
		IsActive       bool   `json:"is_active"`
	}
}

```


## platform/internal/users/delivery/http/handler.go

```go
package http

import (
	"context"
	"errors"
	"log/slog"

	appLogger "github.com/NikolayNam/collabsphere-go/internal/platform/logger"
	"github.com/danielgtaylor/huma/v2"

	userapp "github.com/NikolayNam/collabsphere-go/internal/users/application"
)

type Handler struct {
	svc *userapp.Service
}

func NewHandler(svc *userapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateUser(ctx context.Context, input *CreateUserInput) (*UserResponse, error) {
	u, err := h.svc.CreateUser(ctx, userapp.CreateUserCmd{
		Email:     input.Body.Email,
		Password:  input.Body.Password,
		FirstName: input.Body.FirstName,
		LastName:  input.Body.LastName,
		Phone:     input.Body.Phone,
		Role:      input.Body.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, userapp.ErrValidation):
			return nil, huma.Error400BadRequest("Invalid input")
		case errors.Is(err, userapp.ErrConflict):
			return nil, huma.Error409Conflict("User already exists")
		default:
			l := appLogger.From(ctx)
			if l == nil {
				l = slog.Default()
			}
			l.Error(
				"create user failed",
				"err", err, // ← вот это тебе нужно
				"email", input.Body.Email,
			)
			return nil, huma.Error500InternalServerError("Internal error")
		}
	}

	resp := &UserResponse{}
	resp.Body.Email = u.Email().String()
	resp.Body.FirstName = u.FirstName()
	resp.Body.LastName = u.LastName()
	resp.Body.Phone = u.Phone()
	resp.Body.Role = string(u.Role())
	resp.Body.IsActive = u.IsActive()

	return resp, nil
}

```


## platform/internal/users/delivery/http/operations.go

```go
package http

import "github.com/danielgtaylor/huma/v2"

var createUserOp = huma.Operation{
	OperationID: "create-user",
	Method:      "POST",
	Path:        "/users",
	Tags:        []string{"Users"},
	Summary:     "Create user",
}

```


## platform/internal/users/delivery/http/register.go

```go
package http

import (
	"github.com/danielgtaylor/huma/v2"
)

func Register(api huma.API, h *Handler) {
	huma.Register(api, createUserOp, h.CreateUser)
}

```


## platform/internal/users/domain/email.go

```go
package domain

import (
	"errors"
	"regexp"
	"strings"
)

type Email string

func NewEmail(raw string) (Email, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return "", errors.New("email is empty")
	}

	// Практичная проверка (не RFC-идеал, но защищает от мусора).
	re := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !re.MatchString(s) {
		return "", errors.New("email is invalid")
	}

	return Email(s), nil
}

func (e Email) String() string { return string(e) }
func (e Email) IsZero() bool   { return strings.TrimSpace(string(e)) == "" }

```


## platform/internal/users/domain/ids.go

```go
package domain

import (
	"errors"

	"github.com/google/uuid"
)

type UserID uuid.UUID

func NewUserID() UserID {
	return UserID(uuid.New())
}

func UserIDFromUUID(id uuid.UUID) (UserID, error) {
	if id == uuid.Nil {
		return UserID{}, errors.New("user id is nil")
	}
	return UserID(id), nil
}

func (id UserID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id UserID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

```


## platform/internal/users/domain/password.go

```go
package domain

import "errors"

// PasswordHash — доменный тип (не “просто строка”).
// В домене мы не обязаны знать алгоритм (bcrypt/argon2) — это outside world.
type PasswordHash string

func NewPasswordHash(hash string) (PasswordHash, error) {
	if hash == "" {
		return "", errors.New("password hash is empty")
	}
	return PasswordHash(hash), nil
}

func (h PasswordHash) String() string { return string(h) }
func (h PasswordHash) IsZero() bool   { return string(h) == "" }

```


## platform/internal/users/domain/role.go

```go
package domain

import "errors"

type Role string

const (
	RoleOwner Role = "owner"
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) Valid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleUser:
		return true
	default:
		return false
	}
}

func NewRole(raw string) (Role, error) {
	r := Role(raw)
	if !r.Valid() {
		return "", errors.New("role is invalid")
	}
	return r, nil
}

```


## platform/internal/users/domain/user.go

```go
package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrUserInactive     = errors.New("user is inactive")
	ErrInvalidFirstName = errors.New("first name is invalid")
	ErrInvalidLastName  = errors.New("last name is invalid")
	ErrInvalidPhone     = errors.New("phone is invalid")
	ErrCannotDeactivate = errors.New("cannot deactivate user")
	ErrCannotActivate   = errors.New("cannot activate user")
)

type User struct {
	// identity
	id    UserID
	email Email

	// credentials (секрет)
	passwordHash PasswordHash

	// profile
	firstName string
	lastName  string
	phone     string

	// access
	role     Role
	isActive bool

	// Domain time (только если нужно домену; иначе убрать)
	createdAt time.Time
	updatedAt time.Time
}

type NewUserParams struct {
	ID           UserID
	Email        Email
	PasswordHash PasswordHash
	FirstName    string
	LastName     string
	Phone        string
	Role         Role
	Now          time.Time
}

func NewUser(p NewUserParams) (*User, error) {
	if err := validateUserCore(p.ID, p.Email, p.PasswordHash, p.Role); err != nil {
		return nil, err
	}
	if p.Now.IsZero() {
		return nil, errors.New("now is required")
	}

	fn, ln, ph, err := normalizeProfile(p.FirstName, p.LastName, p.Phone)
	if err != nil {
		return nil, err
	}

	return &User{
		id:           p.ID,
		email:        p.Email,
		passwordHash: p.PasswordHash,
		firstName:    fn,
		lastName:     ln,
		phone:        ph,
		role:         p.Role,
		isActive:     true,
		createdAt:    p.Now,
		updatedAt:    p.Now,
	}, nil
}

// RehydrateUserParams — восстановление из persistence (репозиторий).
type RehydrateUserParams struct {
	ID           UserID
	Email        Email
	PasswordHash PasswordHash
	FirstName    string
	LastName     string
	Phone        string
	Role         Role
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func RehydrateUser(p RehydrateUserParams) (*User, error) {
	if err := validateUserCore(p.ID, p.Email, p.PasswordHash, p.Role); err != nil {
		return nil, err
	}
	if p.CreatedAt.IsZero() || p.UpdatedAt.IsZero() {
		return nil, errors.New("timestamps are required")
	}

	fn, ln, ph, err := normalizeProfile(p.FirstName, p.LastName, p.Phone)
	if err != nil {
		return nil, err
	}

	return &User{
		id:           p.ID,
		email:        p.Email,
		passwordHash: p.PasswordHash,
		firstName:    fn,
		lastName:     ln,
		phone:        ph,
		role:         p.Role,
		isActive:     p.IsActive,
		createdAt:    p.CreatedAt,
		updatedAt:    p.UpdatedAt,
	}, nil
}

/*
	Read-only API (наружу — только безопасное)
*/

func (u *User) ID() UserID     { return u.id }
func (u *User) Email() Email   { return u.email }
func (u *User) Role() Role     { return u.role }
func (u *User) IsActive() bool { return u.isActive }

func (u *User) FirstName() string { return u.firstName }
func (u *User) LastName() string  { return u.lastName }
func (u *User) Phone() string     { return u.phone }

func (u *User) PasswordHash() PasswordHash { return u.passwordHash }

func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

/*
	Domain behavior
*/

func (u *User) Activate(now time.Time) error {
	if now.IsZero() {
		return errors.New("now is required")
	}
	if u.isActive {
		return ErrCannotActivate
	}
	u.isActive = true
	u.updatedAt = now
	return nil
}

func (u *User) Deactivate(now time.Time) error {
	if now.IsZero() {
		return errors.New("now is required")
	}
	if !u.isActive {
		return ErrCannotDeactivate
	}
	u.isActive = false
	u.updatedAt = now
	return nil
}

func (u *User) ChangeRole(role Role, now time.Time) error {
	if now.IsZero() {
		return errors.New("now is required")
	}
	if !role.Valid() {
		return errors.New("invalid role")
	}
	if !u.isActive {
		return ErrUserInactive
	}

	u.role = role
	u.updatedAt = now
	return nil
}

func (u *User) UpdateProfile(firstName, lastName, phone string, now time.Time) error {
	if now.IsZero() {
		return errors.New("now is required")
	}
	if !u.isActive {
		return ErrUserInactive
	}

	fn, ln, ph, err := normalizeProfile(firstName, lastName, phone)
	if err != nil {
		return err
	}

	u.firstName = fn
	u.lastName = ln
	u.phone = ph
	u.updatedAt = now
	return nil
}

func (u *User) SetPasswordHash(hash PasswordHash, now time.Time) error {
	if now.IsZero() {
		return errors.New("now is required")
	}
	if hash.IsZero() {
		return errors.New("password hash is empty")
	}
	if !u.isActive {
		return ErrUserInactive
	}

	u.passwordHash = hash
	u.updatedAt = now
	return nil
}

/*
	internal validation helpers
*/

func validateUserCore(id UserID, email Email, hash PasswordHash, role Role) error {
	if id.IsZero() {
		return errors.New("missing user id")
	}
	if email.IsZero() {
		return errors.New("missing email")
	}
	if hash.IsZero() {
		return errors.New("missing password hash")
	}
	if !role.Valid() {
		return errors.New("invalid role")
	}
	return nil
}

func normalizeProfile(firstName, lastName, phone string) (string, string, string, error) {
	fn, err := normalizeFirstName(firstName)
	if err != nil {
		return "", "", "", err
	}
	ln, err := normalizeLastName(lastName)
	if err != nil {
		return "", "", "", err
	}
	ph, err := normalizePhone(phone)
	if err != nil {
		return "", "", "", err
	}
	return fn, ln, ph, nil
}

func normalizeFirstName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 200 {
		return "", ErrInvalidFirstName
	}
	return s, nil
}

func normalizeLastName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 200 {
		return "", ErrInvalidLastName
	}
	return s, nil
}

func normalizePhone(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) > 50 {
		return "", ErrInvalidPhone
	}
	return s, nil
}

```


## platform/internal/users/domain/user_repository.go

```go
package domain

import "context"

type UserRepository interface {
	ByID(ctx context.Context, id UserID) (*User, error)
	ByEmail(ctx context.Context, email Email) (*User, error)

	Save(ctx context.Context, u *User) error
}

```


## platform/internal/users/repository/interface.go

```go
package repository

```


## platform/internal/users/repository/postgres/dbmodel/membership.go

```go
package dbmodel

```


## platform/internal/users/repository/postgres/dbmodel/user.go

```go
package dbmodel

import (
	basemodel "github.com/NikolayNam/collabsphere-go/shared/contracts/persistence/dbmodel"
)

type User struct {
	basemodel.BaseModel

	Email        string `gorm:"column:email;type:text;not null;uniqueIndex"`
	PasswordHash string `gorm:"column:password_hash;type:text;not null"`

	FirstName string `gorm:"column:first_name;type:text;not null;default:''"`
	LastName  string `gorm:"column:last_name;type:text;not null;default:''"`
	Phone     string `gorm:"column:phone;type:text;not null;default:''"`

	Role     string `gorm:"column:role;type:text;not null"`
	IsActive bool   `gorm:"column:is_active;not null;default:true"`
}

func (User) TableName() string { return "users" }

```


## platform/internal/users/repository/postgres/mapper/user_mapper.go

```go
package mapper

import (
	"time"

	"github.com/NikolayNam/collabsphere-go/internal/users/domain"
	basemodel "github.com/NikolayNam/collabsphere-go/internal/users/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere-go/shared/contracts/persistence/dbmodel"
)

func ToDomainUser(m *basemodel.User) (*domain.User, error) {
	if m == nil {
		return nil, nil
	}

	id, err := domain.UserIDFromUUID(m.ID)
	if err != nil {
		return nil, err
	}

	email, err := domain.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}

	role, err := domain.NewRole(m.Role)
	if err != nil {
		return nil, err
	}

	hash, err := domain.NewPasswordHash(m.PasswordHash)
	if err != nil {
		return nil, err
	}

	return domain.RehydrateUser(domain.RehydrateUserParams{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		Phone:        m.Phone,
		Role:         role,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	})
}

func ToDBUser(u *domain.User) (*basemodel.User, error) {
	if u == nil {
		return nil, nil
	}

	return &basemodel.User{
		BaseModel: dbmodel.BaseModel{
			UUIDPK: dbmodel.UUIDPK{ID: u.ID().UUID()},
			Timestamps: dbmodel.Timestamps{
				CreatedAt: nonZeroOrNow(u.CreatedAt(), time.Now()),
				UpdatedAt: nonZeroOrNow(u.UpdatedAt(), time.Now()),
			},
			// Blame: оставляем persistence callbacks
		},
		Email:        u.Email().String(),
		PasswordHash: u.PasswordHash().String(),
		FirstName:    u.FirstName(),
		LastName:     u.LastName(),
		Phone:        u.Phone(),
		Role:         string(u.Role()),
		IsActive:     u.IsActive(),
	}, nil
}

func nonZeroOrNow(t time.Time, now time.Time) time.Time {
	if t.IsZero() {
		return now
	}
	return t
}

```


## platform/internal/users/repository/postgres/user_repository.go

```go
package postgres

import (
	"context"

	"github.com/NikolayNam/collabsphere-go/internal/users/repository/postgres/dbmodel"
	"github.com/NikolayNam/collabsphere-go/internal/users/repository/postgres/mapper"
	"gorm.io/gorm"

	"github.com/NikolayNam/collabsphere-go/internal/users/domain"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) ByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	var m dbmodel.User
	err := r.db.WithContext(ctx).First(&m, "id = ?", id.UUID()).Error
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainUser(&m)
}

func (r *UserRepo) ByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	var m dbmodel.User
	err := r.db.WithContext(ctx).First(&m, "email = ?", email.String()).Error
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainUser(&m)
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	m, err := mapper.ToDBUser(u)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	// ВАЖНО: это корректно ТОЛЬКО если в таблице users реально есть organization_id.
	// Если у тебя роли/пользователи не привязаны к org, то интерфейс сервиса неправильный.
	var count int64
	err := r.db.WithContext(ctx).
		Model(&dbmodel.User{}).
		Where("email = ?", email).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

```


## platform/internal/users/storage/interface.go

```go
package storage

```


## platform/shared/contracts/persistence/dbmodel/base_model.go

```go
package dbmodel

import (
	"time"

	"github.com/google/uuid"
)

type UUIDPK struct {
	ID uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
}

type Timestamps struct {
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
}

type Blame struct {
	CreatedBy *uuid.UUID `gorm:"column:created_by;type:uuid;index"`
	UpdatedBy *uuid.UUID `gorm:"column:updated_by;type:uuid;index"`
}

type BaseModel struct {
	UUIDPK
	Timestamps
	Blame
}

```


## platform/shared/contracts/persistence/gormblame/gormblame.go

```go
package gormblame

import (
	"errors"
	"fmt"

	"github.com/NikolayNam/collabsphere-go/internal/platform/actorctx"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	cbBeforeCreate = "gormblame:before_create"
	cbBeforeUpdate = "gormblame:before_update"
)

// Register registers GORM callbacks that set created_by / updated_by from context actor ID.
func Register(db *gorm.DB) error {
	if db == nil {
		return errors.New("gormblame: db is nil")
	}

	// Avoid double registration (panic in gorm if same name registered twice).
	// gorm doesn't expose "exists" directly; common pragmatic approach is:
	// - call Replace instead of Register
	_ = db.Callback().Create().Before("gorm:create").Replace(cbBeforeCreate, beforeCreate)
	_ = db.Callback().Update().Before("gorm:update").Replace(cbBeforeUpdate, beforeUpdate)

	return nil
}

func beforeCreate(tx *gorm.DB) {
	apply(tx, true, true) // on create: set created_by (if nil) and updated_by
}

func beforeUpdate(tx *gorm.DB) {
	apply(tx, false, true) // on update: set updated_by only
}

func apply(tx *gorm.DB, setCreated bool, setUpdated bool) {
	if tx == nil || tx.Statement == nil || tx.Error != nil {
		return
	}

	actorID, ok := actorctx.ActorID(tx.Statement.Context)
	if !ok || actorID == uuid.Nil {
		// No actor in context: do nothing (system tasks/migrations/etc.)
		return
	}

	// If schema not parsed (rare), try parse.
	if tx.Statement.Schema == nil {
		if err := tx.Statement.Parse(tx.Statement.Dest); err != nil {
			// Don't break write path
			return
		}
	}

	sch := tx.Statement.Schema
	if sch == nil {
		return
	}

	if setCreated {
		// Only set created_by if it's currently NULL / zero pointer
		_ = setUUIDPtrFieldIfEmpty(tx, sch, "created_by", actorID)
	}

	if setUpdated {
		// Always set updated_by (overwrite)
		_ = setUUIDPtrField(tx, sch, "updated_by", actorID)
	}
}

func setUUIDPtrField(tx *gorm.DB, sch *schema.Schema, dbColumn string, actorID uuid.UUID) error {
	f := sch.LookUpField(dbColumn)
	if f == nil {
		// model doesn't have this column — fine
		return nil
	}

	// Set dest field value for all records (handles struct/slice)
	if err := f.Set(tx.Statement.Context, tx.Statement.ReflectValue, new(actorID)); err != nil {
		// keep error but don't crash; set it to tx.Error so gorm will return it
		err := tx.AddError(fmt.Errorf("gormblame: set %s: %w", dbColumn, err))
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func setUUIDPtrFieldIfEmpty(tx *gorm.DB, sch *schema.Schema, dbColumn string, actorID uuid.UUID) error {
	f := sch.LookUpField(dbColumn)
	if f == nil {
		return nil
	}

	// Check current value; if already set, do not overwrite.
	// For slices, this checks first element only; if you batch-create mixed states, that's on you.
	val, isZero := f.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue)
	if !isZero && val != nil {
		return nil
	}

	return setUUIDPtrField(tx, sch, dbColumn, actorID)
}

```


## platform/shared/errors/errors.go

```go
package apierr

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Status int    `json:"-"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Detail)
}

func BadRequest(detail string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: "bad_request", Detail: detail}
}

func Unauthorized(detail string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: "unauthorized", Detail: detail}
}

func Forbidden(detail string) *APIError {
	return &APIError{Status: http.StatusForbidden, Code: "forbidden", Detail: detail}
}

func NotFound(detail string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: "not_found", Detail: detail}
}

func Conflict(detail string) *APIError {
	return &APIError{Status: http.StatusConflict, Code: "conflict", Detail: detail}
}

func Internal(detail string) *APIError {
	return &APIError{Status: http.StatusInternalServerError, Code: "internal", Detail: detail}
}

```


## platform/shared/pagination/pagination.go

```go
package pagination

```


## platform/shared/searchkit/normalize.go

```go
package searchkit

import (
	"fmt"
	"strings"

	"github.com/NikolayNam/collabsphere-go/shared/strcase"
)

func NormalizeFilters(filters []Filter, allowed FilterSpec) ([]Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	// cache: raw field -> normalized snake
	fieldCache := make(map[string]string, minVal(len(filters), 32))
	// cache: normalized field -> allowed?
	allowedCache := make(map[string]bool, minVal(len(filters), 32))

	out := make([]Filter, 0, len(filters))
	for i, f := range filters {
		raw := strings.TrimSpace(f.Field)
		if raw == "" {
			return nil, fmt.Errorf("filters[%d].field is required", i)
		}

		// 1) normalize field (cached)
		field, ok := fieldCache[raw]
		if !ok {
			field = strcase.CamelToSnake(raw)
			fieldCache[raw] = field
		}
		if field == "" {
			return nil, fmt.Errorf("filters[%d].field is required", i)
		}

		// 2) strict op validation (no aliases)
		if !isValidOp(f.Op) {
			return nil, fmt.Errorf("filters[%d].op invalid: %s", i, f.Op)
		}

		// 3) whitelist check (cached)
		allowedOK, ok := allowedCache[field]
		if !ok {
			_, exists := allowed[field]
			allowedOK = exists
			allowedCache[field] = allowedOK
		}
		if !allowedOK {
			return nil, fmt.Errorf("filters[%d].field invalid: %s", i, f.Field)
		}

		// 4) basic structural validation (op/value compatibility)
		canon := Filter{Field: field, Op: f.Op, Value: f.Value}
		if err := ValidateBasic(canon); err != nil {
			return nil, fmt.Errorf("filters[%d]: %w", i, err)
		}

		out = append(out, canon)
	}

	return out, nil
}

func isValidOp(op Op) bool {
	switch op {
	case OpEQ, OpNE, OpGT, OpGTE, OpLT, OpLTE, OpLike, OpIn, OpBetween, OpIsNull, OpNotNull:
		return true
	default:
		return false
	}
}

func minVal(a, b int) int {
	if a < b {
		return a
	}
	return b
}

```


## platform/shared/searchkit/op_parse.go

```go
package searchkit

import (
	"fmt"
	"strings"
)

// ParseOpStrict — принимает только enum значения. Никаких "=".
func ParseOpStrict(op string) (Op, error) {
	s := strings.ToLower(strings.TrimSpace(op))
	switch Op(s) {
	case OpEQ, OpNE, OpGT, OpGTE, OpLT, OpLTE, OpLike, OpIn, OpBetween, OpIsNull, OpNotNull:
		return Op(s), nil
	default:
		return "", fmt.Errorf("invalid op: %s", op)
	}
}

```


## platform/shared/searchkit/payload.go

```go
package searchkit

// Payload — внутренний (domain-level) контракт для поиска.
// Он НЕ зависит от ogen/sqlx/sqlc и используется между transport -> app -> repo.
type Payload struct {
	Page      int      `json:"page"`
	Size      int      `json:"size"`
	OrderBy   []string `json:"orderBy"`
	OrderDesc bool     `json:"orderDesc"`
	Filters   []Filter `json:"filters"`
}

// Filter — канонический фильтр. ВАЖНО: Op типизированный (enum), не string.
type Filter struct {
	Field string `json:"field"` // приходит как camelCase; нормализуется (camel_to_snake) перед сборкой SQL
	Op    Op     `json:"op"`    // только eq/ne/gt/... (строго)
	Value any    `json:"value"` // для is_null/not_null может быть nil/ignored
}

// Op — перечисление допустимых операций фильтрации.
// Снаружи в API ты разрешаешь только эти значения (enum в OpenAPI).
type Op string

const (
	OpEQ      Op = "eq"
	OpNE      Op = "ne"
	OpGT      Op = "gt"
	OpGTE     Op = "gte"
	OpLT      Op = "lt"
	OpLTE     Op = "lte"
	OpLike    Op = "like"
	OpIn      Op = "in"
	OpBetween Op = "between"
	OpIsNull  Op = "is_null"
	OpNotNull Op = "not_null"
)

```


## platform/shared/searchkit/spec.go

```go
package searchkit

// SortSpec field -> SQL fragment (safe)
type SortSpec map[string]string

// FilterSpec field -> FieldSpec
type FilterSpec map[string]FieldSpec

// FieldSpec описывает, как поле выглядит в SQL и как нормализовать значение.
type FieldSpec struct {
	SQLExpr string // Например: "u.email", "c.full_name"
	Type    string // "uuid"|"text"|"int"|"bool"|"time"|...
}

```


## platform/shared/searchkit/validate.go

```go
package searchkit

import (
	"fmt"
	"strings"
)

// ValidateBasic — проверяет форму фильтра и совместимость Op/Value.
func ValidateBasic(f Filter) error {
	if strings.TrimSpace(f.Field) == "" {
		return fmt.Errorf("field is required")
	}

	switch f.Op {
	case OpEQ, OpNE, OpGT, OpGTE, OpLT, OpLTE, OpLike:
		if f.Value == nil {
			return fmt.Errorf("value is required for op=%s", f.Op)
		}
	case OpIn:
		if f.Value == nil {
			return fmt.Errorf("value is required for op=in")
		}
	case OpBetween:
		if f.Value == nil {
			return fmt.Errorf("value is required for op=between")
		}
	case OpIsNull, OpNotNull:
		// value ignored
	default:
		return fmt.Errorf("unsupported op: %s", f.Op)
	}

	return nil
}

```


## platform/shared/strcase/camel_to_case.go

```go
package strcase

import (
	"fmt"
	"regexp"
	"strings"
)

var camelRx = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func CamelToSnake(s string) string {
	if s == "" {
		return s
	}
	var out = camelRx.ReplaceAllString(s, `${1}_${2}`)
	out = strings.ReplaceAll(out, "-", "_")
	return strings.ToLower(out)
}

func CamelToSnakeStrict(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("value is required")
	}
	return CamelToSnake(s), nil
}

```

