package pipeline

import (
	"context"

	"github.com/julian7/goshipdone/ctx"
	// register internal modules
	_ "github.com/julian7/goshipdone/internal/modules"
	"gopkg.in/yaml.v3"
)

// LoadBuildPipeline creates a new BuildPipeline by reading YAML
// contents of a byte slice. Then, it makes sure default modules
// are loaded, providing safe defaults.
func LoadBuildPipeline(ymlcontent []byte) (*Pipeline, error) {
	pipeline := New([]*Stage{
		{
			Name:   "setup",
			Plural: "setups",
		},
		{
			Name:   "build",
			Plural: "builds",
		},
		{
			Name:   "publish",
			Plural: "publishes",
			SkipFN: func(cx context.Context) bool {
				context, err := ctx.GetShipContext(cx)
				if err != nil {
					return true
				}
				return !context.Publish
			},
		},
	})

	err := yaml.Unmarshal(ymlcontent, pipeline)
	if err != nil {
		return nil, err
	}

	for _, kind := range []string{
		"setup:env",
		"setup:project",
		"setup:git",
		"setup:skip_publish",
	} {
		_ = pipeline.LoadDefault(kind)
	}

	return pipeline, nil
}
