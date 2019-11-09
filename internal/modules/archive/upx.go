package archive

import (
	"os/exec"

	"github.com/julian7/magelib/ctx"
	"github.com/julian7/magelib/modules"
	"github.com/magefile/mage/sh"
)

// UPX is a module for compressing executable binaries in a self-extracting
// format using `upx` tool.
type UPX struct {
	// Build specifies a single build name to find related artifacts to
	// modify.
	Build string
}

//nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "archive:upx",
		Factory: NewUPX,
	})
}

func NewUPX() modules.Pluggable {
	return &UPX{Build: "default"}
}

// Run calls upx on built artifacts, changing their artifact types
func (archive *UPX) Run(context *ctx.Context) error {
	upxCmd, err := exec.LookPath("upx")

	if err != nil {
		return err
	}

	artifacts := context.Artifacts.ByName(archive.Build)
	if len(artifacts) == 0 {
		return nil
	}

	args := make([]string, len(artifacts))

	for i, artifact := range artifacts {
		args[i] = artifact.Filename
	}

	if err := sh.RunV(upxCmd, args...); err != nil {
		return err
	}

	for _, artifact := range artifacts {
		artifact.Format = ctx.FormatUPX
	}

	return nil
}
