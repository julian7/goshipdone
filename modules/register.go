package modules

// nolint: gochecknoglobals
var modRegistry map[string]PluggableCache

type (
	// PluggableFactory is a method, which yields a Pluggable
	PluggableFactory func() Pluggable

	// PluggableCache is a registry entry for a Pluggable
	PluggableCache struct {
		Factory PluggableFactory
		Deps    []string
		Loaded  bool
	}

	// PluggableModule is a Pluggable registration entry for
	// module registration
	PluggableModule struct {
		// Kind is a string representation of stage/type how
		// the module can be referred to. Format: "`stage`:`type`", like
		// "build:script". To register a module for every stage, use
		// "*:`type`" format.
		Kind string
		// Factory is the factory method to create a new module
		// with defaults.
		Factory PluggableFactory
		// Deps contains a list of Kind-s, which are required
		// for this module.
		Deps []string
	}
)

// RegisterModule allows modules to register themselves during init(),
// by providing a definition of type PluggableModule.
func RegisterModule(definition *PluggableModule) {
	if modRegistry == nil {
		modRegistry = make(map[string]PluggableCache)
	}

	modRegistry[definition.Kind] = PluggableCache{
		Factory: definition.Factory,
		Loaded:  false,
	}
	copy(modRegistry[definition.Kind].Deps, definition.Deps)
}

// MissingDepsForModule returns module's dependedcies, which weren't
// loaded. It silently returns nil if module is not registered,
// allowing dynamic modules to be passed through.
//
// Module is referenced by its `Kind`, see `PluggableModule`.
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
