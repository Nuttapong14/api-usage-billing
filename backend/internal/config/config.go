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
	v.SetDefault("database.password", "") // REQUIRED: Set via AUB_DATABASE_PASSWORD env var
	v.SetDefault("database.ssl_mode", "disable")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")

	// Keycloak defaults - client_secret MUST be provided via environment variables
	v.SetDefault("keycloak.url", "http://localhost:8080")
	v.SetDefault("keycloak.realm", "api-billing")
	v.SetDefault("keycloak.client_id", "api-billing")
	v.SetDefault("keycloak.client_secret", "") // REQUIRED: Set via AUB_KEYCLOAK_CLIENT_SECRET env var

	// MinIO defaults - credentials MUST be provided via environment variables
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.access_key", "") // REQUIRED: Set via AUB_MINIO_ACCESS_KEY env var
	v.SetDefault("minio.secret_key", "") // REQUIRED: Set via AUB_MINIO_SECRET_KEY env var
	v.SetDefault("minio.bucket", "invoices")
	v.SetDefault("minio.use_ssl", false)

	// SMTP defaults - credentials MUST be provided via environment variables if needed
	v.SetDefault("smtp.host", "")
	v.SetDefault("smtp.port", 587)
	v.SetDefault("smtp.username", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from", "")
}
