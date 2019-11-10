package modules

import (
	"os"
	"path"

	"github.com/julian7/magelib/ctx"
	"github.com/julian7/magelib/modules"
)

// Project is a module for setting basic project-specific data
type Project struct {
	Name      string
	TargetDir string `yaml:"target"`
}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "setup:project",
		Factory: NewProject,
	})
}

// NewProject is the factory function for Project
func NewProject() modules.Pluggable {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	} else {
		pwd = path.Base(pwd)
	}

	return &Project{
		Name:      pwd,
		TargetDir: "dist",
	}
}

// Run records project's basic information into ctx.Context
func (proj *Project) Run(context *ctx.Context) error {
	context.ProjectName = proj.Name
	context.TargetDir = proj.TargetDir

	return nil
}
