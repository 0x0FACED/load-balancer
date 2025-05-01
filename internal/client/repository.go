package client

import "context"

type Repository interface {
	Create(ctx context.Context, cfg Client) error
	Get(ctx context.Context, id string) (*Client, error)
	Update(ctx context.Context, cfg Client) error
	Delete(ctx context.Context, id string) error
}
