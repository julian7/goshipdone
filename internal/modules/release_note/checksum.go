package release_note

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/julian7/magelib/ctx"
	"github.com/julian7/magelib/modules"
)

// Checksum calculates checksums of artifacts, and stores them in a checksum file
type Checksum struct {
	// Algorithm specifies checksum algorithm
	Algorithm HashAlgorithm
	// Builds specifies a build names to find related artifacts to
	// calculate checksums of.
	Builds []string
	// Name specifies the checksum's name, as it stores in artifacts.
	// Default: "checksum"
	Name string
	// Output is where the checksum file is going to be created
	// Default: "{{.ProjectName}}-{{.Version}}-checksums.txt"
	Output string
	// Skip specifies which os-arch items should be skipped
	Skip []string
}

func init() {
	modules.RegisterModule(&modules.PluggableModule{
		Kind:    "release_note:checksum",
		Factory: NewChecksum,
	})
}

func NewChecksum() modules.Pluggable {
	algo, _ := NewHashAlgorithm("sha256")

	return &Checksum{
		Algorithm: *algo,
		Builds:    []string{"artifact"},
		Name:      "checksum",
		Output:    "{{.ProjectName}}-{{.Version}}-checksums.txt",
	}
}

func (checksum *Checksum) Run(context *ctx.Context) error {
	output, err := checksum.parseOutput(context)
	if err != nil {
		return fmt.Errorf("generating checksum filename: %w", err)
	}

	checksumFilename := path.Join(context.TargetDir, output)

	artifactMap := context.Artifacts.OsArchByNames(checksum.Builds, checksum.Skip)
	if len(artifactMap) == 0 {
		return nil
	}

	checksums := []string{}

	hasher := checksum.Algorithm.Factory()

	for osarch := range artifactMap {
		for _, artifact := range *artifactMap[osarch] {
			hasher.Reset()

			f, err := os.Open(artifact.Location)
			if err != nil {
				return fmt.Errorf("checksumming %s: %w", artifact.Location, err)
			}

			if _, err := io.Copy(hasher, f); err != nil {
				return fmt.Errorf("reading %s for checksumming: %w", artifact.Location, err)
			}

			if err := f.Close(); err != nil {
				return fmt.Errorf("closing %s after checksumming: %w", artifact.Location, err)
			}

			checksums = append(checksums, fmt.Sprintf(
				"%x  %s",
				hasher.Sum(nil),
				artifact.Filename,
			))
		}
	}

	writer, err := os.Create(checksumFilename)
	if err != nil {
		return fmt.Errorf("opening checksum file for writing: %w", err)
	}

	defer writer.Close()

	for _, line := range checksums {
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return fmt.Errorf("writing checksum file: %w", err)
		}
	}

	context.Artifacts.Add(&ctx.Artifact{
		Filename: output,
		Location: checksumFilename,
		Name:     checksum.Name,
	})

	log.Printf("checksum file %s written", checksumFilename)

	return nil
}

func (checksum *Checksum) parseOutput(context *ctx.Context) (string, error) {
	td := &modules.TemplateData{
		Algo:        checksum.Algorithm.String(),
		ProjectName: context.ProjectName,
		Version:     context.Version,
	}

	output, err := td.Parse(
		fmt.Sprintf("checksum-%s", checksum.Name),
		checksum.Output,
	)
	if err != nil {
		return "", fmt.Errorf("rendering %q: %w", checksum.Output, err)
	}

	return output, nil
}
