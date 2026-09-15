package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/loy-go/loy/internal/cli"
	"github.com/loy-go/loy/internal/diagnostics"
)

// Exit codes per Doc 09:
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

	var cmdErr *cli.CommandError
	if errors.As(err, &cmdErr) {
		outputDiagnostics(opts, cmdErr.Diagnostics)
		os.Exit(cmdErr.Code)
	}

	var diagPtr *diagnostics.Diagnostic
	if errors.As(err, &diagPtr) && diagPtr != nil {
		outputDiagnostics(opts, []*diagnostics.Diagnostic{diagPtr})
		os.Exit(ExitCommandError)
	}

	var diag diagnostics.Diagnostic
	if errors.As(err, &diag) {
		outputDiagnostics(opts, []*diagnostics.Diagnostic{&diag})
		os.Exit(ExitCommandError)
	}

	// Check if error is usage-related
	errMsg := err.Error()
	isUsage := strings.Contains(errMsg, "unknown command") ||
		strings.Contains(errMsg, "unknown flag") ||
		strings.Contains(errMsg, "unknown shorthand") ||
		strings.Contains(errMsg, "flag needs an argument") ||
		strings.Contains(errMsg, "accepts ") ||
		strings.Contains(errMsg, "requires ")

	code := diagnostics.CodeCLIExecutionFail
	exitCode := ExitCommandError
	if isUsage {
		code = diagnostics.CodeCLIUsageError
		exitCode = ExitUsageError
	}

	d := &diagnostics.Diagnostic{
		Severity: diagnostics.SeverityError,
		Code:     code,
		Message:  errMsg,
	}
	outputDiagnostics(opts, []*diagnostics.Diagnostic{d})
	os.Exit(exitCode)
}

func outputDiagnostics(opts *cli.GlobalOptions, diags []*diagnostics.Diagnostic) {
	concrete := make([]diagnostics.Diagnostic, 0, len(diags))
	for _, d := range diags {
		if d != nil {
			concrete = append(concrete, *d)
		}
	}

	if opts.JSON {
		formatter := &diagnostics.JSONFormatter{Indent: true}
		_ = formatter.Format(os.Stderr, concrete)
	} else if os.Getenv("GITHUB_ACTIONS") == "true" {
		formatter := &diagnostics.GitHubWorkflowFormatter{}
		_ = formatter.Format(os.Stderr, concrete)
	} else {
		formatter := &diagnostics.HumanFormatter{Color: !opts.NoColor}
		_ = formatter.Format(os.Stderr, concrete)
	}
}
