package pipeline_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/go-test/deep"
	"github.com/julian7/goshipdone/ctx"
	intmod "github.com/julian7/goshipdone/internal/modules"
	"github.com/julian7/goshipdone/modules"
	"github.com/julian7/goshipdone/pipeline"
)

type testModuleRegistration struct {
	reporter func()
}

func testModuleRegistrationFactory() modules.Pluggable {
	return &testModuleRegistration{}
}

func (mod *testModuleRegistration) Run(*ctx.Context) error {
	if mod.reporter != nil {
		mod.reporter()
	}

	return nil
}

type testFailingModuleRegistration struct{}

func testFailingModuleRegistrationFactory() modules.Pluggable {
	return &testFailingModuleRegistration{}
}

func (mod *testFailingModuleRegistration) Run(*ctx.Context) error {
	return errors.New("error")
}

// nolint: funlen
func TestLoadBuildPipeline(t *testing.T) {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "build",
		Type:    "test",
		Factory: testModuleRegistrationFactory,
	})

	tests := []struct {
		name       string
		ymlcontent []byte
		want       *pipeline.Pipeline
		wantErr    bool
	}{
		{
			"empty loads defaults",
			[]byte("---\n"),
			&pipeline.Pipeline{
				Stages: []*pipeline.Stage{
					{
						Name:   "setup",
						Plural: "setups",
						Modules: []*modules.Module{
							{Type: "env", Pluggable: intmod.NewEnv()},
							{Type: "project", Pluggable: intmod.NewProject()},
							{Type: "git_tag", Pluggable: intmod.NewGit()},
							{Type: "skip_publish", Pluggable: intmod.NewSkipPublish()},
						},
					},
					{Name: "build", Plural: "builds"},
					{Name: "publish", Plural: "publishes"},
				},
			},
			false,
		},
		{
			"invoked mod loads",
			[]byte("---\nbuilds:\n  - type: test\n"),
			&pipeline.Pipeline{
				Stages: []*pipeline.Stage{
					{
						Name:   "setup",
						Plural: "setups",
						Modules: []*modules.Module{
							{Type: "env", Pluggable: intmod.NewEnv()},
							{Type: "project", Pluggable: intmod.NewProject()},
							{Type: "git_tag", Pluggable: intmod.NewGit()},
							{Type: "skip_publish", Pluggable: intmod.NewSkipPublish()},
						},
					},
					{
						Name:   "build",
						Plural: "builds",
						Modules: []*modules.Module{
							{Type: "test", Pluggable: testModuleRegistrationFactory()},
						},
					},
					{Name: "publish", Plural: "publishes"},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := pipeline.LoadBuildPipeline(tt.ymlcontent)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadBuildPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("LoadBuildPipeline() %v", diff)
			}
		})
	}
}

// nolint: funlen
func TestBuildPipeline_Run(t *testing.T) {
	var reportCounter int

	var reportCounterLock sync.RWMutex

	reporter := func() {
		reportCounterLock.Lock()
		reportCounter++
		reportCounterLock.Unlock()
	}

	for _, stage := range []string{"setup", "build", "publish"} {
		modules.RegisterModule(&modules.ModuleRegistration{
			Stage: stage,
			Type:  "success",
			Factory: func() modules.Pluggable {
				return &testModuleRegistration{reporter: reporter}
			},
		})
		modules.RegisterModule(&modules.ModuleRegistration{
			Stage:   stage,
			Type:    "failure",
			Factory: testFailingModuleRegistrationFactory,
		})
	}

	buildStage := func(name, plural string, types ...string) *pipeline.Stage {
		stg := pipeline.NewStage(name, plural)

		if len(types) > 0 {
			stg.Modules = []*modules.Module{}

			for _, modType := range types {
				modFactory, ok := modules.LookupModule(
					fmt.Sprintf("%s:%s", name, modType),
				)
				if ok {
					stg.Modules = append(
						stg.Modules,
						&modules.Module{Type: modType, Pluggable: modFactory()},
					)
				}
			}
		}

		return stg
	}

	tests := []struct {
		name       string
		Pipeline   *pipeline.Pipeline
		wantReport int
		wantErr    bool
	}{
		{
			name: "empty",
			Pipeline: &pipeline.Pipeline{
				Stages: []*pipeline.Stage{
					buildStage("setup", "setups"),
					buildStage("build", "builds"),
					buildStage("publish", "publishes"),
				},
			},
			wantReport: 0,
			wantErr:    false,
		},
		{
			name: "success",
			Pipeline: &pipeline.Pipeline{
				Stages: []*pipeline.Stage{
					buildStage("setup", "setups", "success"),
					buildStage("build", "builds", "success"),
					buildStage("publish", "publishes", "success"),
				},
			},
			wantReport: 3,
			wantErr:    false,
		},
		{
			name: "has error",
			Pipeline: &pipeline.Pipeline{
				Stages: []*pipeline.Stage{
					buildStage("setup", "setups", "success"),
					buildStage("build", "builds", "failure"),
					buildStage("publish", "publishes", "success"),
				},
			},
			wantReport: 1,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			reportCounter = 0
			err := tt.Pipeline.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildPipeline.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if reportCounter != tt.wantReport {
				t.Errorf(
					"BuildPipeline.Run() invoked steps %d time%s, want %d",
					reportCounter,
					map[bool]string{false: "s", true: ""}[reportCounter == 1],
					tt.wantReport,
				)
			}
		})
	}
}
