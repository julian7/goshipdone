// pipeline provides a configurable build pipeline, taking
// its inputs from a YAML source.
package pipeline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/julian7/goshipdone/ctx"
	"gopkg.in/yaml.v3"
)

// Pipeline is a generic pipeline, with a registry and stages configured.
type Pipeline struct {
	Stages []*Stage
}

func New(stages []*Stage) *Pipeline {
	pip := &Pipeline{Stages: make([]*Stage, 0, len(stages))}
	pip.Stages = append(pip.Stages, stages...)

	return pip
}

// UnmarshalYAML parses YAML node to load its modules
func (pip *Pipeline) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.New("pipeline definition is not a map")
	}

	l := len(node.Content)
	for i := 0; i < l; i += 2 {
		var stage *Stage
		stageDefName := node.Content[i].Value

		for _, st := range pip.Stages {
			if st.Plural == stageDefName {
				stage = st
			}
		}

		if stage == nil {
			continue
		}

		if err := node.Content[i+1].Decode(stage); err != nil {
			return fmt.Errorf("decoding %s stage: %w", stage.Name, err)
		}
	}
	return nil
}

// LoadDefault loads a module into a stage, if not loaded yet
func (pip *Pipeline) LoadDefault(kind string) error {
	items := strings.SplitN(kind, ":", 2)
	if len(items) != 2 {
		return fmt.Errorf("invalid module kind: %q", kind)
	}

	if stg := pip.StageByName(items[0]); stg != nil {
		if err := stg.Add(items[1], nil, true); err != nil {
			return err
		}
	}

	return nil
}

func (pip *Pipeline) StageByName(name string) *Stage {
	for _, stage := range pip.Stages {
		if stage.Name == name {
			return stage
		}
	}

	return nil
}

// Run executes build pipeline, calling Run on all
// Modules
func (pip *Pipeline) Run() error {
	ctx := ctx.New()

	for _, stg := range pip.Stages {
		if err := stg.Run(ctx); err != nil {
			return err
		}
	}

	return nil
}
