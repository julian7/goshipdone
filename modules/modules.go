package modules

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/julian7/magelib/ctx"
	"gopkg.in/yaml.v3"
)

type (
	// Pluggable is a module, which can be pluggable into a pipeline
	Pluggable interface {
		Run(*ctx.Context) error
	}

	// Modules is a list of Module-s of a single stage
	Modules struct {
		Stage   string   `yaml:"-"`
		Modules []Module `yaml:"-"`
	}

	// Module is a single module, specifying its type and its Pluggable
	Module struct {
		Type string
		Pluggable
	}
)

// NewModules is generating a new, empty Modules of a certain stage
func NewModules(stage string) *Modules {
	return &Modules{Stage: stage}
}

// UnmarshalYAML parses YAML node to load its modules
func (mod *Modules) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("definition of `%s` is not a sequence", mod.Stage)
	}

	for idx, child := range node.Content {
		if child.Kind != yaml.MappingNode {
			return fmt.Errorf("item #%d of `%s` definition is not a map", idx+1, mod.Stage)
		}

		itemType, err := getType(child)

		if err != nil {
			return fmt.Errorf(
				"definition %s, item #%d: %w",
				mod.Stage,
				idx+1,
				err,
			)
		}

		if err := mod.Add(itemType, child, false); err != nil {
			return err
		}
	}

	return nil
}

// Add adds a single module into Modules, decoding a YAML node if provided.
// It is also able to register a node only if not yet registered.
func (mod *Modules) Add(itemType string, node *yaml.Node, once bool) error {
	kind := fmt.Sprintf("%s:%s", mod.Stage, itemType)

	if once && isLoaded(kind) {
		return fmt.Errorf("module %s already loaded", kind)
	}

	targetModFactory, ok := LookupModule(kind)
	if !ok {
		return fmt.Errorf("unknown module %s", kind)
	}

	targetMod := targetModFactory()

	if node != nil {
		if err := node.Decode(targetMod); err != nil {
			return fmt.Errorf("cannot decode module %s: %w", kind, err)
		}
	}

	mod.Modules = append(mod.Modules, Module{
		Type:      itemType,
		Pluggable: targetMod,
	})

	return nil
}

func (mod *Modules) String() string {
	mods := make([]string, 0, len(mod.Modules))
	for _, module := range mod.Modules {
		mods = append(mods, fmt.Sprintf("%+v", module))
	}

	return fmt.Sprintf(
		"{%s [%s]}",
		mod.Stage,
		strings.Join(mods, " "),
	)
}

// Run goes through all internally loaded modules, and run them
// one by one.
func (mod *Modules) Run(context *ctx.Context) error {
	fmt.Printf("====> %s\n", strings.ToUpper(mod.Stage))

	for _, module := range mod.Modules {
		fmt.Printf("----> %s\n", module.Type)

		missing, err := MissingDepsForModule(fmt.Sprintf("%s:%s", mod.Stage, module.Type))
		if err != nil {
			return err
		}

		if len(missing) > 0 {
			return fmt.Errorf("missing dependencies: %s", strings.Join(missing, ", "))
		}

		if err := module.Pluggable.Run(context); err != nil {
			return fmt.Errorf("%s:%s: %w", mod.Stage, module.Type, err)
		}

		fmt.Printf("----< %s done\n", module.Type)
	}

	return nil
}

func getType(node *yaml.Node) (string, error) {
	var itemType string

	for idx := 0; idx < len(node.Content); idx += 2 {
		key := node.Content[idx]
		val := node.Content[idx+1]

		if key.Value == "type" {
			itemType = val.Value
			return itemType, nil
		}
	}

	return "", errors.New("type not defined")
}
