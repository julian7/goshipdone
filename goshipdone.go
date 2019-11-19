// Package goshipdone provides an extendable, build-to-publish pipeline for multiple
// OS-Architecture targets for go packages, built for magefile.
package goshipdone

import (
	"fmt"
	"os"

	"github.com/julian7/goshipdone/pipeline"
	"github.com/spf13/afero"
)

const (
	filenameEnv          = "GOSHIPDONE_CONFIG"
	defaultFilename      = ".goshipdone.yml"
	defaultLocalFilename = ".goshipdone.local.yml"
)

// nolint: gochecknoglobals
var defaultFS = afero.NewOsFs()

// Run executes all steps in the pipeline, defined by YAML configuration file.
// It tries to load files from the following sources:
// - provided filename
// - GOSHIPDONE_CONFIG environment variable
// - .goshipdone.local.yml (if exists)
// - .goshipdone.yml
//
// It returns an error if any of the subsequent processing has an error.
func Run(filename string) error {
	filename = detectFilename(filename)

	content, err := afero.ReadFile(defaultFS, filename)
	if err != nil {
		return fmt.Errorf("loading GoShipDone file: %w", err)
	}

	pipe, err := pipeline.LoadBuildPipeline(content)
	if err != nil {
		return fmt.Errorf("processing GoShipDone file: %w", err)
	}

	if err := pipe.Run(); err != nil {
		return fmt.Errorf("running GoShipDone: %w", err)
	}

	return nil
}

func detectFilename(filename string) string {
	if filename != "" {
		return filename
	}

	if fn, ok := os.LookupEnv(filenameEnv); ok {
		return fn
	}

	if st, err := defaultFS.Stat(defaultLocalFilename); err == nil && st.Mode().IsRegular() {
		return defaultLocalFilename
	}

	return defaultFilename
}
