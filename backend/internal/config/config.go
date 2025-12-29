package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceName string `mapstructure:"service_name"`
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`

	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Keycloak KeycloakConfig `mapstructure:"keycloak"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

type KeycloakConfig struct {
	URL          string `mapstructure:"url"`
	Realm        string `mapstructure:"realm"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

// Load reads configuration from environment variables and optional config files.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.SetEnvPrefix("AUB")
	v.AutomaticEnv()

	// Bind existing environment variable names (DB_*, REDIS_*, etc.) for compatibility
	// with docker-compose and infrastructure configuration
	bindEnvVars(v)

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("service_name", "usage-service")
	v.SetDefault("port", 8080)
	v.SetDefault("environment", "development")

	// Database defaults - credentials MUST be provided via environment variables
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "api_billing")
	v.SetDefault("database.user", "billing")
	v.SetDefault("database.password", "") // REQUIRED: Set via DB_PASSWORD env var
	v.SetDefault("database.ssl_mode", "disable")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")

	// Keycloak defaults - client_secret MUST be provided via environment variables
	v.SetDefault("keycloak.url", "http://localhost:8080")
	v.SetDefault("keycloak.realm", "api-billing")
	v.SetDefault("keycloak.client_id", "api-billing")
	v.SetDefault("keycloak.client_secret", "") // REQUIRED: Set via KEYCLOAK_CLIENT_SECRET env var

	// MinIO defaults - credentials MUST be provided via environment variables
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.access_key", "") // REQUIRED: Set via MINIO_ACCESS_KEY env var
	v.SetDefault("minio.secret_key", "") // REQUIRED: Set via MINIO_SECRET_KEY env var
	v.SetDefault("minio.bucket", "invoices")
	v.SetDefault("minio.use_ssl", false)

	// SMTP defaults - credentials MUST be provided via environment variables if needed
	v.SetDefault("smtp.host", "")
	v.SetDefault("smtp.port", 587)
	v.SetDefault("smtp.username", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from", "")
}

// bindEnvVars binds infrastructure environment variable names to config keys.
// This enables compatibility with docker-compose and .env file conventions.
func bindEnvVars(v *viper.Viper) {
	// Application
	_ = v.BindEnv("environment", "APP_ENV")
	_ = v.BindEnv("port", "USAGE_SERVICE_PORT")

	// Database (DB_* convention from docker-compose)
	_ = v.BindEnv("database.host", "DB_HOST")
	_ = v.BindEnv("database.port", "DB_PORT")
	_ = v.BindEnv("database.name", "DB_NAME")
	_ = v.BindEnv("database.user", "DB_USER")
	_ = v.BindEnv("database.password", "DB_PASSWORD")
	_ = v.BindEnv("database.ssl_mode", "DB_SSL_MODE")

	// Redis (REDIS_* convention)
	_ = v.BindEnv("redis.host", "REDIS_HOST")
	_ = v.BindEnv("redis.port", "REDIS_PORT")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")

	// Keycloak (KEYCLOAK_* convention)
	_ = v.BindEnv("keycloak.url", "KEYCLOAK_URL")
	_ = v.BindEnv("keycloak.realm", "KEYCLOAK_REALM")
	_ = v.BindEnv("keycloak.client_id", "KEYCLOAK_CLIENT_ID")
	_ = v.BindEnv("keycloak.client_secret", "KEYCLOAK_CLIENT_SECRET")

	// MinIO (MINIO_* convention)
	_ = v.BindEnv("minio.endpoint", "MINIO_ENDPOINT")
	_ = v.BindEnv("minio.access_key", "MINIO_ACCESS_KEY")
	_ = v.BindEnv("minio.secret_key", "MINIO_SECRET_KEY")
	_ = v.BindEnv("minio.bucket", "MINIO_BUCKET")
	_ = v.BindEnv("minio.use_ssl", "MINIO_USE_SSL")

	// SMTP (SMTP_* convention)
	_ = v.BindEnv("smtp.host", "SMTP_HOST")
	_ = v.BindEnv("smtp.port", "SMTP_PORT")
	_ = v.BindEnv("smtp.username", "SMTP_USERNAME")
	_ = v.BindEnv("smtp.password", "SMTP_PASSWORD")
	_ = v.BindEnv("smtp.from", "SMTP_FROM")
}
