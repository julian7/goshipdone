package modules

import (
	"fmt"
	"path"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type (
	// Go represents build:go module
	Go struct {
		// Env carries a map of environment variables given to
		// `go build` command. GOOS and GOARCH will be automatically
		// added.
		Env map[string]string
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
		Env     map[string]string
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
		Deps:    []string{"setup:git_tag"},
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
func (build *Go) Run(context *ctx.Context) error {
	targets, err := build.targets(context)

	if err != nil {
		return err
	}

	for _, tar := range targets {
		if err := tar.Run(context); err != nil {
			return err
		}
	}

	return nil
}

func (build *Go) targets(context *ctx.Context) ([]modules.Pluggable, error) {
	targets := []modules.Pluggable{}

	for _, goos := range build.GOOS {
	NextArch:
		for _, goarch := range build.GOArch {
			osarch := fmt.Sprintf("%s-%s", goos, goarch)
			for _, skip := range build.Skip {
				if osarch == skip {
					continue NextArch
				}
			}

			target, err := build.singleTarget(context, goos, goarch)

			if err != nil {
				return nil, err
			}

			targets = append(targets, target)
		}
	}

	return targets, nil
}

func (build *Go) singleTarget(context *ctx.Context, goos, goarch string) (modules.Pluggable, error) {
	td := &modules.TemplateData{
		Arch:        goarch,
		ProjectName: context.ProjectName,
		OS:          goos,
		Version:     context.Version,
	}

	if goos == "windows" {
		td.Ext = ".exe"
	}

	tar := &goSingleTarget{
		Arch: goarch,
		Env:  map[string]string{},
		ID:   build.ID,
		Main: build.Main,
		OS:   goos,
	}

	for key, val := range build.Env {
		tar.Env[key] = val
	}

	tar.Env["GOOS"] = goos
	tar.Env["GOARCH"] = goarch

	var err error

	tasks := []struct {
		name   string
		source string
		target *string
	}{
		{"ldflags", build.LDFlags, &tar.LDFlags},
		{"location", path.Join(
			context.TargetDir,
			"{{.ProjectName}}-{{.OS}}-{{.Arch}}"), &tar.OutDir},
		{"output", build.Output, &tar.Output},
	}

	for _, item := range tasks {
		(*item.target), err = td.Parse("build:go", item.source)
		if err != nil {
			return nil, fmt.Errorf("cannot render %s: %w", item.name, err)
		}
	}

	return tar, nil
}

func (tar *goSingleTarget) Run(context *ctx.Context) error {
	output := path.Join(tar.OutDir, tar.Output)
	err := sh.RunWith(tar.Env, mg.GoCmd(), "build", "-o", output, "-ldflags", tar.LDFlags, tar.Main)
	if err != nil {
		_ = sh.Rm(output)
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
