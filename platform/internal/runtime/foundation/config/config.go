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
	TZ      string `env:"TZ" envDefault:"UTC"`
	APP     App
	DB      DB
	Auth    Auth
	Storage Storage
}

type Auth struct {
	JWTSecret         string        `env:"AUTH_JWT_SECRET"`
	JWTSecretFile     string        `env:"AUTH_JWT_SECRET_FILE"`
	AccessTTL         time.Duration `env:"AUTH_ACCESS_TTL" envDefault:"15m"`
	RefreshSessionTTL time.Duration `env:"AUTH_REFRESH_TTL" envDefault:"720h"`
}

type Storage struct {
	S3 S3
}

type S3 struct {
	Enabled     bool          `env:"STORAGE_S3_ENABLED" envDefault:"false"`
	Endpoint    string        `env:"STORAGE_S3_ENDPOINT"`
	Region      string        `env:"STORAGE_S3_REGION" envDefault:"us-east-1"`
	AccessKey   string        `env:"STORAGE_S3_ACCESS_KEY"`
	SecretKey   string        `env:"STORAGE_S3_SECRET_KEY"`
	Bucket      string        `env:"STORAGE_S3_BUCKET"`
	PathStyle   bool          `env:"STORAGE_S3_PATH_STYLE" envDefault:"true"`
	PresignTTL  time.Duration `env:"STORAGE_S3_PRESIGN_TTL" envDefault:"15m"`
	DownloadTTL time.Duration `env:"STORAGE_S3_DOWNLOAD_TTL" envDefault:"5m"`
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

func (a Auth) JWTSecretValue() (string, error) {
	if strings.TrimSpace(a.JWTSecret) != "" {
		return a.JWTSecret, nil
	}
	if strings.TrimSpace(a.JWTSecretFile) == "" {
		return "", errors.New("auth jwt secret is empty")
	}

	b, err := os.ReadFile(a.JWTSecretFile)
	if err != nil {
		return "", err
	}

	secret := strings.TrimSpace(string(b))
	if secret == "" {
		return "", errors.New("auth jwt secret file is empty")
	}
	return secret, nil
}

func (s S3) Validate() error {
	if !s.Enabled {
		return nil
	}

	switch {
	case strings.TrimSpace(s.Endpoint) == "":
		return errors.New("storage s3 endpoint is empty")
	case strings.TrimSpace(s.Region) == "":
		return errors.New("storage s3 region is empty")
	case strings.TrimSpace(s.AccessKey) == "":
		return errors.New("storage s3 access key is empty")
	case strings.TrimSpace(s.SecretKey) == "":
		return errors.New("storage s3 secret key is empty")
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

func buildSearchPath(primary string) []string {
	values := []string{primary, "auth", "iam", "org", "catalog", "sales", "storage", "integration", "public"}
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
