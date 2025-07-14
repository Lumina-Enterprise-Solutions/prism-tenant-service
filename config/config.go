// File: services/prism-tenant-service/config/config.go
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/config"
)

type Config struct {
	Port                int
	GRPCPort            int
	ServiceName         string
	JaegerEndpoint      string
	VaultAddr           string
	VaultToken          string
	UserServiceGRPCAddr string
}

func Load() *Config {
	loader, err := config.NewLoader()
	if err != nil {
		log.Fatalf("Gagal membuat config loader: %v", err)
	}

	serviceName := "prism-tenant-service"
	pathPrefix := fmt.Sprintf("config/%s", serviceName)

	return &Config{
		Port:                loader.GetInt(fmt.Sprintf("%s/port", pathPrefix), 8080),
		GRPCPort:            loader.GetInt(fmt.Sprintf("%s/grpc_port", pathPrefix), 9003),
		ServiceName:         serviceName,
		JaegerEndpoint:      loader.Get("config/global/jaeger_endpoint", "jaeger:4317"),
		VaultAddr:           os.Getenv("VAULT_ADDR"),
		VaultToken:          os.Getenv("VAULT_TOKEN"),
		UserServiceGRPCAddr: loader.Get("config/prism-user-service/grpc_addr", "user-service:9001"),
	}
}
