package archive

import (
	"fmt"

	"github.com/julian7/magelib/modules"
)

// Show is a module for listing all recorded artifacts so far
type Show struct{}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "archive:show",
		Factory: NewShow,
	})
}

// NewShow returns a new Show module
func NewShow() modules.Pluggable {
	return &Show{}
}

// Run provides a list of artifacts recorded so far
func (Show) Run(results *modules.Results) error {
	for _, art := range results.Artifacts {
		fmt.Printf("- %+v\n", art)
	}

	return nil
}
