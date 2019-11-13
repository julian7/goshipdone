// Package goshipdone provides an extendable, build-to-publish pipeline for multiple
// OS-Architecture targets for go packages, built for magefile.
package goshipdone

import (
	"fmt"
	"io/ioutil"

	"github.com/julian7/goshipdone/pipeline"
)

func Run(filename string) error {
	if filename == "" {
		filename = ".goshipdone.yml"
	}

	content, err := ioutil.ReadFile(filename)
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
