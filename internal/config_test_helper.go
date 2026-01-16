package internal

// boolFields represents the boolean configuration fields used in merge tests.
type boolFields struct {
	AnalyzeDependencies bool
	ShowSecurityInfo    bool
	Verbose             bool
	Quiet               bool
	UseDefaultBranch    bool
}

// createBoolFieldMergeTest creates a test table entry for testing boolean field merging.
// This helper reduces duplication by standardizing the creation of AppConfig test structures
// with boolean fields.
func createBoolFieldMergeTest(name string, dst, src, want boolFields) struct {
	name string
	dst  *AppConfig
	src  *AppConfig
	want *AppConfig
} {
	return struct {
		name string
		dst  *AppConfig
		src  *AppConfig
		want *AppConfig
	}{
		name: name,
		dst: &AppConfig{
			AnalyzeDependencies: dst.AnalyzeDependencies,
			ShowSecurityInfo:    dst.ShowSecurityInfo,
			Verbose:             dst.Verbose,
			Quiet:               dst.Quiet,
			UseDefaultBranch:    dst.UseDefaultBranch,
		},
		src: &AppConfig{
			AnalyzeDependencies: src.AnalyzeDependencies,
			ShowSecurityInfo:    src.ShowSecurityInfo,
			Verbose:             src.Verbose,
			Quiet:               src.Quiet,
			UseDefaultBranch:    src.UseDefaultBranch,
		},
		want: &AppConfig{
			AnalyzeDependencies: want.AnalyzeDependencies,
			ShowSecurityInfo:    want.ShowSecurityInfo,
			Verbose:             want.Verbose,
			Quiet:               want.Quiet,
			UseDefaultBranch:    want.UseDefaultBranch,
		},
	}
}
