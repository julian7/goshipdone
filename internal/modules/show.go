package modules

import (
	"context"
	"log"
	"sort"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
)

// Show is a module for listing all recorded artifacts so far
type Show struct{}

// NewShow returns a new Show module
func NewShow() modules.Pluggable {
	return &Show{}
}

// Run provides a list of artifacts recorded so far
func (Show) Run(cx context.Context) error {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

	envKeys := make([]string, 0, len(context.Env.Vars))
	for key := range context.Env.Vars {
		envKeys = append(envKeys, key)
	}

	sort.Strings(envKeys)

	log.Printf("Environment:")

	for _, env := range envKeys {
		log.Printf("- %s = %q", env, context.Env.GetOrDefault(env, "-unset-"))
	}

	log.Printf("Artifacts:")

	for _, art := range context.Artifacts {
		log.Printf("- %s: %s (%s)", art.ID, art.Filename, art.OSArch())
	}

	return nil
}
