package astmod

import (
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/filesystem"
)

func TestFixARCH005(t *testing.T) {
	fs := filesystem.NewMemFileSystem()

	src := `package service

import (
	"context"
	"example.com/app/internal/user/repository"
)

type Service struct {
	repo *repository.PostgresRepository
}

func NewService(repo *repository.PostgresRepository) *Service {
	return &Service{repo: repo}
}
`
	filePath := "internal/user/service/service.go"
	_ = fs.MkdirAll("internal/user/service", 0755)
	_ = fs.WriteFile(filePath, []byte(src), 0644)

	err := FixARCH005(fs, filePath, "example.com/app/internal/user/repository")
	if err != nil {
		t.Fatalf("FixARCH005 failed: %v", err)
	}

	updated, err := fs.ReadFile(filePath)
	if err != nil {
		t.Fatalf("reading updated file: %v", err)
	}

	res := string(updated)
	if strings.Contains(res, "example.com/app/internal/user/repository") {
		t.Errorf("expected infrastructure import to be removed, got:\n%s", res)
	}
	if !strings.Contains(res, "repo Repository") {
		t.Errorf("expected struct field to be rewritten to interface, got:\n%s", res)
	}
	if !strings.Contains(res, "type Repository interface") {
		t.Errorf("expected consumer interface declaration, got:\n%s", res)
	}
}
