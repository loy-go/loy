package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
	"github.com/loy-go/loy/internal/version"
)

// ToolHandler handles tool execution.
type ToolHandler func(ctx context.Context, args json.RawMessage) (*ToolCallResult, error)

// ResourceHandler handles resource reading.
type ResourceHandler func(ctx context.Context, uri string) (*ResourceContent, error)

// Server coordinates MCP JSON-RPC 2.0 requests over standard I/O.
type Server struct {
	fs          filesystem.FileSystem
	runner      process.Runner
	projectRoot string
	in          io.Reader
	out         io.Writer

	mu        sync.RWMutex
	tools     map[string]Tool
	handlers  map[string]ToolHandler
	resources map[string]Resource
	resRead   map[string]ResourceHandler
}

// NewServer constructs an MCP server.
func NewServer(fs filesystem.FileSystem, runner process.Runner, projectRoot string, in io.Reader, out io.Writer) *Server {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if projectRoot == "" {
		projectRoot = "."
	}

	s := &Server{
		fs:          fs,
		runner:      runner,
		projectRoot: projectRoot,
		in:          in,
		out:         out,
		tools:       make(map[string]Tool),
		handlers:    make(map[string]ToolHandler),
		resources:   make(map[string]Resource),
		resRead:     make(map[string]ResourceHandler),
	}

	s.registerDefaultToolsAndResources()
	return s
}

// RegisterTool registers an executable MCP tool.
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
	s.handlers[tool.Name] = handler
}

// RegisterResource registers an accessible MCP resource.
func (s *Server) RegisterResource(res Resource, handler ResourceHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[res.URI] = res
	s.resRead[res.URI] = handler
}

// Run starts the JSON-RPC line loop reading from in and writing to out.
func (s *Server) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	// Allocate larger buffer for tool call payloads
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, ParseError, "Parse error: invalid JSON")
			continue
		}

		s.handleRequest(ctx, &req)
	}

	return scanner.Err()
}

func (s *Server) handleRequest(ctx context.Context, req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		res := InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: ServerCapabilities{
				Tools:     &ToolsCapability{},
				Resources: &ResourcesCapability{},
			},
			ServerInfo: ServerInfo{
				Name:    "loy-mcp",
				Version: version.Get().Version,
			},
		}
		s.sendResult(req.ID, res)

	case "notifications/initialized":
		// Notification has no response

	case "ping":
		s.sendResult(req.ID, map[string]string{"status": "ok"})

	case "tools/list":
		s.mu.RLock()
		var toolList []Tool
		for _, t := range s.tools {
			toolList = append(toolList, t)
		}
		s.mu.RUnlock()
		s.sendResult(req.ID, map[string]any{"tools": toolList})

	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, InvalidParams, "Invalid params for tools/call")
			return
		}

		s.mu.RLock()
		handler, exists := s.handlers[params.Name]
		s.mu.RUnlock()

		if !exists {
			s.sendError(req.ID, MethodNotFound, fmt.Sprintf("Tool %q not found", params.Name))
			return
		}

		res, err := handler(ctx, params.Arguments)
		if err != nil {
			s.sendResult(req.ID, ToolCallResult{
				Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			})
			return
		}
		s.sendResult(req.ID, res)

	case "resources/list":
		s.mu.RLock()
		var resList []Resource
		for _, r := range s.resources {
			resList = append(resList, r)
		}
		s.mu.RUnlock()
		s.sendResult(req.ID, map[string]any{"resources": resList})

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.sendError(req.ID, InvalidParams, "Invalid params for resources/read")
			return
		}

		s.mu.RLock()
		handler, exists := s.resRead[params.URI]
		s.mu.RUnlock()

		if !exists {
			s.sendError(req.ID, MethodNotFound, fmt.Sprintf("Resource %q not found", params.URI))
			return
		}

		content, err := handler(ctx, params.URI)
		if err != nil {
			s.sendError(req.ID, InternalError, fmt.Sprintf("Reading resource: %v", err))
			return
		}
		s.sendResult(req.ID, map[string]any{
			"contents": []ResourceContent{*content},
		})

	default:
		s.sendError(req.ID, MethodNotFound, fmt.Sprintf("Method %q not supported", req.Method))
	}
}

func (s *Server) sendResult(id any, result any) {
	if id == nil {
		return // Notification, no response needed
	}
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.sendJSON(resp)
}

func (s *Server) sendError(id any, code int, message string) {
	if id == nil && code != ParseError {
		return // Do not respond to notifications on error per JSON-RPC 2.0 section 4.1
	}
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	s.sendJSON(resp)
}

func (s *Server) sendJSON(val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.Marshal(val)
	if err == nil {
		_, _ = s.out.Write(append(data, '\n'))
	}
}
