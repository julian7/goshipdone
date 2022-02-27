package modules

import (
	"bytes"
	"context"
	"text/template"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/withenv"
)

// TemplateData is the data all template-based text replacement takes place.
// Modules are responsible of filling in the appropriate fields, and handle
// dependencies (for example, ArchiveName is often depends on ProjectName, therefore
// modules have to render them in ProjectName then ArchiveName order)
type TemplateData struct {
	// Algo represents algorithm. Hashing and signing modules use them.
	Algo string
	// ArchiveName defines a URL where the resource will be remotely available
	ArchiveName string
	// Env is a copy of environment variables set in ctx.Context
	Env *withenv.Env
	// Git is a copy of git-related info from ctx.Context
	Git *ctx.GitData
	// OSArch defines target operating system and architecture
	OSArch *ctx.OsArch
	// ProjectName defines local filename of the resource
	ProjectName string
	// Version defines artifact's version
	Version string
	// Ext contains executable extension
	Ext string
}

func NewTemplate(cx context.Context) (*TemplateData, error) {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return nil, err
	}

	return &TemplateData{
		Env:         context.Env,
		Git:         context.Git,
		ProjectName: context.ProjectName,
		Version:     context.Version,
	}, nil
}

// Parse parses a string based on TemplateData, and returns output in string format
func (td *TemplateData) Parse(name, text string) (string, error) {
	tmpl := template.New(name).Funcs(template.FuncMap{
		"Arch":     func() string { return td.OSArch.Arch },
		"ArchName": func() string { return td.OSArch.ArchName() },
		"OS":       func() string { return td.OSArch.OS },
		"OSExt": func() string {
			if td.OSArch.OS == "windows" {
				return ".exe"
			}

			return ""
		},
	})
	_, err := tmpl.Parse(text)

	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, td); err != nil {
		return "", err
	}

	return td.Env.Expand(out.String()), nil
}
