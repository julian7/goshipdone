package modules

import "fmt"

// nolint: gochecknoglobals
var modRegistry map[string]PluggableFactory

type (
	// PluggableFactory is a method, which yields a Pluggable
	PluggableFactory func() Pluggable

	// ModuleRegistration is a Pluggable registration entry for
	// module registration
	ModuleRegistration struct {
		// Stage is a string representation of the stage the module is
		// registered into. To register a module into every stage, specify
		// "*" as stage.
		Stage string
		// Type is a string representation of the module's name.
		Type string
		// Factory is the factory method to create a new module
		// with defaults.
		Factory PluggableFactory
	}
)

func (mod *ModuleRegistration) Kind() string {
	return fmt.Sprintf("%s:%s", mod.Stage, mod.Type)
}

// RegisterModule allows modules to register themselves during init(),
// by providing a definition of type ModuleRegistration.
func RegisterModule(definition *ModuleRegistration) {
	if modRegistry == nil {
		modRegistry = make(map[string]PluggableFactory)
	}

	modRegistry[definition.Kind()] = definition.Factory
}

// LookupModule returns a PluggableFactory based on its Kind
// as a side effect, it also flags the module as loaded
func LookupModule(kind string) (PluggableFactory, bool) {
	mod, ok := modRegistry[kind]
	if ok {
		return mod, true
	}

	return nil, false
}
