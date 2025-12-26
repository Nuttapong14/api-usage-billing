package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Nuttapong14/api-usage-billing/internal/api"
	"github.com/Nuttapong14/api-usage-billing/internal/config"
	subscriptionhandler "github.com/Nuttapong14/api-usage-billing/internal/handler/subscription"
	"github.com/Nuttapong14/api-usage-billing/internal/pkg/logger"
	"github.com/Nuttapong14/api-usage-billing/internal/repository"
	subscriptionsvc "github.com/Nuttapong14/api-usage-billing/internal/service/subscription"
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

	subscriptionRepo := repository.NewSubscriptionRepository(db)
	tierRepo := repository.NewTierRepository(db)
	usageRepo := repository.NewUsageRepository(db)

	subscriptionService := subscriptionsvc.NewService(subscriptionRepo, tierRepo, usageRepo)
	tierService := subscriptionsvc.NewTierService(tierRepo, subscriptionRepo)

	serverConfig := api.DefaultServerConfig()
	serverConfig.Port = cfg.Port
	serverConfig.Environment = cfg.Environment
	serverConfig.AppName = "subscription-service"

	server := api.NewServer(serverConfig)
	server.RegisterHealthCheck()

	handler := subscriptionhandler.NewHandler(subscriptionService, tierService)
	RegisterRoutes(server.App(), handler)

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
