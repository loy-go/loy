package cli

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/loy-go/loy/internal/filesystem"
	"github.com/loy-go/loy/internal/process"
	"github.com/loy-go/loy/internal/route"
	"github.com/spf13/cobra"
)

// RouteInfo models a discovered application endpoint (aliased from route.RouteInfo).
type RouteInfo = route.RouteInfo

func newRoutesCmd(fs filesystem.FileSystem, runner process.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "routes [path]",
		Short: "Inspect and list all registered HTTP and WebSocket routes",
		Long:  `loy routes statically inspects application source code to discover all registered endpoints and route bindings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir := "."
			if len(args) > 0 {
				rootDir = args[0]
			}

			cliOpts := GetOptions(cmd.Context())
			routes, err := route.InspectRoutes(fs, rootDir)
			if err != nil {
				return err
			}

			if cliOpts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(routes)
			}

			if len(routes) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No route registrations discovered.")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			_, _ = fmt.Fprintln(w, "METHOD\tPATH\tHANDLER\tSOURCE")
			for _, r := range routes {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s:%d\n", r.Method, r.Path, r.Handler, r.File, r.Line)
			}
			return w.Flush()
		},
	}
}
