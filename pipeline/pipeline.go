// pipeline provides a configurable build pipeline, taking
// its inputs from a YAML source.
package pipeline

import (
	"fmt"

	"github.com/julian7/goshipdone/ctx"
	// register internal modules
	_ "github.com/julian7/goshipdone/internal/modules"

	"github.com/julian7/goshipdone/modules"
	"gopkg.in/yaml.v3"
)

// BuildPipeline represents the whole build pipeline. Module objects
// are loaded into BuildPipeline as read from .pipeline.yml.
type BuildPipeline struct {
	Setups    *modules.Modules `yaml:",omitempty"`
	Builds    *modules.Modules `yaml:",omitempty"`
	Publishes *modules.Modules `yaml:",omitempty"`
}

// LoadBuildPipeline creates a new BuildPipeline by reading YAML
// contents of a byte slice. Then, it makes sure default modules
// are loaded, providing safe defaults.
func LoadBuildPipeline(ymlcontent []byte) (*BuildPipeline, error) {
	pipeline := &BuildPipeline{
		Setups:    modules.NewModules("setup"),
		Builds:    modules.NewModules("build"),
		Publishes: modules.NewModules("publish"),
	}

	pipeline.Publishes.SkipFn = func(context *ctx.Context) bool {
		return !context.Publish
	}

	err := yaml.Unmarshal(ymlcontent, pipeline)
	if err != nil {
		return nil, err
	}

	for _, itemType := range []string{"env", "project", "git_tag", "skip_publish"} {
		_ = pipeline.Setups.Add(itemType, nil, true)
	}

	return pipeline, nil
}

// String returns a string representation of the build pipeline
func (pipeline *BuildPipeline) String() string {
	return fmt.Sprintf(
		"{Setups:%v Builds:%v Publishes:%v}",
		pipeline.Setups,
		pipeline.Builds,
		pipeline.Publishes,
	)
}

// Run executes build pipeline, calling Run on all
// Modules
func (pipeline *BuildPipeline) Run() error {
	steps := []*modules.Modules{
		pipeline.Setups,
		pipeline.Builds,
		pipeline.Publishes,
	}
	ctx := ctx.New()

	for _, step := range steps {
		if err := step.Run(ctx); err != nil {
			return err
		}
	}

	return nil
}
