package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
)

type dummyRunner struct{}

func (d *dummyRunner) Run(ctx context.Context, dir string, name string, args ...string) (*process.Result, error) {
	return &process.Result{ExitCode: 0}, nil
}

func (d *dummyRunner) RunWithInput(ctx context.Context, dir string, in io.Reader, name string, args ...string) (*process.Result, error) {
	return &process.Result{ExitCode: 0}, nil
}

func TestCheckCmd(t *testing.T) {
	runner := &dummyRunner{}

	t.Run("check clean project passes", func(t *testing.T) {
		memFS := filesystem.NewMemFileSystem()
		_ = memFS.MkdirAll("/proj", 0755)
		_ = memFS.WriteFile("/proj/go.mod", []byte("module github.com/test/clean\n\ngo 1.22\n"), 0644)
		_ = memFS.WriteFile("/proj/loy.yaml", []byte("version: 1\nproject:\n  name: clean\n  module: github.com/test/clean\n"), 0644)

		_ = memFS.MkdirAll("/proj/internal/domain/user", 0755)
		_ = memFS.WriteFile("/proj/internal/domain/user/user.go", []byte("package user\ntype User struct{}\n"), 0644)

		cmd := newCheckCmd(memFS, runner)
		cmd.SetContext(WithOptions(context.Background(), &GlobalOptions{Quiet: true}))
		cmd.SetArgs([]string{"/proj"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("expected check to pass, got: %v", err)
		}
	})

	t.Run("check project with layer violation fails", func(t *testing.T) {
		memFS := filesystem.NewMemFileSystem()
		_ = memFS.MkdirAll("/badproj", 0755)
		_ = memFS.WriteFile("/badproj/go.mod", []byte("module github.com/test/bad\n\ngo 1.22\n"), 0644)
		_ = memFS.WriteFile("/badproj/loy.yaml", []byte("version: 1\nproject:\n  name: bad\n  module: github.com/test/bad\n"), 0644)

		_ = memFS.MkdirAll("/badproj/internal/domain/user", 0755)
		_ = memFS.WriteFile("/badproj/internal/domain/user/user.go", []byte(`package user
import "github.com/test/bad/internal/repository/pg"
type User struct { R pg.Repo }
`), 0644)

		_ = memFS.MkdirAll("/badproj/internal/repository/pg", 0755)
		_ = memFS.WriteFile("/badproj/internal/repository/pg/pg.go", []byte("package pg\ntype Repo struct{}\n"), 0644)

		cmd := newCheckCmd(memFS, runner)
		cmd.SetContext(WithOptions(context.Background(), &GlobalOptions{Quiet: true}))
		cmd.SetArgs([]string{"/badproj"})

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected check to fail with layer violation error, but passed")
		}

		cmdErr, ok := err.(*CommandError)
		if !ok || cmdErr.Code != 1 {
			t.Fatalf("expected CommandError with exit code 1, got %v", err)
		}
	})

	t.Run("check with format github outputs annotations", func(t *testing.T) {
		memFS := filesystem.NewMemFileSystem()
		runner := &dummyRunner{}

		_ = memFS.MkdirAll("/ghproj/internal/domain/user", 0755)
		_ = memFS.WriteFile("/ghproj/go.mod", []byte("module github.com/test/gh\n\ngo 1.22\n"), 0644)
		_ = memFS.WriteFile("/ghproj/internal/domain/user/user.go", []byte(`package user
import "github.com/test/gh/internal/repository/pg"
type User struct { R pg.Repo }
`), 0644)
		_ = memFS.MkdirAll("/ghproj/internal/repository/pg", 0755)
		_ = memFS.WriteFile("/ghproj/internal/repository/pg/pg.go", []byte("package pg\ntype Repo struct{}\n"), 0644)

		outBuf := new(bytes.Buffer)
		cmd := newCheckCmd(memFS, runner)
		cmd.SetOut(outBuf)
		cmd.SetContext(WithOptions(context.Background(), &GlobalOptions{Quiet: false}))
		cmd.SetArgs([]string{"/ghproj", "--format", "github"})

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected check to fail")
		}

		out := outBuf.String()
		if !strings.Contains(out, "::error") || !strings.Contains(out, "title=LOY-ARCH-002") {
			t.Errorf("expected ::error annotation with title in output, got: %s", out)
		}
	})

	t.Run("check with format agent outputs self-healing prompt payload", func(t *testing.T) {
		memFS := filesystem.NewMemFileSystem()
		runner := &dummyRunner{}

		_ = memFS.MkdirAll("/agentproj/internal/domain/order", 0755)
		_ = memFS.WriteFile("/agentproj/go.mod", []byte("module github.com/test/agent\n\ngo 1.22\n"), 0644)
		_ = memFS.WriteFile("/agentproj/internal/domain/order/order.go", []byte(`package order
import "github.com/test/agent/internal/repository/pg"
type Order struct { R pg.Repo }
`), 0644)
		_ = memFS.MkdirAll("/agentproj/internal/repository/pg", 0755)
		_ = memFS.WriteFile("/agentproj/internal/repository/pg/pg.go", []byte("package pg\ntype Repo struct{}\n"), 0644)

		outBuf := new(bytes.Buffer)
		cmd := newCheckCmd(memFS, runner)
		cmd.SetOut(outBuf)
		cmd.SetContext(WithOptions(context.Background(), &GlobalOptions{Quiet: false}))
		cmd.SetArgs([]string{"/agentproj", "--format", "agent"})

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected check to fail with layer violation error")
		}

		out := outBuf.String()
		if !strings.Contains(out, "LOY-ARCH-002") {
			t.Errorf("expected LOY-ARCH-002 in agent output, got: %s", out)
		}
		if !strings.Contains(out, "invert_dependency") {
			t.Errorf("expected invert_dependency action in agent output, got: %s", out)
		}
		if !strings.Contains(out, "OrderRepository") {
			t.Errorf("expected OrderRepository in prompt, got: %s", out)
		}

		// Also test clean project with format agent
		cleanBuf := new(bytes.Buffer)
		cmdClean := newCheckCmd(memFS, runner)
		cmdClean.SetOut(cleanBuf)
		cmdClean.SetContext(WithOptions(context.Background(), &GlobalOptions{Quiet: false}))
		_ = memFS.MkdirAll("/cleanproj", 0755)
		_ = memFS.WriteFile("/cleanproj/go.mod", []byte("module github.com/test/cleanagent\n\ngo 1.22\n"), 0644)
		cmdClean.SetArgs([]string{"/cleanproj", "--format", "agent"})

		cleanErr := cmdClean.Execute()
		if cleanErr != nil {
			t.Fatalf("expected clean check to pass, got: %v", cleanErr)
		}
		if strings.TrimSpace(cleanBuf.String()) != "[]" {
			t.Errorf("expected empty json array for clean project, got: %q", cleanBuf.String())
		}
	})
}
