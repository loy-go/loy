package wiring

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/generator/model"
	"github.com/loy-go/loy/internal/generator/naming"
	"github.com/loy-go/loy/internal/generator/splicer"
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

// GenerateModularWiringArtifacts produces a standalone wire_<domain>.go file and registers it in wiring.go.
func GenerateModularWiringArtifacts(featureName, modulePath string) []model.Artifact {
	pascal := naming.ToPascalCase(featureName)
	pkgName := naming.ToPackageName(featureName)

	modularFileContent := fmt.Sprintf(`package app

import (
	%sRepo "%s/internal/%s/repository"
	%sService "%s/internal/%s/service"
	%sHttp "%s/internal/%s/transport/http"
)

// wire%s initializes dependencies for the %s bounded context.
func (a *App) wire%s() error {
	repo, err := %sRepo.NewPostgresRepository(a.db)
	if err != nil {
		return err
	}
	svc, err := %sService.NewService(repo)
	if err != nil {
		return err
	}
	h, err := %sHttp.NewHandler(svc)
	if err != nil {
		return err
	}
	if a.router != nil {
		h.RegisterRoutes(a.router.Group("/api/v1"))
	}
	return nil
}
`, pkgName, modulePath, pkgName,
		pkgName, modulePath, pkgName,
		pkgName, modulePath, pkgName,
		pascal, pkgName,
		pascal,
		pkgName,
		pkgName,
		pkgName,
	)

	splicedCall := fmt.Sprintf("\tif err := a.wire%s(); err != nil {\n\t\treturn err\n\t}", pascal)

	return []model.Artifact{
		{
			Path:        fmt.Sprintf("internal/app/wire_%s.go", pkgName),
			Ownership:   model.DeveloperOwned,
			Permissions: 0644,
			Content:     []byte(modularFileContent),
		},
		{
			Path:      "internal/app/wiring.go",
			Ownership: model.MixedOwned,
			Region:    "services",
			Content:   []byte(splicedCall),
		},
	}
}
