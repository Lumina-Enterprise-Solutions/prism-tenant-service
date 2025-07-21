// File: services/prism-tenant-service/internal/repository/tenant_repository.go
package repository

import (
	"context"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// BARU: DBTX adalah interface yang dipenuhi oleh *pgxpool.Pool dan pgx.Tx.
// Ini memungkinkan metode repositori menjadi fleksibel.
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type TenantRepository interface {
	Create(ctx context.Context, tx pgx.Tx, name string, domain *string) (*model.Tenant, error)
	// DIUBAH: Menggunakan DBTX agar bisa menerima pool atau transaksi.
	GetByName(ctx context.Context, db DBTX, name string) (*model.Tenant, error)
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

// GetByName adalah implementasi baru.
// DIUBAH: Menerima DBTX dan menggunakan db.QueryRow.
func (r *postgresTenantRepository) GetByName(ctx context.Context, db DBTX, name string) (*model.Tenant, error) {
	var tenant model.Tenant
	sql := `SELECT tenant_id, name, domain, created_at, updated_at FROM tenants WHERE name = $1;`
	err := db.QueryRow(ctx, sql, name).Scan(&tenant.TenantID, &tenant.Name, &tenant.Domain, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}
