package modules

import (
	"fmt"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
	"github.com/magefile/mage/sh"
)

// GitTag is a module, which takes a git repo, and filling in
// `Version` information into `ctx.Context`
type GitTag struct{}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "setup:git_tag",
		Factory: NewGitTag,
	})
}

// NewGitTag is the factory function for GitTag
func NewGitTag() modules.Pluggable {
	return &GitTag{}
}

// Run records git tag information into ctx.Context
func (setup *GitTag) Run(context *ctx.Context) error {
	versionTag, err := sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return fmt.Errorf("cannot detect version tag from git: %w", err)
	}

	context.Version = versionTag

	return nil
}
