package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/diagnostics"
	"github.com/loy-go/loy/internal/discovery"
	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/mcp"
	"github.com/loy-go/loy/internal/process"
)

func newMCPCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "mcp [path]",
		Short: "Run the Model Context Protocol (MCP) server over standard I/O",
		Long: `loy mcp starts a Model Context Protocol JSON-RPC 2.0 server over stdin/stdout.
External AI agents (Claude Desktop, Cursor, Kilo, Windsurf) can connect to this server
to discover architectural rules, run AST validations, and scaffold Clean Architecture slices.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}
			ctx := cmd.Context()

			disc, err := discovery.NewDiscoverer(fs, runner)
			if err == nil {
				if discRes, diag := disc.Discover(ctx, targetDir); diag == nil && discRes != nil {
					targetDir = discRes.RootDir
				}
			}

			server := mcp.NewServer(fs, runner, targetDir, cmd.InOrStdin(), cmd.OutOrStdout())
			if err := server.Run(ctx); err != nil {
				return &CommandError{
					Code: 1,
					Diagnostics: []*diagnostics.Diagnostic{{
						Severity: diagnostics.SeverityError,
						Code:     diagnostics.CodeInternalError,
						Message:  fmt.Sprintf("mcp server error: %v", err),
					}},
				}
			}
			return nil
		},
	}
}
