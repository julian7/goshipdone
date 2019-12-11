package modules

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
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

// NewSkipPublish is a factory method for SkipPublish plugin
func NewSkipPublish() modules.Pluggable {
	return &SkipPublish{
		EnvName: "SKIP_PUBLISH",
	}
}

func (mod *SkipPublish) Run(cx context.Context) error {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

	if variable, ok := context.Env.Get(mod.EnvName); ok {
		skip, err := strconv.ParseBool(variable)
		if err != nil {
			return fmt.Errorf("parsing %s as bool: %w", mod.EnvName, err)
		}

		context.Publish = !skip
		log.Printf("publishing is set to %v", context.Publish)
	}

	return nil
}
