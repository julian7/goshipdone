// ctx provides a cumulative structure carried over to each module
// in the build pipeline.
package ctx

import (
	"context"
	"errors"

	"github.com/julian7/withenv"
)

type info struct{}

var Info = &info{}

// Context are a cumulative structure carried over to each module,
// to contain data later steps might require
type Context struct {
	context.Context
	Artifacts   Artifacts
	Env         *withenv.Env
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

func New(ctx context.Context) context.Context {
	return context.WithValue(
		ctx,
		Info,
		&Context{
			Context: ctx,
			Env:     withenv.New(),
			Git:     new(GitData),
		},
	)
}

// GetShipContext returns *ctx.Context from context.Context
func GetShipContext(cx context.Context) (*Context, error) {
	context, ok := cx.Value(Info).(*Context)
	if !ok {
		return nil, errors.New("ship context not provided")
	}

	return context, nil
}
