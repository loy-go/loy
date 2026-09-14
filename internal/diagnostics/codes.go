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

	// Architecture & Rule codes
	CodeArchCycleDependency     = "LOY-ARCH-001"
	CodeArchDomainInfra         = "LOY-ARCH-002"
	CodeArchDomainTransport     = "LOY-ARCH-003"
	CodeArchAppTransport        = "LOY-ARCH-004"
	CodeArchAppConcreteInfra    = "LOY-ARCH-005"
	CodeArchInfraTransport      = "LOY-ARCH-006"
	CodeArchTransportLogic      = "LOY-ARCH-007"
	CodeArchForbiddenImport     = "LOY-ARCH-008"
	CodeArchForbiddenCategory   = "LOY-ARCH-009"
	CodeArchLayerDirection      = "LOY-ARCH-010"
	CodeArchServiceLocator      = "LOY-ARCH-011"
	CodeArchGlobalMutableState  = "LOY-ARCH-012"
	CodeArchWorkspaceBoundary   = "LOY-ARCH-013"
	CodeArchArtifactOwnership   = "LOY-ARCH-014"
	CodeArchPlatformImpure      = "LOY-ARCH-015"
	CodeArchSuppressionError    = "LOY-ARCH-099"

	// Doctor & Environment codes
	CodeDoctorGoMissing         = "LOY-DOC-001"
	CodeDoctorGitMissing        = "LOY-DOC-002"
	CodeDoctorSqlcMissing       = "LOY-DOC-003"
	CodeDoctorDockerMissing     = "LOY-DOC-004"
	CodeDoctorModuleMissing     = "LOY-DOC-005"
	CodeDoctorManifestMissing   = "LOY-DOC-006"
	CodeDoctorTemplMissing      = "LOY-DOC-007"
	CodeDoctorNodeMissing       = "LOY-DOC-008"
	CodeDoctorPackageMgrMissing = "LOY-DOC-009"

	// Internal system error
	CodeInternalError = "LOY-SYS-999"
)
