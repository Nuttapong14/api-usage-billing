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

	"github.com/your-org/api-usage-billing/backend/internal/api"
	"github.com/your-org/api-usage-billing/backend/internal/config"
	billinghandler "github.com/your-org/api-usage-billing/backend/internal/handler/billing"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/logger"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/pdf"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/storage"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
	billingsvc "github.com/your-org/api-usage-billing/backend/internal/service/billing"
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

	invoiceRepo := repository.NewInvoiceRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	tierRepo := repository.NewTierRepository(db)
	usageRepo := repository.NewUsageRepository(db)

	pdfGenerator, err := pdf.NewTemplateGenerator()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load invoice templates")
	}

	var uploader storage.Uploader
	if cfg.MinIO.Endpoint != "" {
		uploader, err = storage.NewMinioUploader(storage.MinioConfig{
			Endpoint:  cfg.MinIO.Endpoint,
			AccessKey: cfg.MinIO.AccessKey,
			SecretKey: cfg.MinIO.SecretKey,
			Bucket:    cfg.MinIO.Bucket,
			UseSSL:    cfg.MinIO.UseSSL,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to connect MinIO")
		}
	}

	billingService := billingsvc.NewService(invoiceRepo, paymentRepo, subscriptionRepo, tierRepo, usageRepo, pdfGenerator, uploader)

	serverConfig := api.DefaultServerConfig()
	serverConfig.Port = cfg.Port
	serverConfig.Environment = cfg.Environment
	serverConfig.AppName = "billing-service"

	server := api.NewServer(serverConfig)
	server.RegisterHealthCheck()

	handler := billinghandler.NewHandler(billingService)
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
