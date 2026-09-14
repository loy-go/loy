package cli

import (
	"fmt"

	"github.com/loy-go/loy/internal/diagnostics"
)

// CommandError represents a failed command execution carrying structured diagnostics and an exit code.
type CommandError struct {
	Code        int
	Diagnostics []*diagnostics.Diagnostic
}

func (e *CommandError) Error() string {
	if len(e.Diagnostics) > 0 && e.Diagnostics[0] != nil {
		return e.Diagnostics[0].Message
	}
	return fmt.Sprintf("command failed with exit code %d", e.Code)
}
