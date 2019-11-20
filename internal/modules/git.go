package modules

import (
	"fmt"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
	"github.com/magefile/mage/sh"
)

// Git is a module, which takes a git repo, and filling in
// `Version` information into `ctx.Context`
type Git struct{}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "setup",
		Type:    "git_tag",
		Factory: NewGit,
	})
}

// NewGit is the factory function for Git
func NewGit() modules.Pluggable {
	return &Git{}
}

// Run records git tag information into ctx.Context
func (*Git) Run(context *ctx.Context) error {
	items := []struct {
		name     string
		required bool
		target   *string
		args     []string
	}{
		{"version info", true, &context.Version, []string{"describe", "--tags", "--always", "--dirty"}},
		{"current tag", false, &context.Git.Tag, []string{"describe", "--exact-match", "--tags"}},
		{"current ref", true, &context.Git.Ref, []string{"-P", "show", "--format=%H", "-s"}},
		{"url", false, &context.Git.URL, []string{"ls-remote", "--get-url"}},
	}

	for _, item := range items {
		val, err := sh.Output("git", item.args...)
		if item.required && err != nil {
			return fmt.Errorf("cannot detect %s from git: %w", item.name, err)
		}

		*item.target = val
	}

	return nil
}
