package client

import (
	"context"
	"fmt"
	"log"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/model"
	userv1 "github.com/Lumina-Enterprise-Solutions/prism-protobufs/gen/go/prism/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type UserServiceClient interface {
	CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*model.User, error)
	Close()
}

type grpcUserServiceClient struct {
	client userv1.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserServiceClient(target string) (UserServiceClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("did not connect to user-service: %w", err)
	}
	client := userv1.NewUserServiceClient(conn)
	return &grpcUserServiceClient{client: client, conn: conn}, nil
}

func (c *grpcUserServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection to user-service: %v", err)
		}
	}
}

func (c *grpcUserServiceClient) CreateUser(ctx context.Context, req *userv1.CreateUserRequest) (*model.User, error) {
	// Inject tenant_id ke dalam metadata gRPC
	md := metadata.Pairs("tenant_id", req.TenantId)
	ctxWithTenant := metadata.NewOutgoingContext(ctx, md)

	res, err := c.client.CreateUser(ctxWithTenant, req)
	if err != nil {
		return nil, err
	}
	return &model.User{
		ID:           res.Id,
		Email:        res.Email,
		PasswordHash: res.PasswordHash,
		RoleName:     res.RoleName,
		Status:       res.Status,
		Is2FAEnabled: res.Is_2FaEnabled,
		TOTPSecret:   res.TotpSecret,
		TenantID:     res.TenantId,
	}, nil
}
