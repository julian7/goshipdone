package modules

import (
	"context"
	"errors"
	"strings"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
)

// Go represents build:go module
type Go struct {
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
	// GOArm is a list of all GOARM variations required. GOARM=6 is
	// used by default, as golang's internal default. Providing multiple
	// GOArm entries provides multiple builds while in GOOS=linux and
	// GOARCH=arm setting.
	GOArm []int32
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
		GOArm:   []int32{6},
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
		for _, goarch := range mod.GOArch {
			arms := []int32{0}
			if goos == "linux" && goarch == "arm" {
				arms = mod.GOArm
			}

			for _, goarm := range arms {
				target := mod.newSingleTarget(goos, goarch, goarm)

				err := target.Setup(cx)
				if err != nil {
					if errors.Is(err, ErrSkippedTarget) {
						continue
					}

					return nil, err
				}

				targets = append(targets, target)
			}
		}
	}

	return targets, nil
}
