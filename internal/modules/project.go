package modules

import (
	"context"
	"os"
	"path"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
)

// Project is a module for setting basic project-specific data
type Project struct {
	Name      string
	TargetDir string `yaml:"target"`
}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "setup",
		Type:    "project",
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
func (mod *Project) Run(cx context.Context) error {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

	context.ProjectName = mod.Name
	context.TargetDir = mod.TargetDir

	return nil
}
