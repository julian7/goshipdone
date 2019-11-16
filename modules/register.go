package modules

import "fmt"

// nolint: gochecknoglobals
var modRegistry map[string]*PluggableCache

type (
	// PluggableFactory is a method, which yields a Pluggable
	PluggableFactory func() Pluggable

	// PluggableCache is a registry entry for a Pluggable
	PluggableCache struct {
		Factory PluggableFactory
		Deps    []string
		Loaded  bool
	}

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
		// Deps contains a list of Kind-s (strings in a form of "stage:type"),
		// which are required for this module.
		Deps []string
	}
)

func (mod *ModuleRegistration) Kind() string {
	return fmt.Sprintf("%s:%s", mod.Stage, mod.Type)
}

// RegisterModule allows modules to register themselves during init(),
// by providing a definition of type ModuleRegistration.
func RegisterModule(definition *ModuleRegistration) {
	if modRegistry == nil {
		modRegistry = make(map[string]*PluggableCache)
	}

	modRegistry[definition.Kind()] = &PluggableCache{
		Factory: definition.Factory,
		Loaded:  false,
	}
	copy(modRegistry[definition.Kind()].Deps, definition.Deps)
}

// MissingDepsForModule returns module's dependedcies, which weren't
// loaded. It silently returns nil if module is not registered,
// allowing dynamic modules to be passed through.
//
// Module is referenced by its `Kind`, see `ModuleRegistration`.
func MissingDepsForModule(kind string) []string {
	mod, ok := modRegistry[kind]
	if !ok {
		return nil
	}

	missing := []string{}

	for _, dep := range mod.Deps {
		if !isLoaded(dep) {
			missing = append(missing, dep)
		}
	}

	return missing
}

// LookupModule returns a PluggableFactory based on its Kind
// as a side effect, it also flags the module as loaded
func LookupModule(kind string) (PluggableFactory, bool) {
	mod, ok := modRegistry[kind]
	if ok {
		mod.Loaded = true
		return mod.Factory, true
	}

	return nil, false
}

func isLoaded(kind string) bool {
	mod, ok := modRegistry[kind]
	if ok {
		return mod.Loaded
	}

	return false
}
