package modules

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
	"github.com/julian7/withenv"
)

type (
	// Go represents build:go module
	Go struct {
		// After is a list of commands have to be ran after builds.
		// Any errors cancel the task.
		After []string
		// Before is a list of commands have to be ran before builds.
		// Any errors cancel the task.
		Before []string
		// GOOS is a list of all GOOS variations required. It is
		// set to [`windows`, `linux`] by default.
		GOOS []string
		// GOArch is a list of all GOARCH variations required. It is
		// set to [`amd64`] by default.
		GOArch []string
		// ID contains the artifact's name used by later stages of the build
		// pipeline. Archives, and Publishes may refer to this name for
		// referencing build results.
		// Default: "default".
		ID string
		// LDFlags is a `modules.TemplateData` template for providing
		// `-ldflags` configuration option to `go build` command.
		// It defaults to `-s -w -X main.version={{.Version}}`.
		LDFlags string
		// Main designates the file / directory where `main` package
		// (as well as `main` function) is defined.
		Main string
		// Output is where the build writes its output. Default:
		// `{{.ProjectName}}{{.Ext}}`
		Output string
		// Skip specifies GOOS-GOArch combinations to be skipped.
		// They are in `{{.Os}}-{{.Arch}}` format.
		//
		// Eg.
		//
		// ```go
		// Go{
		//     GOOS: []string{"linux", "windows"},
		//     GOArch: []string{"amd64", "386"},
		//     Skip: []string{"linux-386"},
		// }
		// ```
		//
		// will run builds for linux-amd64, windows-amd64, and windows-386 only.
		Skip []string
	}

	goSingleTarget struct {
		Arch    string
		Env     *withenv.Env
		ID      string
		LDFlags string
		OutDir  string
		Main    string
		OS      string
		Output  string
	}
)

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "build",
		Type:    "go",
		Factory: NewGo,
	})
}

// NewGo is a Go struct factory
func NewGo() modules.Pluggable {
	return &Go{
		LDFlags: "-s -w -X main.version={{.Version}}",
		GOOS:    []string{"linux", "windows"},
		GOArch:  []string{"amd64"},
		Main:    ".",
		ID:      "default",
		Output:  "{{.ProjectName}}{{.Ext}}",
	}
}

// Run executes a go build step
func (mod *Go) Run(cx context.Context) error {
	targets, err := mod.targets(cx)

	if err != nil {
		return err
	}

	if err := mod.runHooks(cx, mod.Before); err != nil {
		return err
	}

	for _, tar := range targets {
		if err := tar.Run(cx); err != nil {
			return err
		}
	}

	if err := mod.runHooks(cx, mod.After); err != nil {
		return err
	}

	return nil
}

func (mod *Go) runHooks(cx context.Context, hooks []string) error {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

	if len(hooks) == 0 {
		return nil
	}

	for _, hook := range hooks {
		args := strings.Fields(hook)
		if err := context.Env.Run(args[0], args[1:]...); err != nil {
			return err
		}
	}

	return nil
}

func (mod *Go) targets(cx context.Context) ([]modules.Pluggable, error) {
	targets := []modules.Pluggable{}

	for _, goos := range mod.GOOS {
	NextArch:
		for _, goarch := range mod.GOArch {
			osarch := fmt.Sprintf("%s-%s", goos, goarch)
			for _, skip := range mod.Skip {
				if osarch == skip {
					continue NextArch
				}
			}

			target, err := mod.singleTarget(cx, goos, goarch)

			if err != nil {
				return nil, err
			}

			targets = append(targets, target)
		}
	}

	return targets, nil
}

func (mod *Go) singleTarget(cx context.Context, goos, goarch string) (modules.Pluggable, error) {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return nil, err
	}

	td, err := modules.NewTemplate(cx)
	if err != nil {
		return nil, err
	}

	td.Arch = goarch
	td.OS = goos

	if goos == "windows" {
		td.Ext = ".exe"
	}

	tar := &goSingleTarget{
		Arch: goarch,
		Env:  withenv.New(),
		ID:   mod.ID,
		Main: mod.Main,
		OS:   goos,
	}

	for key, val := range context.Env.Vars {
		tar.Env.Set(key, val)
	}

	tar.Env.Set("GOOS", goos)
	tar.Env.Set("GOARCH", goarch)

	tasks := []struct {
		name   string
		source string
		target *string
	}{
		{"ldflags", mod.LDFlags, &tar.LDFlags},
		{"location", path.Join(
			context.TargetDir,
			"{{.ProjectName}}-{{.OS}}-{{.Arch}}"), &tar.OutDir},
		{"output", mod.Output, &tar.Output},
	}

	for _, item := range tasks {
		(*item.target), err = td.Parse("build:go", item.source)
		if err != nil {
			return nil, fmt.Errorf("cannot render %s: %w", item.name, err)
		}
	}

	return tar, nil
}

func (tar *goSingleTarget) Run(cx context.Context) error {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

	output := path.Join(tar.OutDir, tar.Output)

	if err := tar.Env.Run("go", "build", "-o", output, "-ldflags", tar.LDFlags, tar.Main); err != nil {
		_ = os.Remove(output)
		return err
	}

	context.Artifacts.Add(&ctx.Artifact{
		Arch:     tar.Arch,
		Filename: tar.Output,
		Location: output,
		ID:       tar.ID,
		OS:       tar.OS,
	})

	return nil
}
