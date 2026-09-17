package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/mcp"
	"github.com/loy-go/loy/internal/process"
)

type lineReader struct {
	lines []string
	idx   int
}

func (lr *lineReader) Read(p []byte) (n int, err error) {
	if lr.idx >= len(lr.lines) {
		return 0, io.EOF
	}
	line := lr.lines[lr.idx] + "\n"
	lr.idx++
	copy(p, line)
	return len(line), nil
}

func TestMCPServer_FullSuite(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	runner := process.NewExecRunner()

	_ = memFS.MkdirAll("/proj", 0755)
	_ = memFS.WriteFile("/proj/go.mod", []byte("module github.com/example/testmod\n\ngo 1.22\n"), 0644)
	_ = memFS.WriteFile("/proj/loy.yaml", []byte("version: 1\nproject:\n  name: testmod\n"), 0644)

	// Seed route file
	_ = memFS.MkdirAll("/proj/internal/transport/http", 0755)
	_ = memFS.WriteFile("/proj/internal/transport/http/routes.go", []byte(`package http
import "net/http"
func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", nil)
}
`), 0644)

	requests := []string{
		// 1. Initialize
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`,
		// 2. Initialized notification
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		// 3. Ping
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
		// 4. Tools list
		`{"jsonrpc":"2.0","id":3,"method":"tools/list"}`,
		// 5. Tool call: loy_check
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"loy_check","arguments":{}}}`,
		// 6. Tool call: loy_check_architecture
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"loy_check_architecture","arguments":{}}}`,
		// 7. Tool call: loy_make_crud with fields array
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"loy_make_crud","arguments":{"name":"item","fields":["title:string"]}}}`,
		// 8. Tool call: loy_inspect_routes
		`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"loy_inspect_routes","arguments":{}}}`,
		// 9. Tool call: loy_get_graph_diff
		`{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"loy_get_graph_diff","arguments":{"base_ref":"main"}}}`,
		// 10. Tool call: loy_make_migration
		`{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"loy_make_migration","arguments":{"name":"create_items_table"}}}`,
		// 11. Resources list
		`{"jsonrpc":"2.0","id":10,"method":"resources/list"}`,
		// 12. Resource read: rules catalog
		`{"jsonrpc":"2.0","id":11,"method":"resources/read","params":{"uri":"loy://rules/catalog"}}`,
		// 13. Resource read: manifest
		`{"jsonrpc":"2.0","id":12,"method":"resources/read","params":{"uri":"loy://project/manifest"}}`,
		// 14. Resource read: graph
		`{"jsonrpc":"2.0","id":13,"method":"resources/read","params":{"uri":"loy://architecture/graph"}}`,
	}

	inBuf := &lineReader{lines: requests}
	var outBuf bytes.Buffer

	server := mcp.NewServer(memFS, runner, "/proj", inBuf, &outBuf)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := server.Run(ctx)
	if err != nil {
		t.Fatalf("server run failed: %v", err)
	}

	outLines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(outLines) < 13 { // 13 requests with IDs
		t.Fatalf("expected 13 responses, got %d:\n%s", len(outLines), outBuf.String())
	}

	// 1. Initialize response
	var initResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[0]), &initResp)
	if initResp.ID != float64(1) || initResp.Error != nil {
		t.Errorf("unexpected init response: %+v", initResp)
	}

	// 2. Ping response
	var pingResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[1]), &pingResp)
	if pingResp.ID != float64(2) {
		t.Errorf("unexpected ping response: %+v", pingResp)
	}

	// 3. Tools list response
	var toolsResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[2]), &toolsResp)
	toolsData, _ := json.Marshal(toolsResp.Result)
	toolsStr := string(toolsData)
	for _, expectedTool := range []string{"loy_check", "loy_check_architecture", "loy_make_crud", "loy_inspect_routes", "loy_get_graph_diff", "loy_make_migration"} {
		if !strings.Contains(toolsStr, expectedTool) {
			t.Errorf("expected tools list to contain %s: %s", expectedTool, toolsStr)
		}
	}

	// 4. loy_check response
	var checkResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[3]), &checkResp)
	checkData, _ := json.Marshal(checkResp.Result)
	if !strings.Contains(string(checkData), "All architecture rules passed") {
		t.Errorf("expected clean architecture check: %s", string(checkData))
	}

	// 5. loy_check_architecture response
	var checkArchResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[4]), &checkArchResp)
	checkArchData, _ := json.Marshal(checkArchResp.Result)
	if !strings.Contains(string(checkArchData), "All architecture rules passed") {
		t.Errorf("expected clean architecture check: %s", string(checkArchData))
	}

	// 6. loy_make_crud response
	var crudResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[5]), &crudResp)
	crudData, _ := json.Marshal(crudResp.Result)
	if !strings.Contains(string(crudData), "Successfully generated CRUD vertical slice") {
		t.Errorf("expected successful crud generation: %s", string(crudData))
	}

	// Verify CRUD files created on disk
	exists, _ := memFS.Exists("/proj/internal/item/domain/item.go")
	if !exists {
		t.Errorf("expected internal/item/domain/item.go to be generated on disk")
	}

	// 7. loy_inspect_routes response
	var routesResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[6]), &routesResp)
	routesData, _ := json.Marshal(routesResp.Result)
	if !strings.Contains(string(routesData), "/health") {
		t.Errorf("expected routes response to contain /health: %s", string(routesData))
	}

	// 8. loy_get_graph_diff response
	var diffResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[7]), &diffResp)
	diffData, _ := json.Marshal(diffResp.Result)
	if !strings.Contains(string(diffData), "Architecture Drift Report") {
		t.Errorf("expected graph diff report: %s", string(diffData))
	}

	// 9. loy_make_migration response
	var migResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[8]), &migResp)
	migData, _ := json.Marshal(migResp.Result)
	if !strings.Contains(string(migData), "create_items_table") {
		t.Errorf("expected migration response: %s", string(migData))
	}

	// 10. Resources list response
	var resListResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[9]), &resListResp)
	resListData, _ := json.Marshal(resListResp.Result)
	for _, expectedRes := range []string{"loy://rules/catalog", "loy://project/manifest", "loy://architecture/graph"} {
		if !strings.Contains(string(resListData), expectedRes) {
			t.Errorf("expected resources list to contain %s: %s", expectedRes, string(resListData))
		}
	}

	// 11. Resource read: rules catalog
	var resReadResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[10]), &resReadResp)
	resReadData, _ := json.Marshal(resReadResp.Result)
	if !strings.Contains(string(resReadData), "ARCH-001") || !strings.Contains(string(resReadData), "ARCH-015") {
		t.Errorf("expected full rules catalog content: %s", string(resReadData))
	}

	// 12. Resource read: manifest
	var manResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[11]), &manResp)
	manData, _ := json.Marshal(manResp.Result)
	if !strings.Contains(string(manData), "testmod") {
		t.Errorf("expected manifest content: %s", string(manData))
	}

	// 13. Resource read: graph
	var graphResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[12]), &graphResp)
	graphData, _ := json.Marshal(graphResp.Result)
	if !strings.Contains(string(graphData), "mermaid") {
		t.Errorf("expected mermaid graph content: %s", string(graphData))
	}
}

func TestMCPServer_ViolationSelfHealingPrompt(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	runner := process.NewExecRunner()

	_ = memFS.MkdirAll("/badproj/internal/domain/order", 0755)
	_ = memFS.WriteFile("/badproj/go.mod", []byte("module github.com/example/badmod\n\ngo 1.22\n"), 0644)
	_ = memFS.WriteFile("/badproj/loy.yaml", []byte("version: 1\nproject:\n  name: badmod\n"), 0644)

	// Introduce layer violation: Domain imports Infrastructure
	_ = memFS.WriteFile("/badproj/internal/domain/order/order.go", []byte(`package order
import "github.com/example/badmod/internal/repository/database"
type Order struct { DB database.DB }
`), 0644)
	_ = memFS.MkdirAll("/badproj/internal/repository/database", 0755)
	_ = memFS.WriteFile("/badproj/internal/repository/database/db.go", []byte("package database\ntype DB struct{}\n"), 0644)

	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"loy_check_architecture","arguments":{}}}`,
	}

	inBuf := &lineReader{lines: requests}
	var outBuf bytes.Buffer

	server := mcp.NewServer(memFS, runner, "/badproj", inBuf, &outBuf)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = server.Run(ctx)

	outLines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(outLines) == 0 {
		t.Fatalf("expected response")
	}

	var resp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[0]), &resp)
	respData, _ := json.Marshal(resp.Result)
	respStr := string(respData)

	if !strings.Contains(respStr, "LOY-ARCH-002") {
		t.Errorf("expected LOY-ARCH-002 violation in response: %s", respStr)
	}
	if !strings.Contains(respStr, "invert_dependency") {
		t.Errorf("expected remediation action invert_dependency: %s", respStr)
	}
	if !strings.Contains(respStr, "OrderRepository") {
		t.Errorf("expected remediation prompt with OrderRepository: %s", respStr)
	}
}

func TestMCPServer_ErrorsAndEdgeCases(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	runner := process.NewExecRunner()
	_ = memFS.MkdirAll("/empty", 0755)

	requests := []string{
		// 1. Invalid JSON parse error
		`{invalid-json}`,
		// 2. Unknown method
		`{"jsonrpc":"2.0","id":1,"method":"unknown/method"}`,
		// 3. Tool not found
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"non_existent_tool","arguments":{}}}`,
		// 4. Invalid params for tools/call
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":"invalid"}`,
		// 5. Resource not found
		`{"jsonrpc":"2.0","id":4,"method":"resources/read","params":{"uri":"loy://unknown/resource"}}`,
		// 6. Invalid params for resources/read
		`{"jsonrpc":"2.0","id":5,"method":"resources/read","params":"invalid"}`,
		// 7. loy_make_crud missing name
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"loy_make_crud","arguments":{"name":""}}}`,
		// 8. loy_make_migration missing name
		`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"loy_make_migration","arguments":{"name":""}}}`,
		// 9. loy_inspect_routes on empty project
		`{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"loy_inspect_routes","arguments":{}}}`,
		// 10. Reading loy://project/manifest when missing
		`{"jsonrpc":"2.0","id":9,"method":"resources/read","params":{"uri":"loy://project/manifest"}}`,
	}

	inBuf := &lineReader{lines: requests}
	var outBuf bytes.Buffer

	server := mcp.NewServer(memFS, runner, "/empty", inBuf, &outBuf)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = server.Run(ctx)

	outLines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(outLines) < 10 {
		t.Fatalf("expected 10 response lines, got %d:\n%s", len(outLines), outBuf.String())
	}

	// 1. Parse error
	var parseResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[0]), &parseResp)
	if parseResp.Error == nil || parseResp.Error.Code != mcp.ParseError {
		t.Errorf("expected ParseError, got: %+v", parseResp)
	}

	// 2. Method not found
	var mNotFoundResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[1]), &mNotFoundResp)
	if mNotFoundResp.Error == nil || mNotFoundResp.Error.Code != mcp.MethodNotFound {
		t.Errorf("expected MethodNotFound, got: %+v", mNotFoundResp)
	}

	// 3. Tool not found
	var tNotFoundResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[2]), &tNotFoundResp)
	if tNotFoundResp.Error == nil || tNotFoundResp.Error.Code != mcp.MethodNotFound {
		t.Errorf("expected MethodNotFound for tool, got: %+v", tNotFoundResp)
	}

	// 4. Invalid params for tools/call
	var invParamResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[3]), &invParamResp)
	if invParamResp.Error == nil || invParamResp.Error.Code != mcp.InvalidParams {
		t.Errorf("expected InvalidParams for tools/call, got: %+v", invParamResp)
	}

	// 5. Resource not found
	var rNotFoundResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[4]), &rNotFoundResp)
	if rNotFoundResp.Error == nil || rNotFoundResp.Error.Code != mcp.MethodNotFound {
		t.Errorf("expected MethodNotFound for resource, got: %+v", rNotFoundResp)
	}

	// 6. Invalid params for resources/read
	var invResParamResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[5]), &invResParamResp)
	if invResParamResp.Error == nil || invResParamResp.Error.Code != mcp.InvalidParams {
		t.Errorf("expected InvalidParams for resources/read, got: %+v", invResParamResp)
	}

	// 7. loy_make_crud missing name isError
	var crudErrResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[6]), &crudErrResp)
	crudErrData, _ := json.Marshal(crudErrResp.Result)
	if !strings.Contains(string(crudErrData), "isError") && !strings.Contains(string(crudErrData), "name is required") {
		t.Errorf("expected isError for missing name in loy_make_crud: %+v", crudErrResp)
	}

	// 8. loy_make_migration missing name isError
	var migErrResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[7]), &migErrResp)
	migErrData, _ := json.Marshal(migErrResp.Result)
	if !strings.Contains(string(migErrData), "isError") && !strings.Contains(string(migErrData), "name is required") {
		t.Errorf("expected isError for missing name in loy_make_migration: %+v", migErrResp)
	}

	// 9. loy_inspect_routes empty
	var routesEmptyResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[8]), &routesEmptyResp)
	routesEmptyData, _ := json.Marshal(routesEmptyResp.Result)
	if !strings.Contains(string(routesEmptyData), "No registered HTTP or WebSocket route bindings") {
		t.Errorf("expected empty routes message: %s", string(routesEmptyData))
	}

	// 10. Missing manifest error
	var manErrResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[9]), &manErrResp)
	if manErrResp.Error == nil || manErrResp.Error.Code != mcp.InternalError {
		t.Errorf("expected InternalError for missing manifest: %+v", manErrResp)
	}
}

func TestMCPServer_ASTAndPlanPreviewTools(t *testing.T) {
	memFS := filesystem.NewMemFileSystem()
	runner := process.NewExecRunner()

	_ = memFS.MkdirAll("/proj", 0755)
	_ = memFS.WriteFile("/proj/go.mod", []byte("module github.com/example/asttest\n\ngo 1.22\n"), 0644)
	_ = memFS.WriteFile("/proj/loy.yaml", []byte("version: 1\nproject:\n  name: asttest\n"), 0644)

	// Seed struct file for ast insertion
	_ = memFS.MkdirAll("/proj/internal/user/model", 0755)
	_ = memFS.WriteFile("/proj/internal/user/model/user.go", []byte(`package model

// User entity
type User struct {
	ID int64 `+"`json:\"id\"`"+`
}
`), 0644)

	// Seed route file for ast add_route
	_ = memFS.MkdirAll("/proj/internal/user/transport/http", 0755)
	_ = memFS.WriteFile("/proj/internal/user/transport/http/routes.go", []byte(`package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router) {
	router.Get("/users", nil)
}
`), 0644)

	// Seed wiring file for ast bind_dependency
	_ = memFS.MkdirAll("/proj/internal/app", 0755)
	_ = memFS.WriteFile("/proj/internal/app/wiring.go", []byte(`package app

func WireApp(a *App) error {
	return nil
}
`), 0644)

	requests := []string{
		// 1. loy_ast_insert_field
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"loy_ast_insert_field","arguments":{"file":"internal/user/model/user.go","struct_name":"User","field_name":"Email","field_type":"string","tags":"json:\"email\""}}}`,
		// 2. loy_ast_add_route
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"loy_ast_add_route","arguments":{"file":"internal/user/transport/http/routes.go","method":"POST","path":"/users","handler":"h.Create"}}}`,
		// 3. loy_ast_bind_dependency
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"loy_ast_bind_dependency","arguments":{"file":"internal/app/wiring.go","provider_func":"userRepo.NewPostgresRepository(a.db)","dep_name":"userRepo"}}}`,
		// 4. loy_plan_preview
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"loy_plan_preview","arguments":{"generator":"crud","name":"product","fields":"title:string price:float"}}}`,
	}

	inBuf := &lineReader{lines: requests}
	var outBuf bytes.Buffer

	server := mcp.NewServer(memFS, runner, "/proj", inBuf, &outBuf)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Run(ctx)
	if err != nil {
		t.Fatalf("server run failed: %v", err)
	}

	outLines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(outLines) != 4 {
		t.Fatalf("expected 4 responses, got %d:\n%s", len(outLines), outBuf.String())
	}

	// 1. Verify loy_ast_insert_field
	userFile, err := memFS.ReadFile("/proj/internal/user/model/user.go")
	if err != nil || !strings.Contains(string(userFile), "Email string `json:\"email\"`") {
		t.Errorf("expected Email field inserted in user.go:\n%s", string(userFile))
	}

	// 2. Verify loy_ast_add_route
	routeFile, err := memFS.ReadFile("/proj/internal/user/transport/http/routes.go")
	if err != nil || !strings.Contains(string(routeFile), `router.Post("/users", h.Create)`) {
		t.Errorf("expected route inserted in routes.go:\n%s", string(routeFile))
	}

	// 3. Verify loy_ast_bind_dependency
	wiringFile, err := memFS.ReadFile("/proj/internal/app/wiring.go")
	if err != nil || !strings.Contains(string(wiringFile), "userRepo, _ := userRepo.NewPostgresRepository(a.db)") {
		t.Errorf("expected dependency bound in wiring.go:\n%s", string(wiringFile))
	}

	// 4. Verify loy_plan_preview
	var previewResp mcp.JSONRPCResponse
	_ = json.Unmarshal([]byte(outLines[3]), &previewResp)
	previewData, _ := json.Marshal(previewResp.Result)
	if !strings.Contains(string(previewData), "ready") || !strings.Contains(string(previewData), "product") {
		t.Errorf("expected preview response containing product artifacts and ready status, got: %s", string(previewData))
	}
}
