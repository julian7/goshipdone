package pipeline_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/go-test/deep"
	"github.com/julian7/magelib/ctx"
	intmod "github.com/julian7/magelib/internal/modules"
	"github.com/julian7/magelib/modules"
	"github.com/julian7/magelib/pipeline"
)

type testPluggableModule struct {
	reporter func()
}

func testPluggableModuleFactory() modules.Pluggable {
	return &testPluggableModule{}
}

func (mod *testPluggableModule) Run(*ctx.Context) error {
	if mod.reporter != nil {
		mod.reporter()
	}

	return nil
}

type testFailingPluggableModule struct{}

func testFailingPluggableModuleFactory() modules.Pluggable {
	return &testFailingPluggableModule{}
}

func (mod *testFailingPluggableModule) Run(*ctx.Context) error {
	return errors.New("error")
}

// nolint: funlen
func TestLoadBuildPipeline(t *testing.T) {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "archive:test",
		Factory: testPluggableModuleFactory,
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
						{Type: "project", Pluggable: intmod.NewProject()},
						{Type: "git_tag", Pluggable: intmod.NewGitTag()},
					},
				},
				Builds:       &modules.Modules{Stage: "build"},
				Archives:     &modules.Modules{Stage: "archive"},
				ReleaseNotes: &modules.Modules{Stage: "release_note"},
				Publishes:    &modules.Modules{Stage: "publish"},
			},
			false,
		},
		{
			"invoked mod loads",
			[]byte("---\narchives:\n  - type: test\n"),
			&pipeline.BuildPipeline{
				Setups: &modules.Modules{
					Stage: "setup",
					Modules: []modules.Module{
						{Type: "project", Pluggable: intmod.NewProject()},
						{Type: "git_tag", Pluggable: intmod.NewGitTag()},
					},
				},
				Builds: &modules.Modules{Stage: "build"},
				Archives: &modules.Modules{Stage: "archive", Modules: []modules.Module{
					{Type: "test", Pluggable: testPluggableModuleFactory()},
				}},
				ReleaseNotes: &modules.Modules{Stage: "release_note"},
				Publishes:    &modules.Modules{Stage: "publish"},
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

	for _, stage := range []string{"setup", "build", "archive", "release_note", "publish"} {
		modules.RegisterModule(&modules.PluggableModule{
			Kind: fmt.Sprintf("%s:success", stage),
			Factory: func() modules.Pluggable {
				return &testPluggableModule{reporter: reporter}
			},
		})
		modules.RegisterModule(&modules.PluggableModule{
			Kind:    fmt.Sprintf("%s:failure", stage),
			Factory: testFailingPluggableModuleFactory,
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
		name         string
		Setups       modules.Modules
		Builds       modules.Modules
		Archives     modules.Modules
		ReleaseNotes modules.Modules
		Publishes    modules.Modules
		wantReport   int
		wantErr      bool
	}{
		{
			name:         "empty",
			Setups:       buildModule("setup"),
			Builds:       buildModule("build"),
			Archives:     buildModule("archive"),
			ReleaseNotes: buildModule("release_note"),
			Publishes:    buildModule("publish"),
			wantReport:   0,
			wantErr:      false,
		},
		{
			name:         "success",
			Setups:       buildModule("setup", "success"),
			Builds:       buildModule("build", "success"),
			Archives:     buildModule("archive", "success"),
			ReleaseNotes: buildModule("release_note", "success"),
			Publishes:    buildModule("publish", "success"),
			wantReport:   5,
			wantErr:      false,
		},
		{
			name:         "has error",
			Setups:       buildModule("setup", "success"),
			Builds:       buildModule("build", "success"),
			Archives:     buildModule("archive", "failure"),
			ReleaseNotes: buildModule("release_note", "success"),
			Publishes:    buildModule("publish", "success"),
			wantReport:   2,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportCounter = 0
			pipeline := &pipeline.BuildPipeline{
				Setups:       &tt.Setups,
				Builds:       &tt.Builds,
				Archives:     &tt.Archives,
				ReleaseNotes: &tt.ReleaseNotes,
				Publishes:    &tt.Publishes,
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
