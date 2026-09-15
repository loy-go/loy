package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"loymart/internal/organization/domain"
	"loymart/internal/organization/service"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) FindByID(ctx context.Context, id int64) (*domain.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *mockRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Organization, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Organization), args.Error(1)
}

func (m *mockRepository) Create(ctx context.Context, entity *domain.Organization) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *mockRepository) Update(ctx context.Context, entity *domain.Organization) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *mockRepository) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func TestOrganizationService_GetByID(t *testing.T) {
	repo := new(mockRepository)
	svc, err := service.NewService(repo)
	assert.NoError(t, err)

	expected := &domain.Organization{ID: 1}
	repo.On("FindByID", mock.Anything, int64(1)).Return(expected, nil)

	res, err := svc.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	repo.AssertExpectations(t)
}
