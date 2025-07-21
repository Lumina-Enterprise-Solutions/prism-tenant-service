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
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TenantService interface {
	CreateTenantWithAdmin(ctx context.Context, req *tenantv1.CreateTenantWithAdminRequest) (*tenantv1.CreateTenantWithAdminResponse, error)
	GetTenantByName(ctx context.Context, name string) (*tenantv1.Tenant, error)
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
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	tenant, err := s.tenantRepo.Create(ctx, tx, req.GetOrganizationName(), nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat tenant: %w", err)
	}

	if err := s.roleRepo.CreateDefaultRoles(ctx, tx, tenant.TenantID); err != nil {
		return nil, fmt.Errorf("gagal membuat peran default: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gagal commit transaksi pembuatan tenant: %w", err)
	}

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
		log.Error().Err(err).Str("tenant_id", tenant.TenantID).Msg("KRITIS: Gagal membuat admin untuk tenant yang baru dibuat. Diperlukan tindakan manual.")
		return nil, fmt.Errorf("gagal membuat admin user via gRPC: %w", err)
	}

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

// GetTenantByName adalah implementasi baru.
func (s *tenantService) GetTenantByName(ctx context.Context, name string) (*tenantv1.Tenant, error) {
	// DIUBAH: Memanggil repo dengan s.db (pool), bukan transaksi.
	tenant, err := s.tenantRepo.GetByName(ctx, s.db, name)
	if err != nil {
		return nil, err
	}

	resp := &tenantv1.Tenant{
		TenantId:  tenant.TenantID,
		Name:      tenant.Name,
		CreatedAt: timestamppb.New(tenant.CreatedAt),
		UpdatedAt: timestamppb.New(tenant.UpdatedAt),
	}
	// FIX: Lakukan pointer assignment secara langsung.
	if tenant.Domain != nil {
		resp.Domain = tenant.Domain
	}

	return resp, nil
}
