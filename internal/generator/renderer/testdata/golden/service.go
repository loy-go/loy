package service

import "context"

// AccountRepository defines repository contract.
type AccountRepository interface {
	FindByID(ctx context.Context, id string) (*Account, error)
}

// Account entity model.
type Account struct {
	ID string
}
