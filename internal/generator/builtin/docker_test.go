package builtin_test

import (
	"context"
	"strings"
	"testing"

	"github.com/uloydev/loy/internal/generator"
	"github.com/uloydev/loy/internal/generator/builtin"
)

func TestDockerGenerator(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mod := "github.com/example/demoapp"

	t.Run("standard api docker scaffolding", func(t *testing.T) {
		gen := builtin.NewDockerGenerator(mod).
			WithCapabilities(true, true, true, false).
			WithTarget("api")

		artifacts, err := gen.Generate(ctx, generator.Input{Name: "demoapp"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(artifacts) != 2 {
			t.Fatalf("expected 2 artifacts, got %d", len(artifacts))
		}

		// Dockerfile checks
		df := string(artifacts[0].Content)
		if !strings.Contains(df, "FROM golang:1.23-alpine AS builder") {
			t.Errorf("missing builder stage in Dockerfile:\n%s", df)
		}
		if !strings.Contains(df, "FROM alpine:3.20 AS runner") {
			t.Errorf("missing runner stage in Dockerfile:\n%s", df)
		}
		if !strings.Contains(df, "USER 10001:10001") {
			t.Errorf("missing non-root user in Dockerfile:\n%s", df)
		}
		if !strings.Contains(df, "ARG TARGET=api") {
			t.Errorf("missing ARG TARGET=api in Dockerfile:\n%s", df)
		}

		// Compose checks
		dc := string(artifacts[1].Content)
		if !strings.Contains(dc, "postgres:16-alpine") {
			t.Errorf("missing postgres in docker-compose.yml:\n%s", dc)
		}
		if !strings.Contains(dc, "valkey/valkey:7-alpine") {
			t.Errorf("missing valkey in docker-compose.yml:\n%s", dc)
		}
		if !strings.Contains(dc, "condition: service_healthy") {
			t.Errorf("missing healthcheck dependency in docker-compose.yml:\n%s", dc)
		}
	})

	t.Run("fullstack with vite three-stage build", func(t *testing.T) {
		gen := builtin.NewDockerGenerator(mod).
			WithCapabilities(true, false, false, true).
			WithTarget("web")

		artifacts, err := gen.Generate(ctx, generator.Input{Name: "demoapp"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		df := string(artifacts[0].Content)
		if !strings.Contains(df, "FROM node:20-alpine AS assets") {
			t.Errorf("missing Node asset builder stage in Dockerfile:\n%s", df)
		}
		if !strings.Contains(df, "COPY --from=assets") {
			t.Errorf("missing assets copy in Dockerfile:\n%s", df)
		}
		if !strings.Contains(df, "ARG TARGET=web") {
			t.Errorf("missing ARG TARGET=web in Dockerfile:\n%s", df)
		}
	})

	t.Run("workspace monorepo context awareness", func(t *testing.T) {
		gen := builtin.NewDockerGenerator(mod).
			WithWorkspace(true)

		artifacts, err := gen.Generate(ctx, generator.Input{Name: "demoapp"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		df := string(artifacts[0].Content)
		if !strings.Contains(df, "COPY go.work go.work.sum* ./") {
			t.Errorf("missing go.work copy in workspace Dockerfile:\n%s", df)
		}
		if !strings.Contains(df, "COPY packages/ ./packages/") {
			t.Errorf("missing packages/ copy in workspace Dockerfile:\n%s", df)
		}
	})
}
