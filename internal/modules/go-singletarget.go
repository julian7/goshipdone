package modules

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
	"github.com/julian7/withenv"
)

var ErrSkippedTarget = errors.New("target is skipped")

type goSingleTarget struct {
	mod     *Go
	Env     *withenv.Env
	ID      string
	LDFlags string
	OutDir  string
	Main    string
	osarch  *ctx.OsArch
	Output  string
}

func (mod *Go) newSingleTarget(goos, goarch string, goarm int32) *goSingleTarget {
	return &goSingleTarget{
		mod:    mod,
		Env:    withenv.New(),
		ID:     mod.ID,
		Main:   mod.Main,
		osarch: &ctx.OsArch{OS: goos, Arch: goarch, ArmVersion: goarm},
	}
}

func (tar *goSingleTarget) Setup(cx context.Context) error {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

	for key, val := range context.Env.Vars {
		tar.Env.Set(key, val)
	}

	tar.SetGoEnv()

	osarch := tar.OSArch()
	for _, skip := range tar.mod.Skip {
		if osarch == skip {
			return ErrSkippedTarget
		}
	}

	tasks := []struct {
		name   string
		source string
		target *string
	}{
		{"ldflags", tar.mod.LDFlags, &tar.LDFlags},
		{"location", path.Join(
			context.TargetDir,
			"{{.ProjectName}}-{{OS}}-{{ArchName}}"), &tar.OutDir},
		{"output", tar.mod.Output, &tar.Output},
	}

	td, err := modules.NewTemplate(cx)
	if err != nil {
		return err
	}

	td.OSArch = tar.osarch

	for _, item := range tasks {
		(*item.target), err = td.Parse("build:go", item.source)
		if err != nil {
			return fmt.Errorf("cannot render %s: %w", item.name, err)
		}
	}

	return nil
}

func (tar *goSingleTarget) SetGoEnv() {
	tar.Env.Set("GOOS", tar.osarch.OS)
	tar.Env.Set("GOARCH", tar.osarch.Arch)

	if tar.osarch.ArmVersion != 0 {
		tar.Env.Set("GOARM", strconv.Itoa(int(tar.osarch.ArmVersion)))
	}
}

func (tar *goSingleTarget) OSArch() string {
	return tar.osarch.String()
}

func (tar *goSingleTarget) ArchName() string {
	return tar.osarch.ArchName()
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
		Filename: tar.Output,
		Location: output,
		ID:       tar.ID,
		OsArch:   tar.osarch,
	})

	return nil
}
