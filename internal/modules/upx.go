package modules

import (
	"context"
	"os/exec"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
	"github.com/magefile/mage/sh"
)

// UPX is a module for compressing executable binaries in a self-extracting
// format using `upx` tool.
type UPX struct {
	// Builds specifies a build names to find related artifacts to
	// modify.
	Builds []string
	// Skip specifies which os-arch items should be skipped
	Skip []string
}

func NewUPX() modules.Pluggable {
	return &UPX{Builds: []string{"default"}}
}

// Run calls upx on built artifacts, changing their artifact types
func (archive *UPX) Run(cx context.Context) error {
	upxCmd, err := exec.LookPath("upx")

	if err != nil {
		return err
	}

	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

	artifactMap := context.Artifacts.OsArchByIDs(archive.Builds, archive.Skip)
	if len(artifactMap) == 0 {
		return nil
	}

	args := []string{}

	for osarch := range artifactMap {
		for _, artifact := range *artifactMap[osarch] {
			args = append(args, artifact.Location)
		}
	}

	if err := sh.RunV(upxCmd, args...); err != nil {
		return err
	}

	return nil
}
