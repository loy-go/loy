package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uloydev/loy/internal/cli"
	"github.com/uloydev/loy/internal/diagnostics"
)

// Exit codes per Doc 09 and Phase 1 Plan:
// 0: success
// 1: command/project failure
// 2: usage/argument error
// 3: internal system error
const (
	ExitSuccess       = 0
	ExitCommandError  = 1
	ExitUsageError    = 2
	ExitInternalError = 3
)

func main() {
	rootCmd := cli.NewRootCmd()

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		handleError(rootCmd, err)
	}
}

func handleError(cmd *cobra.Command, err error) {
	opts := cli.GetOptions(cmd.Context())

	// Fallback inspect os.Args if flag parsing failed before setting options
	if !opts.JSON {
		for _, arg := range os.Args {
			if arg == "--json" {
				opts.JSON = true
				break
			}
		}
	}

	var diag diagnostics.Diagnostic
	if errors.As(err, &diag) {
		if opts.JSON {
			formatter := &diagnostics.JSONFormatter{Indent: true}
			_ = formatter.Format(os.Stderr, []diagnostics.Diagnostic{diag})
		} else {
			formatter := &diagnostics.HumanFormatter{Color: !opts.NoColor}
			_ = formatter.Format(os.Stderr, []diagnostics.Diagnostic{diag})
		}
		os.Exit(ExitCommandError)
	}

	// Check if error is usage-related (unknown command, unknown flag, argument error)
	errMsg := err.Error()
	isUsage := strings.Contains(errMsg, "unknown command") ||
		strings.Contains(errMsg, "unknown flag") ||
		strings.Contains(errMsg, "unknown shorthand") ||
		strings.Contains(errMsg, "accepts ") ||
		strings.Contains(errMsg, "requires ")

	if opts.JSON {
		diagErr := diagnostics.NewError(diagnostics.CodeCLIUsageError, errMsg)
		if !isUsage {
			diagErr.Code = diagnostics.CodeCLIExecutionFail
		}
		formatter := &diagnostics.JSONFormatter{Indent: true}
		_ = formatter.Format(os.Stderr, []diagnostics.Diagnostic{diagErr})
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
	}

	if isUsage {
		os.Exit(ExitUsageError)
	}
	os.Exit(ExitCommandError)
}
