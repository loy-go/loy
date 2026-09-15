package service

import (
	"context"
	"fmt"

	"loymart/internal/notification/domain"
)

// Service coordinates application use cases for notifications.
type Service struct {
	repo domain.Repository
}

// NewService constructs a new application Service.
func NewService(repo domain.Repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository dependency is required")
	}
	return &Service{repo: repo}, nil
}

// GetByID returns the Notification entity matching the given ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*domain.Notification, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns a list of Notification entities.
func (s *Service) List(ctx context.Context, limit, offset int) ([]*domain.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.FindAll(ctx, limit, offset)
}

// Create handles the creation of a new Notification.
func (s *Service) Create(ctx context.Context, entity *domain.Notification) error {
	if entity == nil {
		return fmt.Errorf("entity cannot be nil")
	}
	return s.repo.Create(ctx, entity)
}

// Update modifies an existing Notification.
func (s *Service) Update(ctx context.Context, entity *domain.Notification) error {
	if entity == nil {
		return fmt.Errorf("entity cannot be nil")
	}
	return s.repo.Update(ctx, entity)
}

// Delete removes a Notification by ID.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
