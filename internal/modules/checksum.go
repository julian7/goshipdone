package modules

import (
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
)

// Checksum calculates checksums of artifacts, and stores them in a checksum file
type Checksum struct {
	// Algorithm specifies checksum algorithm
	Algorithm HashAlgorithm
	// Builds specifies a build names to find related artifacts to
	// calculate checksums of.
	Builds []string
	// ID specifies the checksum's name, as it stores in artifacts.
	// Default: "checksum"
	ID string
	// Output is where the checksum file is going to be created
	// Default: "{{.ProjectName}}-{{.Version}}-checksums.txt"
	Output string
	// Skip specifies which os-arch items should be skipped
	Skip []string
}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "build",
		Type:    "checksum",
		Factory: NewChecksum,
	})
}

func NewChecksum() modules.Pluggable {
	algo, _ := NewHashAlgorithm("sha256")

	return &Checksum{
		Algorithm: *algo,
		Builds:    []string{"artifact"},
		ID:        "checksum",
		Output:    "{{.ProjectName}}-{{.Version}}-checksums.txt",
	}
}

func (checksum *Checksum) Run(context *ctx.Context) error {
	output, err := checksum.parseOutput(context)
	if err != nil {
		return fmt.Errorf("generating checksum filename: %w", err)
	}

	checksumFilename := path.Join(context.TargetDir, output)

	artifactMap := context.Artifacts.OsArchByIDs(checksum.Builds, checksum.Skip)
	if len(artifactMap) == 0 {
		return nil
	}

	checksums := []string{}

	hasher := checksum.Algorithm.Factory()

	for osarch := range artifactMap {
		for _, artifact := range *artifactMap[osarch] {
			sum, err := checksumArtifact(hasher, artifact)
			if err != nil {
				return err
			}

			checksums = append(checksums, sum)
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
		ID:       checksum.ID,
	})

	log.Printf("checksum file %s written", checksumFilename)

	return nil
}

func checksumArtifact(hasher hash.Hash, artifact *ctx.Artifact) (string, error) {
	hasher.Reset()

	f, err := os.Open(artifact.Location)
	if err != nil {
		return "", fmt.Errorf("checksumming %s: %w", artifact.Location, err)
	}

	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("reading %s for checksumming: %w", artifact.Location, err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("closing %s after checksumming: %w", artifact.Location, err)
	}

	return fmt.Sprintf(
		"%x  %s",
		hasher.Sum(nil),
		artifact.Filename,
	), nil
}

func (checksum *Checksum) parseOutput(context *ctx.Context) (string, error) {
	td := modules.NewTemplate(context)
	td.Algo = checksum.Algorithm.String()

	output, err := td.Parse("checksum", checksum.Output)
	if err != nil {
		return "", fmt.Errorf("rendering %q: %w", checksum.Output, err)
	}

	return output, nil
}
