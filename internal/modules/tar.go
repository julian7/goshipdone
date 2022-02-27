package modules

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
)

type (
	// Tar is a module for building an archive from prior builds
	Tar struct {
		// Builds specifies which build names should be added to the archive.
		Builds []string
		// CommonDir contains a common directory name for all files inside
		// the tar archive. An empty CommonDir skips creating subdirectories.
		// Default: `{{.ProjectName}}-{{.Version}}-{{OS}}-{{ArchName}}`.
		CommonDir string
		// Compression specifies which compression should be applied to the
		// archive.
		Compression Compression
		// Files contains a list of static files should be added to the
		// archive file. They are interpretered as glob.
		Files []string
		// ID contains the artifact's name used by later stages of the build
		// pipeline. Archives, and Publishes may refer to this name for
		// referencing build results.
		// Default: "archive".
		ID string
		// Output is where the build writes its output. Default:
		// `{{.ProjectName}}-{{.Version}}-{{OS}}-{{ArchName}}.tar{{.Ext}}`
		// where `{{.Ext}}` contains the compression's default extension
		Output string
		// Skip specifies GOOS-GOArch combinations to be skipped.
		// They are in `{{.Os}}-{{.Arch}}` format.
		// It filters builds to be included.
		Skip []string
	}
)

func NewTar() modules.Pluggable {
	return &Tar{
		Builds:      []string{"default"},
		CommonDir:   "{{.ProjectName}}-{{.Version}}-{{OS}}-{{ArchName}}",
		Compression: Compression{&CompressNONE{}},
		Files:       []string{"README*"},
		ID:          "archive",
		Output:      "{{.ProjectName}}-{{.Version}}-{{OS}}-{{ArchName}}.tar{{.Ext}}",
		Skip:        []string{},
	}
}

func (mod *Tar) Run(cx context.Context) error {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

	builds := context.Artifacts.OsArchByIDs(mod.Builds, mod.Skip)

	if err := validateBuilds(builds); err != nil {
		return err
	}

	for osarch := range builds {
		target, err := mod.singleTarget(cx, builds[osarch])
		if err != nil {
			return err
		}

		if err := target.Run(cx); err != nil {
			return err
		}
	}

	return nil
}

func validateBuilds(builds map[string]*ctx.Artifacts) error {
	numTargets := 0
	lastosarch := ""

	for osarch := range builds {
		targets := len(*builds[osarch])

		if numTargets == 0 {
			numTargets = targets
			lastosarch = osarch

			continue
		}

		if numTargets < targets {
			return errNumTargets(lastosarch, osarch, builds)
		}

		if numTargets > targets {
			return errNumTargets(osarch, lastosarch, builds)
		}
	}

	return nil
}

func errNumTargets(bad, good string, builds map[string]*ctx.Artifacts) error {
	targets := map[string]bool{}

	if len(builds) == 0 {
		return fmt.Errorf("no builds found")
	}

	_, ok := builds[good]
	if !ok || len(*builds[good]) == 0 {
		return fmt.Errorf("invalid reference to good artifacts: %s", good)
	}

	for _, art := range *builds[good] {
		targets[art.OsArch.String()] = true
	}

	_, ok = builds[bad]
	if ok && len(*builds[bad]) > 0 {
		for _, art := range *builds[bad] {
			targets[art.OsArch.String()] = false
		}
	}

	var goodTargets, badTargets []string

	for name := range targets {
		if targets[name] {
			goodTargets = append(goodTargets, name)
		} else {
			badTargets = append(badTargets, name)
		}
	}

	sort.Strings(goodTargets)
	sort.Strings(badTargets)

	if len(badTargets) == 0 {
		return fmt.Errorf(
			"no targets found for build%s %s",
			map[bool]string{true: "", false: "s"}[len(goodTargets) == 1],
			strings.Join(goodTargets, ", "),
		)
	}

	return fmt.Errorf(
		"build %s is missing os-arch target%s %s",
		bad,
		map[bool]string{true: "", false: "s"}[len(goodTargets) == 1],
		strings.Join(goodTargets, ", "),
	)
}
