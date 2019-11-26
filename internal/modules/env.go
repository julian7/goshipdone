package modules

import (
	"os"
	"path"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
)

const (
	EnvConfigHome = "XDG_CONFIG_HOME"
	EnvHome       = "HOME"
	EnvHomePath   = "HOMEPATH"
)

// Env module sets up context's Env hash
type Env struct{}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "setup",
		Type:    "env",
		Factory: NewEnv,
	})
}

func NewEnv() modules.Pluggable {
	return &Env{}
}

func (*Env) Run(context *ctx.Context) error {
	if err := context.Env.Load(os.Environ()); err != nil {
		return err
	}

	if _, ok := context.Env.Get(EnvConfigHome); !ok {
		for _, homeEnv := range []string{EnvHome, EnvHomePath} {
			home, ok := context.Env.Get(homeEnv)
			if ok {
				context.Env.Set(EnvConfigHome, path.Join(
					home,
					".config",
				))

				break
			}
		}
	}

	return nil
}
