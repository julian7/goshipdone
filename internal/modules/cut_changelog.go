package modules

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"regexp"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/modules"
)

const (
	reCut             = `(?ms)^(## \[(?i:%s)\].+?)\n(?:## |\[.+?\]:)`
	reHasSplitLinks   = `^## \[.+?\][^:]`
	reLinkResolutions = `(?m)^\[(.+?)\]:\s*(.+)$`
	reDuplicateLinks  = `(?s)(\[.+\]\(.+?\))\(.+?\)`
)

type CutChangelog struct {
	// ID is the artifact ID of the changelog slice other modules will be
	// able to refer to. Default: "changelog".
	ID string
	// Input points to the original changelog file this module can take a
	// slice of. It must be in https://keepachangelog.org/ format. Default:
	// "CHANGELOG.md".
	Input string
	// Output is the filename of the changelog slice under Dist folder.
	// If empty, it will be the same as Input.
	// Default: "".
	Output string
}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "build",
		Type:    "changelog",
		Factory: NewCutChangelog,
	})
}

func NewCutChangelog() modules.Pluggable {
	return &CutChangelog{
		ID:     "changelog",
		Input:  "CHANGELOG.md",
		Output: "",
	}
}

func (mod *CutChangelog) Run(context *ctx.Context) error {
	contents, err := ioutil.ReadFile(mod.Input)
	if err != nil {
		return fmt.Errorf("reading original CHANGELOG %s: %w", mod.Input, err)
	}

	ver := context.Git.Tag
	if ver == "" {
		ver = "unreleased"
	}

	cut := regexp.MustCompile(fmt.Sprintf(reCut, regexp.QuoteMeta(ver)))

	matches := cut.FindSubmatch(contents)
	if len(matches) != 2 {
		return fmt.Errorf("cannot detect changelog segment for %s", ver)
	}

	if regexp.MustCompile(reHasSplitLinks).Match(matches[1]) {
		linkMatches := regexp.MustCompile(reLinkResolutions).FindAllSubmatch(contents, -1)

		for _, link := range linkMatches {
			matches[1] = bytes.ReplaceAll(
				matches[1],
				[]byte(fmt.Sprintf("[%s]", string(link[1]))),
				[]byte(fmt.Sprintf("[%s](%s)", string(link[1]), string(link[2]))),
			)
		}
	}

	matches[1] = regexp.MustCompile(reDuplicateLinks).ReplaceAll(matches[1], []byte("$1"))

	outfile := mod.Output
	if outfile == "" {
		outfile = mod.Input
	}

	outfile = path.Join(context.TargetDir, outfile)

	if err := ioutil.WriteFile(outfile, matches[1], 0o644); err != nil {
		return fmt.Errorf("writing sliced CHANGELOG %s: %w", outfile, err)
	}

	context.Artifacts.Add(&ctx.Artifact{
		ID:       mod.ID,
		Filename: mod.Input,
		Location: outfile,
	})

	return nil
}
