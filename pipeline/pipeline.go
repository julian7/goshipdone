package pipeline

import (
	"fmt"

	// register archive modules
	"github.com/julian7/magelib/ctx"
	_ "github.com/julian7/magelib/internal/modules/archive"

	// register build modules
	_ "github.com/julian7/magelib/internal/modules/build"
	// register setup modules
	_ "github.com/julian7/magelib/internal/modules/setup"
	"github.com/julian7/magelib/modules"
	"gopkg.in/yaml.v3"
)

// BuildPipeline represents the whole build pipeline. Module objects
// are loaded into BuildPipeline as read from .pipeline.yml.
type BuildPipeline struct {
	Setups       *modules.Modules `yaml:",omitempty"`
	Builds       *modules.Modules `yaml:",omitempty"`
	Archives     *modules.Modules `yaml:",omitempty"`
	ReleaseNotes *modules.Modules `yaml:",omitempty"`
	Publishes    *modules.Modules `yaml:",omitempty"`
}

// LoadBuildPipeline creates a new BuildPipeline by reading YAML
// contents of a byte slice. Then, it makes sure default modules
// are loaded, providing safe defaults.
func LoadBuildPipeline(ymlcontent []byte) (*BuildPipeline, error) {
	pipeline := &BuildPipeline{
		Setups:       modules.NewModules("setup"),
		Builds:       modules.NewModules("build"),
		Archives:     modules.NewModules("archive"),
		ReleaseNotes: modules.NewModules("release_note"),
		Publishes:    modules.NewModules("publish"),
	}

	err := yaml.Unmarshal(ymlcontent, pipeline)
	if err != nil {
		return nil, err
	}

	for _, itemType := range []string{"project", "git_tag"} {
		_ = pipeline.Setups.Add(itemType, nil, true)
	}

	return pipeline, nil
}

// String returns a string representation of the build pipeline
func (pipeline *BuildPipeline) String() string {
	return fmt.Sprintf(
		"{Setups:%v Builds:%v Archives:%v ReleaseNotes:%v Publishes:%v}",
		pipeline.Setups,
		pipeline.Builds,
		pipeline.Archives,
		pipeline.ReleaseNotes,
		pipeline.Publishes,
	)
}

// Run executes build pipeline, calling Run on all
// Modules
func (pipeline *BuildPipeline) Run() error {
	steps := []*modules.Modules{
		pipeline.Setups,
		pipeline.Builds,
		pipeline.Archives,
		pipeline.ReleaseNotes,
		pipeline.Publishes,
	}
	ctx := &ctx.Context{}

	for _, step := range steps {
		if err := step.Run(ctx); err != nil {
			return err
		}
	}

	return nil
}
