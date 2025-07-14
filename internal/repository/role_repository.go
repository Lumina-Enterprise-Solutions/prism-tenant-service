// File: services/prism-tenant-service/internal/repository/role_repository.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type RoleRepository interface {
	CreateDefaultRoles(ctx context.Context, tx pgx.Tx, tenantID string) (adminRoleID string, err error)
}

type postgresRoleRepository struct{}

func NewRoleRepository() RoleRepository {
	return &postgresRoleRepository{}
}

func (r *postgresRoleRepository) CreateDefaultRoles(ctx context.Context, tx pgx.Tx, tenantID string) (string, error) {
	var adminRoleID string

	// Buat peran 'admin'
	sqlAdmin := `INSERT INTO roles (name, description, tenant_id) VALUES ($1, $2, $3) RETURNING id;`
	err := tx.QueryRow(ctx, sqlAdmin, "admin", "Administrator with all permissions", tenantID).Scan(&adminRoleID)
	if err != nil {
		return "", err
	}

	// Buat peran 'user'
	sqlUser := `INSERT INTO roles (name, description, tenant_id) VALUES ($1, $2, $3);`
	_, err = tx.Exec(ctx, sqlUser, "user", "Standard user with basic permissions", tenantID)
	if err != nil {
		return "", err
	}

	// TODO: Nanti bisa ditambahkan untuk assign permission default ke peran-peran ini.

	return adminRoleID, nil
}
