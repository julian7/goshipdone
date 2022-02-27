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
	Arch    string
	Arm     int32
	Env     *withenv.Env
	ID      string
	LDFlags string
	OutDir  string
	Main    string
	OS      string
	Output  string
}

func (mod *Go) newSingleTarget(goos, goarch string, goarm int32) *goSingleTarget {
	return &goSingleTarget{
		mod:  mod,
		Arch: goarch,
		Arm:  goarm,
		Env:  withenv.New(),
		ID:   mod.ID,
		Main: mod.Main,
		OS:   goos,
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

	osarch := tar.OsArch()
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
			"{{.ProjectName}}-{{.OS}}-{{.ArchName}}"), &tar.OutDir},
		{"output", tar.mod.Output, &tar.Output},
	}

	td, err := modules.NewTemplate(cx)
	if err != nil {
		return err
	}

	tar.SetTemplateOsArch(td)

	for _, item := range tasks {
		(*item.target), err = td.Parse("build:go", item.source)
		if err != nil {
			return fmt.Errorf("cannot render %s: %w", item.name, err)
		}
	}

	return nil
}

func (tar *goSingleTarget) SetTemplateOsArch(td *modules.TemplateData) {
	td.Arch = tar.Arch
	td.ArmVersion = tar.Arm
	td.ArchName = tar.ArchName()
	td.OS = tar.OS

	if tar.OS == "windows" {
		td.Ext = ".exe"
	}
}

func (tar *goSingleTarget) SetGoEnv() {
	tar.Env.Set("GOOS", tar.OS)
	tar.Env.Set("GOARCH", tar.Arch)

	if tar.Arm != 0 {
		tar.Env.Set("GOARM", strconv.Itoa(int(tar.Arm)))
	}
}

func (tar *goSingleTarget) OsArch() string {
	return fmt.Sprintf("%s-%s", tar.OS, tar.ArchName())
}

func (tar *goSingleTarget) ArchName() string {
	if tar.Arm > 0 {
		return fmt.Sprintf("%sv%d", tar.Arch, tar.Arm)
	}

	return tar.Arch
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
		Arch:       tar.Arch,
		ArchName:   tar.ArchName(),
		ArmVersion: tar.Arm,
		Filename:   tar.Output,
		Location:   output,
		ID:         tar.ID,
		OS:         tar.OS,
	})

	return nil
}
