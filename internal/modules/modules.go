package modules

import "github.com/julian7/goshipdone/modules"

func Register() {
	for _, mod := range []*modules.ModuleRegistration{
		{Stage: "*", Type: "show", Factory: NewShow},
		{Stage: "setup", Type: "env", Factory: NewEnv},
		{Stage: "setup", Type: "git", Factory: NewGit},
		{Stage: "setup", Type: "skip_publish", Factory: NewSkipPublish},
		{Stage: "setup", Type: "project", Factory: NewProject},
		{Stage: "build", Type: "go", Factory: NewGo},
		{Stage: "build", Type: "checksum", Factory: NewChecksum},
		{Stage: "build", Type: "changelog", Factory: NewCutChangelog},
		{Stage: "build", Type: "tar", Factory: NewTar},
		{Stage: "build", Type: "upx", Factory: NewUPX},
		{Stage: "publish", Type: "artifact", Factory: NewArtifact},
		{Stage: "publish", Type: "scp", Factory: NewSCP},
	} {
		modules.RegisterModule(mod)
	}
}
