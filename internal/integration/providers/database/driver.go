package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DriverAdapter handles opening a database connection for migrations and queries.
type DriverAdapter interface {
	Name() string
	DriverName() string
	Open(dsn string) (*sql.DB, error)
}

// PostgresDriverAdapter implements DriverAdapter for PostgreSQL via jackc/pgx/v5/stdlib.
type PostgresDriverAdapter struct{}

func (p PostgresDriverAdapter) Name() string       { return "postgres" }
func (p PostgresDriverAdapter) DriverName() string { return "pgx" }
func (p PostgresDriverAdapter) Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}
	return db, nil
}

// DriverRegistry manages supported database drivers for migrations.
type DriverRegistry struct {
	mu      sync.RWMutex
	drivers map[string]DriverAdapter
}

// NewDriverRegistry creates a new DriverRegistry.
func NewDriverRegistry() *DriverRegistry {
	r := &DriverRegistry{
		drivers: make(map[string]DriverAdapter),
	}
	r.Register(PostgresDriverAdapter{})
	// Also register aliases 'postgresql' and 'pgx'
	r.drivers["postgresql"] = PostgresDriverAdapter{}
	r.drivers["pgx"] = PostgresDriverAdapter{}
	return r
}

// Register registers a new driver adapter.
func (r *DriverRegistry) Register(adapter DriverAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers[adapter.Name()] = adapter
}

// Get resolves a driver adapter by name.
func (r *DriverRegistry) Get(name string) (DriverAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if adapter, ok := r.drivers[name]; ok {
		return adapter, nil
	}
	return nil, fmt.Errorf("unsupported database driver: %s", name)
}
