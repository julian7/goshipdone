package build

import (
	"fmt"
	"path"

	"github.com/julian7/magelib/ctx"
	"github.com/julian7/magelib/modules"
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
		// LDFlags is a `modules.TemplateData` template for providing
		// `-ldflags` configuration option to `go build` command.
		// It defaults to `-s -w -X main.version={{.Version}}`.
		LDFlags string
		// Main designates the file / directory where `main` package
		// (as well as `main` function) is defined.
		Main string
		// Name contains the artifact's name used by later stages of
		// the build pipeline. Archives, ReleaseNotes, and Publishes
		// may refer to this name for referencing build results.
		// Default: "default".
		Name string
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

	goRuntime struct {
		*Go
		*ctx.Context
	}

	goSingleTarget struct {
		*ctx.Context
		Arch    string
		Env     map[string]string
		LDFlags string
		Main    string
		Name    string
		OS      string
		Output  string
	}
)

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "build:go",
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
		Name:    "default",
		Output:  "{{.ProjectName}}{{.Ext}}",
	}
}

// Run executes a go build step
func (build *Go) Run(context *ctx.Context) error {
	rt := &goRuntime{Go: build, Context: context}

	targets, err := rt.targets()

	if err != nil {
		return err
	}

	for _, tar := range targets {
		if err := tar.build(); err != nil {
			return err
		}
	}

	return nil
}

func (rt *goRuntime) targets() ([]*goSingleTarget, error) {
	targets := []*goSingleTarget{}

	for _, goos := range rt.Go.GOOS {
	NextArch:
		for _, goarch := range rt.Go.GOArch {
			osarch := fmt.Sprintf("%s-%s", goos, goarch)
			for _, skip := range rt.Go.Skip {
				if osarch == skip {
					continue NextArch
				}
			}

			target, err := rt.buildSingleTarget(goos, goarch)

			if err != nil {
				return nil, err
			}

			targets = append(targets, target)
		}
	}

	return targets, nil
}

func (rt *goRuntime) buildSingleTarget(goos, goarch string) (*goSingleTarget, error) {
	td := &modules.TemplateData{
		Arch:        goarch,
		ProjectName: rt.Context.ProjectName,
		OS:          goos,
		Version:     rt.Context.Version,
	}

	if goos == "windows" {
		td.Ext = ".exe"
	}

	tar := &goSingleTarget{
		Arch:    goarch,
		Env:     map[string]string{},
		Main:    rt.Go.Main,
		Name:    rt.Go.Name,
		OS:      goos,
		Context: rt.Context,
	}

	for key, val := range rt.Go.Env {
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
		{"ldflags", rt.Go.LDFlags, &tar.LDFlags},
		{"output", path.Join(
			tar.Context.TargetDir,
			"{{.ProjectName}}-{{.OS}}-{{.Arch}}",
			rt.Go.Output,
		), &tar.Output},
	}

	for _, item := range tasks {
		(*item.target), err = td.Parse(
			fmt.Sprintf("buildgo-%s-%s-%s-%s", rt.Go.Main, goos, goarch, item.name),
			item.source,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot render %s: %w", item.name, err)
		}
	}

	return tar, nil
}

func (tar *goSingleTarget) build() error {
	err := sh.RunWith(tar.Env, mg.GoCmd(), "build", "-o", tar.Output, "-ldflags", tar.LDFlags, tar.Main)
	if err != nil {
		_ = sh.Rm(tar.Output)
		return err
	}

	tar.Artifacts.Add(ctx.FormatRaw, tar.Name, tar.Output, tar.OS, tar.Arch)

	return nil
}
