package mocks

import (
	"context"

	"github.com/0x0FACED/load-balancer/internal/client"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Get(ctx context.Context, clientID string) (*client.Client, error) {
	args := m.Called(ctx, clientID)
	if cfg := args.Get(0); cfg != nil {
		return cfg.(*client.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepo) Close() error {
	return m.Called().Error(0)
}
