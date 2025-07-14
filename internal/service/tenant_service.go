// File: services/prism-tenant-service/internal/service/tenant_service.go
package service

import (
	"context"
	"fmt"

	tenantv1 "github.com/Lumina-Enterprise-Solutions/prism-protobufs/gen/go/prism/tenant/v1"
	userv1 "github.com/Lumina-Enterprise-Solutions/prism-protobufs/gen/go/prism/user/v1"
	"github.com/Lumina-Enterprise-Solutions/prism-tenant-service/internal/client"
	"github.com/Lumina-Enterprise-Solutions/prism-tenant-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantService interface {
	CreateTenantWithAdmin(ctx context.Context, req *tenantv1.CreateTenantWithAdminRequest) (*tenantv1.CreateTenantWithAdminResponse, error)
}

type tenantService struct {
	db                *pgxpool.Pool
	tenantRepo        repository.TenantRepository
	roleRepo          repository.RoleRepository
	userServiceClient client.UserServiceClient
}

func NewTenantService(db *pgxpool.Pool, tenantRepo repository.TenantRepository, roleRepo repository.RoleRepository, userClient client.UserServiceClient) TenantService {
	return &tenantService{
		db:                db,
		tenantRepo:        tenantRepo,
		roleRepo:          roleRepo,
		userServiceClient: userClient,
	}
}

func (s *tenantService) CreateTenantWithAdmin(ctx context.Context, req *tenantv1.CreateTenantWithAdminRequest) (*tenantv1.CreateTenantWithAdminResponse, error) {
	// 1. Mulai transaksi database
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx) // Pastikan rollback jika terjadi error

	// 2. Buat tenant baru
	tenant, err := s.tenantRepo.Create(ctx, tx, req.GetOrganizationName(), nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat tenant: %w", err)
	}

	// 4. Panggil user-service untuk membuat admin user
	createUserReq := &userv1.CreateUserRequest{
		Email:     req.GetAdminEmail(),
		Password:  req.GetPasswordHash(),
		FirstName: req.GetAdminFirstName(),
		LastName:  req.GetAdminLastName(),
		Role:      "admin",
		TenantId:  tenant.TenantID,
	}
	createdUser, err := s.userServiceClient.CreateUser(ctx, createUserReq)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat admin user via gRPC: %w", err)
	}

	// 5. Jika semua berhasil, commit transaksi
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal commit transaksi: %w", err)
	}

	// 6. Siapkan response
	response := &tenantv1.CreateTenantWithAdminResponse{
		TenantId: tenant.TenantID,
		UserId:   createdUser.ID,
		AdminUser: &userv1.UserAuthDetailsResponse{
			Id:            createdUser.ID,
			Email:         createdUser.Email,
			PasswordHash:  createdUser.PasswordHash,
			RoleName:      "admin",
			Status:        createdUser.Status,
			Is_2FaEnabled: createdUser.Is2FAEnabled,
			TotpSecret:    createdUser.TOTPSecret,
			TenantId:      createdUser.TenantID,
		},
	}

	return response, nil
}
