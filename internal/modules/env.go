package modules

import (
	"os"
	"path"
	"strings"

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
	for _, env := range os.Environ() {
		data := strings.SplitN(env, "=", 2)
		if len(data) != 2 {
			continue
		}
		context.Env.Set(data[0], data[1])
	}

	if !context.Env.Has(EnvConfigHome) {
		for _, homeEnv := range []string{EnvHome, EnvHomePath} {
			home, ok := context.Env.Get(homeEnv)
			if ok {
				context.Env.Set(home, path.Join(
					home,
					".config",
				))
				break
			}
		}
	}
	return nil
}
