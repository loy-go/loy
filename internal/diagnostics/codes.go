package diagnostics

// Diagnostic codes standardized across the Loy platform.
const (
	// CLI codes
	CodeCLIUsageError    = "LOY-CLI-001"
	CodeCLIExecutionFail = "LOY-CLI-002"

	// Filesystem codes
	CodeFSNotFound      = "LOY-FS-001"
	CodeFSPathTraversal = "LOY-FS-002"
	CodeFSPermission    = "LOY-FS-003"

	// Process codes
	CodeProcessTimeout   = "LOY-PROC-001"
	CodeProcessExecution = "LOY-PROC-002"

	// Configuration & Manifest codes
	CodeConfigSyntaxError     = "LOY-CFG-001"
	CodeConfigValidationError = "LOY-CFG-002"
	CodeConfigPresetNotFound  = "LOY-CFG-003"

	// Project & Workspace codes
	CodeProjectRootNotFound  = "LOY-PRJ-001"
	CodeProjectAlreadyExists = "LOY-PRJ-002"
	CodeWorkspaceConflict    = "LOY-WRK-001"
	CodeWorkspaceTargetError = "LOY-WRK-002"

	// Generator & Template codes
	CodeGenConflict       = "LOY-GEN-001"
	CodeGenRegionCorrupt  = "LOY-GEN-002"
	CodeGenTemplateError  = "LOY-GEN-003"
	CodeGenExecutionError = "LOY-GEN-004"

	// Internal system error
	CodeInternalError = "LOY-SYS-999"
)
