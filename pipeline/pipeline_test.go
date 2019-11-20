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
		want       *pipeline.BuildPipeline
		wantErr    bool
	}{
		{
			"empty loads defaults",
			[]byte("---\n"),
			&pipeline.BuildPipeline{
				Setups: &modules.Modules{
					Stage: "setup",
					Modules: []modules.Module{
						{Type: "env", Pluggable: intmod.NewEnv()},
						{Type: "project", Pluggable: intmod.NewProject()},
						{Type: "git_tag", Pluggable: intmod.NewGit()},
						{Type: "skip_publish", Pluggable: intmod.NewSkipPublish()},
					},
				},
				Builds:    &modules.Modules{Stage: "build"},
				Publishes: &modules.Modules{Stage: "publish"},
			},
			false,
		},
		{
			"invoked mod loads",
			[]byte("---\nbuilds:\n  - type: test\n"),
			&pipeline.BuildPipeline{
				Setups: &modules.Modules{
					Stage: "setup",
					Modules: []modules.Module{
						{Type: "env", Pluggable: intmod.NewEnv()},
						{Type: "project", Pluggable: intmod.NewProject()},
						{Type: "git_tag", Pluggable: intmod.NewGit()},
						{Type: "skip_publish", Pluggable: intmod.NewSkipPublish()},
					},
				},
				Builds: &modules.Modules{Stage: "build", Modules: []modules.Module{
					{Type: "test", Pluggable: testModuleRegistrationFactory()},
				}},
				Publishes: &modules.Modules{Stage: "publish"},
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

	buildModule := func(stage string, types ...string) modules.Modules {
		mods := modules.Modules{Stage: stage}

		if len(types) > 0 {
			mods.Modules = []modules.Module{}

			for _, modType := range types {
				modFactory, ok := modules.LookupModule(
					fmt.Sprintf("%s:%s", stage, modType),
				)
				if ok {
					mods.Modules = append(
						mods.Modules,
						modules.Module{Type: modType, Pluggable: modFactory()},
					)
				}
			}
		}

		return mods
	}

	tests := []struct {
		name       string
		Setups     modules.Modules
		Builds     modules.Modules
		Publishes  modules.Modules
		wantReport int
		wantErr    bool
	}{
		{
			name:       "empty",
			Setups:     buildModule("setup"),
			Builds:     buildModule("build"),
			Publishes:  buildModule("publish"),
			wantReport: 0,
			wantErr:    false,
		},
		{
			name:       "success",
			Setups:     buildModule("setup", "success"),
			Builds:     buildModule("build", "success"),
			Publishes:  buildModule("publish", "success"),
			wantReport: 3,
			wantErr:    false,
		},
		{
			name:       "has error",
			Setups:     buildModule("setup", "success"),
			Builds:     buildModule("build", "failure"),
			Publishes:  buildModule("publish", "success"),
			wantReport: 1,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			reportCounter = 0
			pipeline := &pipeline.BuildPipeline{
				Setups:    &tt.Setups,
				Builds:    &tt.Builds,
				Publishes: &tt.Publishes,
			}
			err := pipeline.Run()
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
