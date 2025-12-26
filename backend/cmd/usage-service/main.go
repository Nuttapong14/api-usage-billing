package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Nuttapong14/api-usage-billing/internal/api"
	"github.com/Nuttapong14/api-usage-billing/internal/config"
	usageevent "github.com/Nuttapong14/api-usage-billing/internal/event/usage"
	apikeyhandler "github.com/Nuttapong14/api-usage-billing/internal/handler/apikey"
	usagehandler "github.com/Nuttapong14/api-usage-billing/internal/handler/usage"
	infraredis "github.com/Nuttapong14/api-usage-billing/internal/infrastructure/redis"
	"github.com/Nuttapong14/api-usage-billing/internal/pkg/hash"
	"github.com/Nuttapong14/api-usage-billing/internal/pkg/logger"
	"github.com/Nuttapong14/api-usage-billing/internal/repository"
	apikeyservice "github.com/Nuttapong14/api-usage-billing/internal/service/apikey"
	usageservice "github.com/Nuttapong14/api-usage-billing/internal/service/usage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	appLogger := logger.New("info", cfg.Environment != "production")
	log.Logger = appLogger

	db, err := connectDatabase(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}

	redisClient, err := infraredis.NewClient(infraredis.Config{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect redis")
	}
	defer func() {
		_ = redisClient.Close()
	}()

	counter := usageservice.NewCounterStore(redisClient.Raw())
	usageRepo := repository.NewUsageRepository(db)
	usageSvc := usageservice.NewService(usageRepo, counter)

	customerRepo := repository.NewCustomerRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	apiKeyCache := apikeyservice.NewCache(redisClient.Raw())
	apiKeyPubSub := infraredis.NewPubSub(redisClient)
	if err := apikeyservice.NewRevocationSubscriber(apiKeyPubSub, apiKeyCache).Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe to api key revocations")
	}
	defer func() {
		_ = apiKeyPubSub.Close()
	}()

	keyPrefix := hash.DefaultLivePrefix
	if cfg.Environment != "production" {
		keyPrefix = hash.DefaultTestPrefix
	}
	apiKeySvc := apikeyservice.NewService(apiKeyRepo, customerRepo, apiKeyCache, apiKeyPubSub, apikeyservice.Config{
		KeyPrefix: keyPrefix,
	})

	consumer := setupUsageConsumer(redisClient.Raw(), usageSvc)
	if err := consumer.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start usage consumer")
	}
	defer func() {
		_ = consumer.Stop()
	}()

	serverConfig := api.DefaultServerConfig()
	serverConfig.Port = cfg.Port
	serverConfig.Environment = cfg.Environment
	serverConfig.AppName = "usage-service"

	server := api.NewServer(serverConfig)
	server.RegisterHealthCheck()

	handler := usagehandler.NewHandler(usageSvc)
	apiKeyHandler := apikeyhandler.NewHandler(apiKeySvc)
	RegisterRoutes(server.App(), handler, apiKeyHandler)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatal().Err(err).Msg("server stopped")
		}
	}()

	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown server")
	}
}

func connectDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func setupUsageConsumer(rdb redis.UniversalClient, svc usageservice.Service) *usageevent.Consumer {
	cfg := usageevent.DefaultConsumerConfig()
	if hostname, err := os.Hostname(); err == nil {
		cfg.ConsumerName = fmt.Sprintf("usage-%s", hostname)
	}

	consumer := usageevent.NewConsumer(rdb, cfg)
	consumer.RegisterHandler(func(ctx context.Context, event *usageevent.Event) error {
		return svc.RecordUsage(ctx, event)
	})

	return consumer
}
