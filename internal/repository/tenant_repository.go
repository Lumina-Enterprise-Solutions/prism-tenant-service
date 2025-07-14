// File: services/prism-tenant-service/internal/repository/tenant_repository.go
package repository

import (
	"context"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/model"
	"github.com/jackc/pgx/v5"
)

type TenantRepository interface {
	Create(ctx context.Context, tx pgx.Tx, name string, domain *string) (*model.Tenant, error)
}

type postgresTenantRepository struct{}

func NewTenantRepository() TenantRepository {
	return &postgresTenantRepository{}
}

func (r *postgresTenantRepository) Create(ctx context.Context, tx pgx.Tx, name string, domain *string) (*model.Tenant, error) {
	var tenant model.Tenant
	sql := `INSERT INTO tenants (name, domain) VALUES ($1, $2) RETURNING tenant_id, name, domain, created_at, updated_at;`
	err := tx.QueryRow(ctx, sql, name, domain).Scan(&tenant.TenantID, &tenant.Name, &tenant.Domain, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}
