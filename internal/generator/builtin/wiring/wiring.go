package wiring

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/uloydev/loy/internal/filesystem"
	"github.com/uloydev/loy/internal/generator/model"
	"github.com/uloydev/loy/internal/generator/naming"
	"github.com/uloydev/loy/internal/generator/splicer"
)

const DefaultWiringTemplate = `package app

import (
	// loy:region:imports
	// loy:endregion
)

// wireDependencies sets up repositories, services, handlers and registers routes.
func (a *App) wireDependencies() error {
	// loy:region:repositories
	// loy:endregion

	// loy:region:services
	// loy:endregion

	// loy:region:handlers
	// loy:endregion

	// loy:region:routes
	// loy:endregion

	return nil
}
`

// SplicerManager manages splicing feature registration into internal/app/wiring.go.
type SplicerManager struct {
	fs      filesystem.FileSystem
	splicer *splicer.Splicer
}

// NewSplicerManager constructs a SplicerManager.
func NewSplicerManager(fs filesystem.FileSystem) *SplicerManager {
	return &SplicerManager{
		fs:      fs,
		splicer: splicer.New(),
	}
}

// EnsureWiringFile guarantees internal/app/wiring.go exists with valid regions.
func (m *SplicerManager) EnsureWiringFile(ctx context.Context, rootDir string) error {
	wiringPath, err := filesystem.CleanAndValidatePath(rootDir, filepath.Join(rootDir, "internal/app/wiring.go"))
	if err != nil {
		return err
	}

	exists, err := m.fs.Exists(wiringPath)
	if err != nil {
		return err
	}
	if !exists {
		// Ensure parent dir
		parent := filepath.Dir(wiringPath)
		if err := m.fs.MkdirAll(parent, 0755); err != nil {
			return err
		}
		return m.fs.WriteFile(wiringPath, []byte(DefaultWiringTemplate), 0644)
	}
	return nil
}

// GenerateWiringArtifacts produces managed region artifacts for a feature.
func GenerateWiringArtifacts(featureName, modulePath string) []model.Artifact {
	camel := naming.ToCamelCase(featureName)
	pkgName := naming.ToPackageName(featureName)

	importsCode := fmt.Sprintf("\t%sRepo \"%s/internal/%s/repository\"\n\t%sService \"%s/internal/%s/service\"\n\t%sHttp \"%s/internal/%s/transport/http\"",
		pkgName, modulePath, pkgName,
		pkgName, modulePath, pkgName,
		pkgName, modulePath, pkgName,
	)

	repoCode := fmt.Sprintf("\t%sRepo, _ := %sRepo.NewPostgresRepository(a.db)\n\t_ = %sRepo", camel, pkgName, camel)
	serviceCode := fmt.Sprintf("\t%sSvc, _ := %sService.NewService(%sRepo)\n\t_ = %sSvc", camel, pkgName, camel, camel)
	handlerCode := fmt.Sprintf("\t%sHandler, _ := %sHttp.NewHandler(%sSvc)", camel, pkgName, camel)
	routeCode := fmt.Sprintf("\t%sHandler.RegisterRoutes(a.router.Group(\"/api/v1\"))", camel)

	return []model.Artifact{
		{
			Path:      "internal/app/wiring.go",
			Ownership: model.MixedOwned,
			Region:    "imports",
			Content:   []byte(importsCode),
		},
		{
			Path:      "internal/app/wiring.go",
			Ownership: model.MixedOwned,
			Region:    "repositories",
			Content:   []byte(repoCode),
		},
		{
			Path:      "internal/app/wiring.go",
			Ownership: model.MixedOwned,
			Region:    "services",
			Content:   []byte(serviceCode),
		},
		{
			Path:      "internal/app/wiring.go",
			Ownership: model.MixedOwned,
			Region:    "handlers",
			Content:   []byte(handlerCode),
		},
		{
			Path:      "internal/app/wiring.go",
			Ownership: model.MixedOwned,
			Region:    "routes",
			Content:   []byte(routeCode),
		},
	}
}
