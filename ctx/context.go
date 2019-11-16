// ctx provides a cumulative structure carried over to each module
// in the build pipeline.
package ctx

import "context"

// Context are a cumulative structure carried over to each module,
// to contain data later steps might require
type Context struct {
	context.Context
	Artifacts   Artifacts
	Env         Env
	Git         *GitData
	ProjectName string
	Publish     bool
	TargetDir   string
	Version     string
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

func New() *Context {
	return &Context{
		Context: context.Background(),
		Env:     NewEnv(),
		Git:     new(GitData),
	}
}
