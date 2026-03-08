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

type Auth struct {
	JWTSecret         string        `env:"AUTH_JWT_SECRET"`
	JWTSecretFile     string        `env:"AUTH_JWT_SECRET_FILE"`
	AccessTTL         time.Duration `env:"AUTH_ACCESS_TTL" envDefault:"15m"`
	RefreshSessionTTL time.Duration `env:"AUTH_REFRESH_TTL" envDefault:"720h"`
	GuestAccessTTL    time.Duration `env:"AUTH_GUEST_ACCESS_TTL" envDefault:"24h"`
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
	Title         string        `env:"APPLICATION_TITLE,required"`
	Version       string        `env:"APPLICATION_VERSION,required"`
	Address       string        `env:"APPLICATION_ADDRESS"`
	Host          string        `env:"APPLICATION_HOST" envDefault:"0.0.0.0"`
	Port          string        `env:"APPLICATION_PORT" envDefault:"8080"`
	PublicBaseURL string        `env:"APPLICATION_PUBLIC_BASE_URL"`
	TimeoutRead   time.Duration `env:"APPLICATION_TIMEOUT_READ" envDefault:"15s"`
	TimeoutWrite  time.Duration `env:"APPLICATION_TIMEOUT_WRITE" envDefault:"15s"`
	TimeoutIdle   time.Duration `env:"APPLICATION_TIMEOUT_IDLE" envDefault:"60s"`
	Debug         bool          `env:"APPLICATION_DEBUG" envDefault:"false"`
	Environment   string        `env:"APPLICATION_ENVIRONMENT" envDefault:"dev"`
	LogLevel      string        `env:"APPLICATION_LOG_LEVEL" envDefault:"INFO"`
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

type Collab struct {
	GuestInviteTTL time.Duration `env:"COLLAB_GUEST_INVITE_TTL" envDefault:"72h"`
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
	var c Config

	if err := env.Parse(&c); err != nil {
		log.Fatalf("failed to parse env: %s", err)
	}

	if strings.TrimSpace(c.APP.Address) == "" {
		c.APP.Address = c.APP.ListenAddress()
	}

	if err := applyTZ(c.TZ); err != nil {
		log.Fatalf("invalid TZ: %s", err)
	}

	return &c
}

func (d DB) PasswordValue() (string, error) {
	return readRequiredSecret("postgres password", d.Password, d.PasswordFile)
}

func (a Auth) JWTSecretValue() (string, error) {
	return readRequiredSecret("auth jwt secret", a.JWTSecret, a.JWTSecretFile)
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
