package model

// Ownership represents the governance rule for an artifact or file.
type Ownership string

const (
	// DeveloperOwned files are created once and fully owned by the developer.
	// Overwrites are strictly prohibited unless --force is specified.
	DeveloperOwned Ownership = "developer"

	// GeneratedOwned files are fully managed by the generator.
	// Regenerations overwrite unless developer modifications are detected without --force.
	GeneratedOwned Ownership = "generated"

	// MixedOwned files contain developer code alongside Loy-managed regions.
	// Reconciled only through structured comment regions (// loy:region:...).
	MixedOwned Ownership = "mixed"
)

// String returns the string representation of Ownership.
func (o Ownership) String() string {
	return string(o)
}
