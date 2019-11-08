package setup

import (
	"fmt"

	"github.com/julian7/magelib/modules"
	"github.com/magefile/mage/sh"
)

// GitTag is a module, which takes a git repo, and filling in
// `Version` information into `modules.Results`
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

// Run records git tag information into modules.Results
func (setup *GitTag) Run(results *modules.Results) error {
	versionTag, err := sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return fmt.Errorf("cannot detect version tag from git: %w", err)
	}

	results.Version = versionTag

	return nil
}
