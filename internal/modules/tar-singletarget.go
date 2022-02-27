package modules

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
)

type tarSingleTarget struct {
	CommonDir   string
	Compression Compression
	DirsWritten map[string]bool
	Files       []string
	ID          string
	OSArch      *ctx.OsArch
	Output      string
	Targets     *ctx.Artifacts
}

func (mod *Tar) singleTarget(cx context.Context, artifacts *ctx.Artifacts) (*tarSingleTarget, error) {
	art := (*artifacts)[0]
	ret := &tarSingleTarget{
		Compression: mod.Compression,
		DirsWritten: map[string]bool{},
		Files:       make([]string, len(mod.Files)),
		ID:          mod.ID,
		OSArch: &ctx.OsArch{
			Arch:       art.OS,
			ArmVersion: art.ArmVersion,
			OS:         art.OS,
		},
		Targets: artifacts,
	}
	for i := range mod.Files {
		ret.Files[i] = mod.Files[i]
	}

	td, err := modules.NewTemplate(cx)
	if err != nil {
		return nil, err
	}

	td.OSArch = ret.OSArch

	td.Ext = mod.Compression.Extension()

	for _, task := range []struct {
		name   string
		source string
		target *string
	}{
		{"commondir", mod.CommonDir, &ret.CommonDir},
		{"output", mod.Output, &ret.Output},
	} {
		var err error

		*task.target, err = td.Parse("archive:tar", task.source)
		if err != nil {
			return nil, fmt.Errorf("rendering %q: %w", task.source, err)
		}
	}

	ret.Output = path.Clean(ret.Output)

	return ret, nil
}

func (target *tarSingleTarget) Run(cx context.Context) error {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

	archiveFile := path.Join(context.TargetDir, target.Output)

	archive, err := os.Create(archiveFile)
	if err != nil {
		return fmt.Errorf("cannot create archive file %s: %w", archiveFile, err)
	}

	defer archive.Close()

	compressedArchive := target.Compression.Writer(archive)
	defer compressedArchive.Close()

	tw := tar.NewWriter(compressedArchive)
	defer tw.Close()

	for _, artifact := range *target.Targets {
		if err := target.writeArtifact(tw, artifact); err != nil {
			return fmt.Errorf("writing %s: %w", archiveFile, err)
		}
	}

	for _, file := range target.Files {
		if err := target.writeFileGlob(tw, file); err != nil {
			return fmt.Errorf("writing %s: %w", archiveFile, err)
		}
	}

	context.Artifacts.Add(&ctx.Artifact{
		Arch:       target.OSArch.Arch,
		ArchName:   target.OSArch.ArchName(),
		ArmVersion: target.OSArch.ArmVersion,
		Filename:   target.Output,
		Location:   archiveFile,
		ID:         target.ID,
		OS:         target.OSArch.OS,
	})

	return nil
}

func (target *tarSingleTarget) writeArtifact(tw *tar.Writer, artifact *ctx.Artifact) error {
	filename := path.Join(target.CommonDir, artifact.Filename)
	if err := target.writeDirs(tw, path.Dir(filename)); err != nil {
		return err
	}

	return target.writeFile(tw, filename, artifact.Location)
}

func (target *tarSingleTarget) writeFileGlob(tw *tar.Writer, source string) error {
	matches, err := filepath.Glob(source)
	if err != nil {
		return err
	}

	for _, filename := range matches {
		fullfn := path.Join(target.CommonDir, filename)
		if err := target.writeDirs(tw, path.Dir(fullfn)); err != nil {
			return err
		}

		if err := target.writeFile(tw, fullfn, filename); err != nil {
			return err
		}
	}

	return nil
}

func (target *tarSingleTarget) writeFile(tw *tar.Writer, destpath, source string) error {
	fi, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("can't stat file %s: %w", source, err)
	}

	hdr, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}

	hdr.Name = destpath

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	sourceReader, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourceReader.Close()

	buf := make([]byte, 4096)

	if n, err := io.CopyBuffer(tw, sourceReader, buf); err != nil {
		return fmt.Errorf(
			"copying %s to archive %s (%d bytes written, %d bytes reported): %w",
			source,
			destpath,
			n,
			fi.Size(),
			err,
		)
	}

	return nil
}

func (target *tarSingleTarget) writeDirs(tw *tar.Writer, fullpath string) error {
	if fullpath == "." {
		return nil
	}

	dirs := []string{fullpath}

	for {
		fullpath = path.Dir(fullpath)
		if fullpath == "." {
			break
		}

		dirs = append(dirs, fullpath)
	}

	for i := range dirs {
		dirname := dirs[len(dirs)-i-1]

		if err := target.writeDir(tw, dirname); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", dirname, err)
		}
	}

	return nil
}

func (target *tarSingleTarget) writeDir(tw *tar.Writer, dirname string) error {
	if _, ok := target.DirsWritten[dirname]; ok {
		return nil
	}

	st, err := os.Stat(path.Dir(target.Output))
	if err != nil {
		return err
	}

	hdr, err := tar.FileInfoHeader(st, "")
	if err != nil {
		return err
	}

	hdr.Name = dirname + "/"

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	target.DirsWritten[dirname] = true

	return nil
}
