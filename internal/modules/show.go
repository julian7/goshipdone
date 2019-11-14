package modules

import (
	"log"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
)

// Show is a module for listing all recorded artifacts so far
type Show struct{}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "*",
		Type:    "show",
		Factory: NewShow,
	})
}

// NewShow returns a new Show module
func NewShow() modules.Pluggable {
	return &Show{}
}

// Run provides a list of artifacts recorded so far
func (Show) Run(context *ctx.Context) error {
	for _, art := range context.Artifacts {
		log.Printf("- %s: %s (%s)", art.ID, art.Filename, art.OsArch())
	}

	return nil
}
