// File: services/prism-tenant-service/internal/grpc_server/server.go
package grpc_server

import (
	"context"

	tenantv1 "github.com/Lumina-Enterprise-Solutions/prism-protobufs/gen/go/prism/tenant/v1"
	"github.com/Lumina-Enterprise-Solutions/prism-tenant-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TenantServer adalah implementasi dari TenantServiceServer yang dihasilkan oleh protobuf.
type TenantServer struct {
	tenantv1.UnimplementedTenantServiceServer
	tenantService service.TenantService
}

// NewTenantServer membuat instance baru dari gRPC server kita.
func NewTenantServer(svc service.TenantService) *TenantServer {
	return &TenantServer{tenantService: svc}
}

// CreateTenantWithAdmin adalah implementasi dari RPC yang kita definisikan di .proto.
func (s *TenantServer) CreateTenantWithAdmin(ctx context.Context, req *tenantv1.CreateTenantWithAdminRequest) (*tenantv1.CreateTenantWithAdminResponse, error) {
	// Delegasikan logika bisnis yang kompleks ke service layer.
	// Handler/server gRPC hanya bertanggung jawab untuk validasi input dasar dan
	// memetakan antara DTO (protobuf) dan model internal.
	response, err := s.tenantService.CreateTenantWithAdmin(ctx, req)
	if err != nil {
		// Terjemahkan error dari service layer ke status gRPC yang sesuai.
		// Ini adalah praktik yang baik untuk memberikan feedback yang jelas ke client.
		// TODO: Tambahkan penanganan error yang lebih spesifik, misalnya jika tenant sudah ada.
		return nil, status.Errorf(codes.Internal, "gagal memproses pembuatan tenant dan admin: %v", err)
	}

	return response, nil
}
