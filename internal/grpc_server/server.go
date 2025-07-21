package grpc_server

import (
	"context"
	"errors"

	tenantv1 "github.com/Lumina-Enterprise-Solutions/prism-protobufs/gen/go/prism/tenant/v1"
	"github.com/Lumina-Enterprise-Solutions/prism-tenant-service/internal/service"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TenantServer struct {
	tenantv1.UnimplementedTenantServiceServer
	tenantService service.TenantService
}

func NewTenantServer(svc service.TenantService) *TenantServer {
	return &TenantServer{tenantService: svc}
}

func (s *TenantServer) CreateTenantWithAdmin(ctx context.Context, req *tenantv1.CreateTenantWithAdminRequest) (*tenantv1.CreateTenantWithAdminResponse, error) {
	response, err := s.tenantService.CreateTenantWithAdmin(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "gagal memproses pembuatan tenant dan admin: %v", err)
	}
	return response, nil
}

// Implementasi handler untuk RPC baru
func (s *TenantServer) GetTenantByName(ctx context.Context, req *tenantv1.GetTenantByNameRequest) (*tenantv1.Tenant, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "nama organisasi wajib diisi")
	}

	tenant, err := s.tenantService.GetTenantByName(ctx, req.GetName())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "tenant dengan nama '%s' tidak ditemukan", req.GetName())
		}
		return nil, status.Errorf(codes.Internal, "gagal mengambil tenant: %v", err)
	}
	return tenant, nil
}
