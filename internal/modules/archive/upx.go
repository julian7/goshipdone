package archive

import (
	"os/exec"

	"github.com/julian7/magelib/modules"
	"github.com/magefile/mage/sh"
)

// UPX is a module for compressing executable binaries in a self-extracting
// format using `upx` tool.
type UPX struct {
	Name string
}

//nolint: // nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "archive:upx",
		Factory: NewUPX,
	})
}

func NewUPX() modules.Pluggable {
	return &UPX{Name: "default"}
}

func (archive *UPX) Run(results *modules.Results) error {
	upxCmd, err := exec.LookPath("upx")

	if err != nil {
		return err
	}

	artifacts := results.ArtifactsByName(archive.Name)
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
		artifact.Format = modules.FormatUPX
	}

	return nil
}
