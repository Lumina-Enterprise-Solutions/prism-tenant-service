// File: services/prism-tenant-service/main.go (LENGKAP)
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/client"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/enhanced_logger"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/telemetry"
	tenantv1 "github.com/Lumina-Enterprise-Solutions/prism-protobufs/gen/go/prism/tenant/v1"
	"github.com/Lumina-Enterprise-Solutions/prism-tenant-service/config"
	tenantclient "github.com/Lumina-Enterprise-Solutions/prism-tenant-service/internal/client"
	"github.com/Lumina-Enterprise-Solutions/prism-tenant-service/internal/grpc_server"
	"github.com/Lumina-Enterprise-Solutions/prism-tenant-service/internal/repository"
	"github.com/Lumina-Enterprise-Solutions/prism-tenant-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func setupDependencies(cfg *config.Config) (*pgxpool.Pool, error) {
	vaultClient, err := client.NewVaultClient(cfg.VaultAddr, cfg.VaultToken)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat klien Vault: %w", err)
	}

	secretPath := "secret/data/prism"
	if err := vaultClient.LoadSecretsToEnv(secretPath, "database_url"); err != nil {
		return nil, fmt.Errorf("gagal memuat database_url dari Vault: %w", err)
	}

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat connection pool: %w", err)
	}
	return dbpool, nil
}

func main() {
	enhanced_logger.Init()
	cfg := config.Load()
	serviceLogger := enhanced_logger.WithService(cfg.ServiceName)
	enhanced_logger.LogStartup(cfg.ServiceName, cfg.Port, map[string]interface{}{
		"grpc_port": cfg.GRPCPort,
	})

	tp, err := telemetry.InitTracerProvider(cfg.ServiceName, cfg.JaegerEndpoint)
	if err != nil {
		serviceLogger.Fatal().Err(err).Msg("Gagal menginisialisasi tracer")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			serviceLogger.Error().Err(err).Msg("Gagal mematikan tracer provider")
		}
	}()

	dbpool, err := setupDependencies(cfg)
	if err != nil {
		serviceLogger.Fatal().Err(err).Msg("Gagal menginisialisasi dependensi")
	}
	defer dbpool.Close()

	userServiceClient, err := tenantclient.NewUserServiceClient(cfg.UserServiceGRPCAddr)
	if err != nil {
		serviceLogger.Fatal().Err(err).Msg("Gagal membuat user service client")
	}
	defer userServiceClient.Close()

	tenantRepo := repository.NewTenantRepository()
	roleRepo := repository.NewRoleRepository()
	tenantService := service.NewTenantService(dbpool, tenantRepo, roleRepo, userServiceClient)
	tenantGrpcServer := grpc_server.NewTenantServer(tenantService)

	// Setup gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		serviceLogger.Fatal().Err(err).Msgf("Gagal listen di port gRPC %d", cfg.GRPCPort)
	}
	grpcServer := grpc.NewServer()
	tenantv1.RegisterTenantServiceServer(grpcServer, tenantGrpcServer)
	go func() {
		serviceLogger.Info().Int("port", cfg.GRPCPort).Msg("Memulai gRPC server...")
		if err := grpcServer.Serve(lis); err != nil {
			serviceLogger.Fatal().Err(err).Msg("gRPC server gagal berjalan")
		}
	}()

	// Setup HTTP Server (hanya untuk health check)
	router := gin.New()
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "healthy"}) })

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: router,
	}
	go func() {
		serviceLogger.Info().Int("port", cfg.Port).Msg("Memulai HTTP server untuk health check...")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serviceLogger.Fatal().Err(err).Msg("Server HTTP gagal berjalan")
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	serviceLogger.Info().Msg("Memulai graceful shutdown...")
	grpcServer.GracefulStop()
	serviceLogger.Info().Msg("gRPC server berhenti.")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		serviceLogger.Fatal().Err(err).Msg("Server HTTP terpaksa dimatikan")
	}
	enhanced_logger.LogShutdown(cfg.ServiceName)
}
