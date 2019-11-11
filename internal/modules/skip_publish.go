package modules

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/julian7/magelib/ctx"
	"github.com/julian7/magelib/modules"
)

// SkipPublish module controls whether publish phase should be executed, by
// reading from an environment variable. This is an automatically loaded
// extension.
type SkipPublish struct {
	// EnvName specifies which environment variable should be used to
	// signal skipping publish. Default: `SKIP_PUBLISH`, and while it reads
	// what strconv.ParseBool understands, the only reasonable value for this
	// variable is falsey (eg. "false", 0, and similar).
	EnvName string `yaml:"env_name"`
}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "setup:skip_publish",
		Factory: NewSkipPublish,
	})
}

// NewSkipPublish is a factory method for SkipPublish plugin
func NewSkipPublish() modules.Pluggable {
	return &SkipPublish{
		EnvName: "SKIP_PUBLISH",
	}
}

func (mod *SkipPublish) Run(context *ctx.Context) error {
	if variable, ok := os.LookupEnv(mod.EnvName); ok {
		skip, err := strconv.ParseBool(variable)
		if err != nil {
			return fmt.Errorf("parsing %s as bool: %w", mod.EnvName, err)
		}

		context.Publish = !skip
		log.Printf("publishing is set to %v", context.Publish)
	}

	return nil
}
