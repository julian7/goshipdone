package modules

import (
	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
	"github.com/magefile/mage/sh"
)

// SCP is a module for uploading artifacts to a remote server via scp
type SCP struct {
	// Builds specifies which build names should be added to the archive.
	Builds []string
	// Skip specifies GOOS-GOArch combinations to be skipped.
	// They are in `{{.Os}}-{{.Arch}}` format.
	// It filters builds to be included.
	Skip []string
	// Target specifies SCP endpoint as the last parameter of the `scp`
	// command. Example: staticfiles@remoteserver.com:/var/www/default/public
	Target string
}

func init() {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "publish:scp",
		Factory: NewSCP,
	})
}

// NewSCP is a factory function for SCP module
func NewSCP() modules.Pluggable {
	return &SCP{
		Builds: []string{"archive"},
		Skip:   []string{},
		Target: "",
	}
}

// Run takes specified artifacts, and uploads them to a SSH server
func (scp *SCP) Run(context *ctx.Context) error {
	builds := context.Artifacts.OsArchByIDs(scp.Builds, scp.Skip)

	cmdArgs := []string{}
	for osarch := range builds {
		for _, artifact := range *builds[osarch] {
			cmdArgs = append(cmdArgs, artifact.Location)
		}
	}

	cmdArgs = append(cmdArgs, scp.Target)

	return sh.RunV("scp", cmdArgs...)
}
