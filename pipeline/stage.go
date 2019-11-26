package pipeline

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
	"gopkg.in/yaml.v3"
)

// Stage is a single stage in the pipeline
type Stage struct {
	loaded  map[string]bool
	Modules []*modules.Module       `yaml:"-"`
	Name    string                  `yaml:"-"`
	Plural  string                  `yaml:"-"`
	SkipFN  func(*ctx.Context) bool `yaml:"-"`
}

func NewStage(name, plural string) *Stage {
	return &Stage{Name: name, Plural: plural}
}

// UnmarshalYAML parses YAML node to load its modules
func (stg *Stage) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("definition of `%s` is not a sequence", stg.Name)
	}

	for idx, child := range node.Content {
		if child.Kind != yaml.MappingNode {
			return fmt.Errorf("item #%d of `%s` definition is not a map", idx+1, stg.Name)
		}

		itemType, err := getType(child)

		if err != nil {
			return fmt.Errorf(
				"definition %s, item #%d: %w",
				stg.Name,
				idx+1,
				err,
			)
		}

		if err := stg.Add(itemType, child, false); err != nil {
			return err
		}
	}

	return nil
}

// Add adds a single module into Stage, decoding a YAML node if provided.
// It is also able to register a node only if not yet registered.
// By default, Stage allows registration of its own stage only, but
// modules registered for all stages are also accepted, if there is no
// specific module registration exists.
//
// Eg. if there are two different modules registered for "*:dump" and
// "build:dump", a reference to "dump" kind in publishes will fire "*:dump"
// module, but a similar "dump" kind in builds will fire "build:dump".
func (stg *Stage) Add(itemType string, node *yaml.Node, once bool) error {
	var kind string

	var targetModFactory modules.PluggableFactory

	var ok bool

	for _, stage := range []string{stg.Name, "*"} {
		kind = fmt.Sprintf("%s:%s", stage, itemType)

		targetModFactory, ok = modules.LookupModule(kind)
		if ok {
			break
		}
	}

	if !ok {
		return fmt.Errorf("unknown module %s:%s", stg.Name, itemType)
	}

	if once && stg.isLoaded(kind) {
		return fmt.Errorf("module %s already loaded", kind)
	}

	targetMod := targetModFactory()

	if node != nil {
		if err := node.Decode(targetMod); err != nil {
			return fmt.Errorf("cannot decode module %s: %w", kind, err)
		}
	}

	stg.Modules = append(stg.Modules, &modules.Module{
		Type:      itemType,
		Pluggable: targetMod,
	})

	stg.flagLoaded(kind)

	return nil
}

// Run goes through all internally loaded modules, and run them
// one by one.
func (stg *Stage) Run(context *ctx.Context) error {
	log.Printf("====> %s", strings.ToUpper(stg.Name))

	startMod := time.Now()

	if stg.SkipFN != nil && stg.SkipFN(context) {
		log.Printf("SKIPPED")
	} else {
		for _, module := range stg.Modules {
			if err := module.Run(context); err != nil {
				return fmt.Errorf("stage %s: %w", stg.Name, err)
			}
		}
	}

	log.Printf("<==== %s done in %s", strings.ToUpper(stg.Name), time.Since(startMod))

	return nil
}

func (stg *Stage) isLoaded(kind string) bool {
	if stg.loaded == nil {
		return false
	}

	_, loaded := stg.loaded[kind]

	return loaded
}

func (stg *Stage) flagLoaded(kind string) {
	if stg.loaded == nil {
		stg.loaded = map[string]bool{}
	}

	stg.loaded[kind] = true
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
