// ctx provides a cumulative structure carried over to each module
// in the build pipeline.
package ctx

// Context are a cumulative structure carried over to each module,
// to contain data later steps might require
type Context struct {
	Artifacts   Artifacts
	ProjectName string
	Publish     bool
	TargetDir   string
	Version     string
	Git         GitData
}

// GitData contains git-specific information on the repository
type GitData struct {
	// Tag contains git tag information, if the repo is on a specific tag
	Tag string
	// Ref contains the full SHA1 checksum of the current commit
	Ref string
	// URL contains git repo's URL, collected from current branch's upstream
	URL string
}
