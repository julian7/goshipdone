// modules provides an extendable interface for executable components in the
// build pipeline.
package modules

import (
	"context"
	"fmt"
	"log"
	"time"
)

type (
	// Pluggable is a module, which can be pluggable into a pipeline
	Pluggable interface {
		Run(context.Context) error
	}

	// Module is a single module, specifying its type and its Pluggable
	Module struct {
		Type string
		Pluggable
	}
)

// Run executes a module, and measures its wallclock time spent
func (mod *Module) Run(ctx context.Context) error {
	log.Printf("----> %s", mod.Type)

	start := time.Now()

	if err := mod.Pluggable.Run(ctx); err != nil {
		return fmt.Errorf("%s: %w", mod.Type, err)
	}

	log.Printf("<---- %s done in %s", mod.Type, time.Since(start))

	return nil
}
