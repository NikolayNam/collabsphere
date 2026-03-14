package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
)

type Config struct {
	TZ               string `env:"TZ" envDefault:"UTC"`
	APP              App
	DB               DB
	Auth             Auth
	Storage          Storage
	Collab           Collab
	Realtime         Realtime
	Conference       Conference
	Transcription    Transcription
	DocumentAnalysis DocumentAnalysis
}

type RuntimeProfile string

const (
	ProfileAPI       RuntimeProfile = "api"
	ProfileWorker    RuntimeProfile = "worker"
	ProfileContracts RuntimeProfile = "contracts"
	ProfileMigrate   RuntimeProfile = "migrate"
	ProfileSeed      RuntimeProfile = "seed"
)

type Auth struct {
	JWTSecret             string        `env:"AUTH_JWT_SECRET"`
	JWTSecretFile         string        `env:"AUTH_JWT_SECRET_FILE"`
	AccessTTL             time.Duration `env:"AUTH_ACCESS_TTL" envDefault:"15m"`
	RefreshSessionTTL     time.Duration `env:"AUTH_REFRESH_TTL" envDefault:"720h"`
	GuestAccessTTL        time.Duration `env:"AUTH_GUEST_ACCESS_TTL" envDefault:"24h"`
	PasswordLoginEnabled  bool          `env:"AUTH_PASSWORD_LOGIN_ENABLED" envDefault:"true"`
	LocalSignupEnabled    bool          `env:"AUTH_LOCAL_SIGNUP_ENABLED" envDefault:"true"`
	PlatformBootstrapIDs  string        `env:"AUTH_PLATFORM_BOOTSTRAP_ACCOUNT_IDS"`
	PlatformAutoGrantFile string        `env:"AUTH_PLATFORM_AUTO_GRANT_FILE"`
	BrowserDefaultReturn  string        `env:"AUTH_BROWSER_DEFAULT_RETURN_URL"`
	BrowserRedirects      string        `env:"AUTH_BROWSER_REDIRECT_ORIGINS"`
	BrowserTicketTTL      time.Duration `env:"AUTH_BROWSER_TICKET_TTL" envDefault:"1m"`
	Zitadel               Zitadel
}

type Zitadel struct {
	Enabled          bool          `env:"AUTH_ZITADEL_ENABLED" envDefault:"false"`
	IssuerURL        string        `env:"AUTH_ZITADEL_ISSUER_URL"`
	ClientID         string        `env:"AUTH_ZITADEL_CLIENT_ID"`
	ClientSecret     string        `env:"AUTH_ZITADEL_CLIENT_SECRET"`
	ClientSecretFile string        `env:"AUTH_ZITADEL_CLIENT_SECRET_FILE"`
	AdminToken       string        `env:"AUTH_ZITADEL_ADMIN_TOKEN"`
	AdminTokenFile   string        `env:"AUTH_ZITADEL_ADMIN_TOKEN_FILE"`
	RedirectURL      string        `env:"AUTH_ZITADEL_REDIRECT_URL"`
	Scopes           string        `env:"AUTH_ZITADEL_SCOPES" envDefault:"openid profile email"`
	StateTTL         time.Duration `env:"AUTH_ZITADEL_STATE_TTL" envDefault:"15m"`
	NonceTTL         time.Duration `env:"AUTH_ZITADEL_NONCE_TTL" envDefault:"15m"`
	HTTPTimeout      time.Duration `env:"AUTH_ZITADEL_HTTP_TIMEOUT" envDefault:"10s"`
}

type Storage struct {
	S3 S3
}

type S3 struct {
	Enabled        bool          `env:"STORAGE_S3_ENABLED" envDefault:"false"`
	Endpoint       string        `env:"STORAGE_S3_ENDPOINT"`
	PublicEndpoint string        `env:"STORAGE_S3_PUBLIC_ENDPOINT"`
	Region         string        `env:"STORAGE_S3_REGION" envDefault:"us-east-1"`
	AccessKey      string        `env:"STORAGE_S3_ACCESS_KEY"`
	AccessKeyFile  string        `env:"STORAGE_S3_ACCESS_KEY_FILE"`
	SecretKey      string        `env:"STORAGE_S3_SECRET_KEY"`
	SecretKeyFile  string        `env:"STORAGE_S3_SECRET_KEY_FILE"`
	Bucket         string        `env:"STORAGE_S3_BUCKET"`
	PathStyle      bool          `env:"STORAGE_S3_PATH_STYLE" envDefault:"true"`
	PresignTTL     time.Duration `env:"STORAGE_S3_PRESIGN_TTL" envDefault:"15m"`
	DownloadTTL    time.Duration `env:"STORAGE_S3_DOWNLOAD_TTL" envDefault:"5m"`
}

type App struct {
	Title             string        `env:"APPLICATION_TITLE"`
	Version           string        `env:"APPLICATION_VERSION"`
	Address           string        `env:"APPLICATION_ADDRESS"`
	Host              string        `env:"APPLICATION_HOST" envDefault:"0.0.0.0"`
	Port              string        `env:"APPLICATION_PORT" envDefault:"8080"`
	PublicBaseURL     string        `env:"APPLICATION_PUBLIC_BASE_URL"`
	TimeoutRead       time.Duration `env:"APPLICATION_TIMEOUT_READ" envDefault:"15s"`
	TimeoutWrite      time.Duration `env:"APPLICATION_TIMEOUT_WRITE" envDefault:"15s"`
	TimeoutIdle       time.Duration `env:"APPLICATION_TIMEOUT_IDLE" envDefault:"60s"`
	MetricsEnabled    bool          `env:"APPLICATION_METRICS_ENABLED" envDefault:"false"`
	MetricsPath       string        `env:"APPLICATION_METRICS_PATH" envDefault:"/metrics"`
	Debug             bool          `env:"APPLICATION_DEBUG" envDefault:"false"`
	Environment       string        `env:"APPLICATION_ENVIRONMENT" envDefault:"dev"`
	LogLevel          string        `env:"APPLICATION_LOG_LEVEL" envDefault:"INFO"`
	TrustProxyHeaders bool          `env:"APPLICATION_TRUST_PROXY_HEADERS" envDefault:"false"`
}

type DB struct {
	Host string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port int    `env:"POSTGRES_PORT" envDefault:"5432"`

	DBName   string `env:"POSTGRES_DB" envDefault:"postgres"`
	DBSchema string `env:"POSTGRES_SCHEMA"`

	Username string `env:"POSTGRES_USER" envDefault:"postgres"`

	Password     string `env:"POSTGRES_PASSWORD"`
	PasswordFile string `env:"POSTGRES_PASSWORD_FILE"`

	Debug bool `env:"POSTGRES_DEBUG" envDefault:"false"`

	MaxOpenConns    int           `env:"POSTGRES_MAX_OPEN_CONNS" envDefault:"30"`
	MaxIdleConns    int           `env:"POSTGRES_MAX_IDLE_CONNS" envDefault:"15"`
	ConnMaxLifetime time.Duration `env:"POSTGRES_CONN_MAX_LIFETIME" envDefault:"30m"`
	ConnMaxIdleTime time.Duration `env:"POSTGRES_CONN_MAX_IDLE_TIME" envDefault:"5m"`
}

type Collab struct {
	GuestInviteTTL          time.Duration `env:"COLLAB_GUEST_INVITE_TTL" envDefault:"72h"`
	WSAllowQueryAccessToken bool          `env:"COLLAB_WS_ALLOW_QUERY_ACCESS_TOKEN" envDefault:"false"`
}

type Realtime struct {
	Redis Redis
}

type Redis struct {
	Enabled       bool          `env:"REALTIME_REDIS_ENABLED" envDefault:"false"`
	Address       string        `env:"REALTIME_REDIS_ADDRESS"`
	Password      string        `env:"REALTIME_REDIS_PASSWORD"`
	PasswordFile  string        `env:"REALTIME_REDIS_PASSWORD_FILE"`
	DB            int           `env:"REALTIME_REDIS_DB" envDefault:"0"`
	ChannelPrefix string        `env:"REALTIME_REDIS_CHANNEL_PREFIX" envDefault:"collabsphere:realtime"`
	PresenceTTL   time.Duration `env:"REALTIME_REDIS_PRESENCE_TTL" envDefault:"30s"`
	TypingTTL     time.Duration `env:"REALTIME_REDIS_TYPING_TTL" envDefault:"5s"`
}

type Conference struct {
	Provider string `env:"CONFERENCE_PROVIDER" envDefault:"mediasoup"`
}

type Transcription struct {
	Enabled         bool          `env:"TRANSCRIPTION_ENABLED" envDefault:"false"`
	Endpoint        string        `env:"TRANSCRIPTION_ENDPOINT"`
	APIKey          string        `env:"TRANSCRIPTION_API_KEY"`
	APIKeyFile      string        `env:"TRANSCRIPTION_API_KEY_FILE"`
	Model           string        `env:"TRANSCRIPTION_MODEL" envDefault:"whisper-1"`
	RequestTimeout  time.Duration `env:"TRANSCRIPTION_REQUEST_TIMEOUT" envDefault:"10m"`
	WorkerPollEvery time.Duration `env:"TRANSCRIPTION_WORKER_POLL_EVERY" envDefault:"10s"`
}

type DocumentAnalysis struct {
	Enabled         bool          `env:"DOCUMENT_ANALYSIS_ENABLED" envDefault:"false"`
	Endpoint        string        `env:"DOCUMENT_ANALYSIS_ENDPOINT"`
	APIKey          string        `env:"DOCUMENT_ANALYSIS_API_KEY"`
	APIKeyFile      string        `env:"DOCUMENT_ANALYSIS_API_KEY_FILE"`
	Provider        string        `env:"DOCUMENT_ANALYSIS_PROVIDER" envDefault:"generic-http"`
	Model           string        `env:"DOCUMENT_ANALYSIS_MODEL" envDefault:"legal-doc-ocr-v1"`
	RequestTimeout  time.Duration `env:"DOCUMENT_ANALYSIS_REQUEST_TIMEOUT" envDefault:"2m"`
	WorkerPollEvery time.Duration `env:"DOCUMENT_ANALYSIS_WORKER_POLL_EVERY" envDefault:"10s"`
}

func New() *Config {
	return NewFor(ProfileAPI)
}

func NewFor(profile RuntimeProfile) *Config {
	var c Config

	if err := env.Parse(&c); err != nil {
		log.Fatalf("failed to parse env: %s", err)
	}

	if strings.TrimSpace(c.APP.Address) == "" {
		c.APP.Address = c.APP.ListenAddress()
	}
	c.applyProfileDefaults(profile)

	if err := applyTZ(c.TZ); err != nil {
		log.Fatalf("invalid TZ: %s", err)
	}
	if err := c.ValidateFor(profile); err != nil {
		log.Fatalf("invalid %s configuration: %s", normalizeProfile(profile), err)
	}

	return &c
}

func (c Config) Validate() error {
	return c.ValidateFor(ProfileAPI)
}

func (c Config) ValidateFor(profile RuntimeProfile) error {
	switch normalizeProfile(profile) {
	case ProfileAPI:
		if err := c.APP.Validate(); err != nil {
			return err
		}
		if err := c.DB.Validate(); err != nil {
			return err
		}
		if err := c.Auth.Validate(c.APP); err != nil {
			return err
		}
		if err := c.validatePlatformAuthConfig(); err != nil {
			return err
		}
		if err := c.Storage.S3.Validate(); err != nil {
			return err
		}
		if err := c.Realtime.Redis.Validate(); err != nil {
			return err
		}
		if err := c.Conference.Validate(); err != nil {
			return err
		}
		if err := c.Transcription.Validate(); err != nil {
			return err
		}
		if err := c.DocumentAnalysis.Validate(); err != nil {
			return err
		}
		return nil
	case ProfileWorker:
		if err := c.DB.Validate(); err != nil {
			return err
		}
		if err := c.Auth.ValidateJWT(); err != nil {
			return err
		}
		if err := c.Storage.S3.Validate(); err != nil {
			return err
		}
		if err := c.Conference.Validate(); err != nil {
			return err
		}
		if err := c.Transcription.Validate(); err != nil {
			return err
		}
		if err := c.DocumentAnalysis.Validate(); err != nil {
			return err
		}
		return nil
	case ProfileContracts:
		return c.APP.ValidateContracts()
	case ProfileMigrate, ProfileSeed:
		return c.DB.Validate()
	default:
		return fmt.Errorf("unknown runtime profile %q", profile)
	}
}

func (d DB) PasswordValue() (string, error) {
	return readRequiredSecret("postgres password", d.Password, d.PasswordFile)
}

func (a App) MetricsRoutePath() string {
	path := strings.TrimSpace(a.MetricsPath)
	if path == "" {
		return "/metrics"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func (a App) NormalizedEnvironment() string {
	value := strings.ToLower(strings.TrimSpace(a.Environment))
	if value == "" {
		return "dev"
	}
	return value
}

func (a App) Validate() error {
	if strings.TrimSpace(a.Title) == "" {
		return errors.New("application title is empty")
	}
	if strings.TrimSpace(a.Version) == "" {
		return errors.New("application version is empty")
	}
	return a.ValidateServerRuntime()
}

func (a App) ValidateContracts() error {
	if strings.TrimSpace(a.Title) == "" {
		return errors.New("application title is empty")
	}
	if strings.TrimSpace(a.Version) == "" {
		return errors.New("application version is empty")
	}
	return nil
}

func (a App) ValidateServerRuntime() error {
	if strings.TrimSpace(a.ListenAddress()) == "" {
		return errors.New("application listen address is empty")
	}
	if a.TimeoutRead <= 0 {
		return errors.New("application read timeout must be positive")
	}
	if a.TimeoutWrite <= 0 {
		return errors.New("application write timeout must be positive")
	}
	if a.TimeoutIdle <= 0 {
		return errors.New("application idle timeout must be positive")
	}
	if _, err := validatePublicBaseURL(a.PublicBaseURL); err != nil {
		return err
	}
	return nil
}

func (a Auth) JWTSecretValue() (string, error) {
	return readRequiredSecret("auth jwt secret", a.JWTSecret, a.JWTSecretFile)
}

func (a Auth) BrowserRedirectOriginList() []string {
	return splitList(a.BrowserRedirects)
}

func (a Auth) Validate(app App) error {
	if err := a.ValidateJWT(); err != nil {
		return err
	}
	if err := a.ValidateBrowser(app); err != nil {
		return err
	}
	return nil
}

func (a Auth) ValidateJWT() error {
	if _, err := a.JWTSecretValue(); err != nil {
		return err
	}
	switch {
	case a.AccessTTL <= 0:
		return errors.New("auth access ttl must be positive")
	case a.RefreshSessionTTL <= 0:
		return errors.New("auth refresh ttl must be positive")
	case a.GuestAccessTTL <= 0:
		return errors.New("auth guest access ttl must be positive")
	default:
		return nil
	}
}

func (a Auth) ValidateBrowser(app App) error {
	switch {
	case a.BrowserTicketTTL <= 0:
		return errors.New("auth browser ticket ttl must be positive")
	}
	if err := validateBrowserReturnURL(a.BrowserDefaultReturn); err != nil {
		return err
	}
	if err := validateBrowserRedirectOrigins(a.BrowserRedirectOriginList()); err != nil {
		return err
	}
	if _, err := validatePublicBaseURL(app.PublicBaseURL); err != nil {
		return err
	}
	if err := a.Zitadel.Validate(); err != nil {
		return err
	}
	return nil
}

func (c Config) validatePlatformAuthConfig() error {
	if _, err := c.Auth.PlatformBootstrapAccountUUIDs(); err != nil {
		return fmt.Errorf("auth platform bootstrap ids: %w", err)
	}
	if _, err := c.Auth.PlatformAutoGrantRules(); err != nil {
		return fmt.Errorf("auth platform auto grant rules: %w", err)
	}
	return nil
}

func (c *Config) applyProfileDefaults(profile RuntimeProfile) {
	if normalizeProfile(profile) != ProfileContracts {
		return
	}

	c.Auth.JWTSecret = "contracts-placeholder-secret"
	c.Auth.JWTSecretFile = ""
	c.Auth.PlatformBootstrapIDs = ""
	c.Auth.PlatformAutoGrantFile = ""
	c.Auth.Zitadel = Zitadel{}
	c.Storage.S3 = S3{}
	c.Realtime.Redis = Redis{}
	c.Conference = Conference{}
	c.Transcription = Transcription{}
	c.DocumentAnalysis = DocumentAnalysis{}
}

func normalizeProfile(profile RuntimeProfile) RuntimeProfile {
	value := strings.ToLower(strings.TrimSpace(string(profile)))
	if value == "" {
		return ProfileAPI
	}
	return RuntimeProfile(value)
}

func (a Auth) PlatformBootstrapAccountUUIDs() ([]uuid.UUID, error) {
	parts := splitList(a.PlatformBootstrapIDs)
	out := make([]uuid.UUID, 0, len(parts))
	seen := make(map[uuid.UUID]struct{}, len(parts))
	for _, part := range parts {
		parsed, err := uuid.Parse(part)
		if err != nil {
			return nil, fmt.Errorf("invalid account id %q: %w", part, err)
		}
		if _, ok := seen[parsed]; ok {
			continue
		}
		seen[parsed] = struct{}{}
		out = append(out, parsed)
	}
	return out, nil
}

func (z Zitadel) ClientSecretValue() (string, error) {
	if !z.Enabled {
		return "", nil
	}
	return readRequiredSecret("auth zitadel client secret", z.ClientSecret, z.ClientSecretFile)
}

func (z Zitadel) AdminTokenValue() (string, error) {
	return readOptionalSecret("auth zitadel admin token", z.AdminToken, z.AdminTokenFile)
}

func (z Zitadel) ScopeList() []string {
	return splitList(z.Scopes)
}

func (z Zitadel) Validate() error {
	if !z.Enabled {
		return nil
	}
	if _, err := z.ClientSecretValue(); err != nil {
		return err
	}
	if strings.TrimSpace(z.IssuerURL) == "" {
		return errors.New("auth zitadel issuer url is empty")
	}
	if strings.TrimSpace(z.ClientID) == "" {
		return errors.New("auth zitadel client id is empty")
	}
	if strings.TrimSpace(z.RedirectURL) == "" {
		return errors.New("auth zitadel redirect url is empty")
	}
	if len(z.ScopeList()) == 0 {
		return errors.New("auth zitadel scopes are empty")
	}
	if z.StateTTL <= 0 {
		return errors.New("auth zitadel state ttl must be positive")
	}
	if z.NonceTTL <= 0 {
		return errors.New("auth zitadel nonce ttl must be positive")
	}
	if z.HTTPTimeout <= 0 {
		return errors.New("auth zitadel http timeout must be positive")
	}
	return nil
}

func (s S3) AccessKeyValue() (string, error) {
	return readRequiredSecret("storage s3 access key", s.AccessKey, s.AccessKeyFile)
}

func (s S3) SecretKeyValue() (string, error) {
	return readRequiredSecret("storage s3 secret key", s.SecretKey, s.SecretKeyFile)
}

func (r Redis) PasswordValue() (string, error) {
	return readOptionalSecret("realtime redis password", r.Password, r.PasswordFile)
}

func (t Transcription) APIKeyValue() (string, error) {
	return readOptionalSecret("transcription api key", t.APIKey, t.APIKeyFile)
}

func (d DocumentAnalysis) APIKeyValue() (string, error) {
	return readOptionalSecret("document analysis api key", d.APIKey, d.APIKeyFile)
}

func (s S3) Validate() error {
	if !s.Enabled {
		return nil
	}
	if _, err := s.AccessKeyValue(); err != nil {
		return err
	}
	if _, err := s.SecretKeyValue(); err != nil {
		return err
	}

	switch {
	case strings.TrimSpace(s.Endpoint) == "":
		return errors.New("storage s3 endpoint is empty")
	case strings.TrimSpace(s.Region) == "":
		return errors.New("storage s3 region is empty")
	case strings.TrimSpace(s.Bucket) == "":
		return errors.New("storage s3 bucket is empty")
	case s.PresignTTL <= 0:
		return errors.New("storage s3 presign ttl must be positive")
	case s.DownloadTTL <= 0:
		return errors.New("storage s3 download ttl must be positive")
	default:
		return nil
	}
}

func (r Redis) Validate() error {
	if !r.Enabled {
		return nil
	}
	if _, err := r.PasswordValue(); err != nil {
		return err
	}

	switch {
	case strings.TrimSpace(r.Address) == "":
		return errors.New("realtime redis address is empty")
	case strings.TrimSpace(r.ChannelPrefix) == "":
		return errors.New("realtime redis channel prefix is empty")
	case r.PresenceTTL <= 0:
		return errors.New("realtime redis presence ttl must be positive")
	case r.TypingTTL <= 0:
		return errors.New("realtime redis typing ttl must be positive")
	default:
		return nil
	}
}

func (c Conference) ProviderValue() string {
	provider := strings.ToLower(strings.TrimSpace(c.Provider))
	if provider == "" {
		return "mediasoup"
	}
	return provider
}

func (c Conference) Validate() error {
	if c.ProviderValue() != "mediasoup" {
		return errors.New("only mediasoup conference provider is supported")
	}
	return nil
}

func (t Transcription) Validate() error {
	if !t.Enabled {
		return nil
	}
	if _, err := t.APIKeyValue(); err != nil {
		return err
	}

	switch {
	case strings.TrimSpace(t.Endpoint) == "":
		return errors.New("transcription endpoint is empty")
	case t.RequestTimeout <= 0:
		return errors.New("transcription request timeout must be positive")
	case t.WorkerPollEvery <= 0:
		return errors.New("transcription worker poll interval must be positive")
	default:
		return nil
	}
}

func (d DocumentAnalysis) Validate() error {
	if !d.Enabled {
		return nil
	}
	if _, err := d.APIKeyValue(); err != nil {
		return err
	}

	switch {
	case strings.TrimSpace(d.Endpoint) == "":
		return errors.New("document analysis endpoint is empty")
	case strings.TrimSpace(d.Provider) == "":
		return errors.New("document analysis provider is empty")
	case d.RequestTimeout <= 0:
		return errors.New("document analysis request timeout must be positive")
	case d.WorkerPollEvery <= 0:
		return errors.New("document analysis worker poll interval must be positive")
	default:
		return nil
	}
}

func applyTZ(tz string) error {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
}

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
	q.Set("search_path", strings.Join(buildSearchPath(d.DBSchema), ","))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (a App) ListenAddress() string {
	if strings.TrimSpace(a.Address) != "" {
		return strings.TrimSpace(a.Address)
	}

	host := strings.TrimSpace(a.Host)
	if host == "" {
		host = "0.0.0.0"
	}
	port := strings.TrimSpace(a.Port)
	if port == "" {
		port = "8080"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Sprintf("%s:%s", host, port)
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func (d DB) Validate() error {
	if strings.TrimSpace(d.Host) == "" {
		return errors.New("postgres host is empty")
	}
	if d.Port <= 0 {
		return errors.New("postgres port must be positive")
	}
	if strings.TrimSpace(d.DBName) == "" {
		return errors.New("postgres database name is empty")
	}
	if strings.TrimSpace(d.DBSchema) == "" {
		return errors.New("postgres schema is empty")
	}
	if strings.TrimSpace(d.Username) == "" {
		return errors.New("postgres username is empty")
	}
	if _, err := d.PasswordValue(); err != nil {
		return err
	}
	if d.MaxOpenConns < 0 {
		return errors.New("postgres max open conns must be non-negative")
	}
	if d.MaxIdleConns < 0 {
		return errors.New("postgres max idle conns must be non-negative")
	}
	if d.MaxOpenConns > 0 && d.MaxIdleConns > d.MaxOpenConns {
		return errors.New("postgres max idle conns must be less than or equal to max open conns")
	}
	if d.ConnMaxLifetime < 0 {
		return errors.New("postgres conn max lifetime must be non-negative")
	}
	if d.ConnMaxIdleTime < 0 {
		return errors.New("postgres conn max idle time must be non-negative")
	}
	return nil
}

func validatePublicBaseURL(raw string) (*url.URL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("application public base url is invalid: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("application public base url must be absolute")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("application public base url scheme is invalid")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("application public base url must not contain query or fragment")
	}
	path := strings.TrimSpace(parsed.EscapedPath())
	if path != "" && path != "/" {
		return nil, errors.New("application public base url must not contain a path")
	}
	return parsed, nil
}

func validateBrowserReturnURL(raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	if strings.HasPrefix(value, "/") {
		return nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("auth browser default return url is invalid: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("auth browser default return url must be absolute or start with '/'")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("auth browser default return url scheme is invalid")
	}
	return nil
}

func validateBrowserRedirectOrigins(origins []string) error {
	for _, origin := range origins {
		parsed, err := url.Parse(origin)
		if err != nil {
			return fmt.Errorf("auth browser redirect origin %q is invalid: %w", origin, err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("auth browser redirect origin %q must be absolute", origin)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("auth browser redirect origin %q scheme is invalid", origin)
		}
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			return fmt.Errorf("auth browser redirect origin %q must not contain query or fragment", origin)
		}
		path := strings.TrimSpace(parsed.EscapedPath())
		if path != "" && path != "/" {
			return fmt.Errorf("auth browser redirect origin %q must not contain a path", origin)
		}
	}
	return nil
}

func readRequiredSecret(label, value, file string) (string, error) {
	secret, err := readOptionalSecret(label, value, file)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("%s is empty", label)
	}
	return secret, nil
}

func readOptionalSecret(label, value, file string) (string, error) {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value), nil
	}
	if strings.TrimSpace(file) == "" {
		return "", nil
	}

	b, err := os.ReadFile(strings.TrimSpace(file))
	if err != nil {
		return "", fmt.Errorf("read %s file: %w", label, err)
	}
	secret := strings.TrimSpace(string(b))
	if secret == "" {
		return "", fmt.Errorf("%s file is empty", label)
	}
	return secret, nil
}

func splitList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return out
}
func buildSearchPath(primary string) []string {
	values := []string{primary, "auth", "iam", "org", "catalog", "sales", "storage", "integration", "collab", "public"}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}

	return out
}
