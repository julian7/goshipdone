// modules provides an extendable interface for executable components in the
// build pipeline.
package modules

import (
	"fmt"
	"log"
	"time"

	"github.com/julian7/goshipdone/ctx"
)

type (
	// Pluggable is a module, which can be pluggable into a pipeline
	Pluggable interface {
		Run(*ctx.Context) error
	}

	// Module is a single module, specifying its type and its Pluggable
	Module struct {
		Type string
		Pluggable
	}
)

// Run executes a module, and measures its wallclock time spent
func (mod *Module) Run(context *ctx.Context) error {
	log.Printf("----> %s", mod.Type)

	start := time.Now()

	if err := mod.Pluggable.Run(context); err != nil {
		return fmt.Errorf("%s: %w", mod.Type, err)
	}

	log.Printf("<---- %s done in %s", mod.Type, time.Since(start))

	return nil
}
