package modules

import (
	"bytes"
	"text/template"
)

// TemplateData is the data all template-based text replacement takes place.
// Modules are responsible of filling in the appropriate fields, and handle
// dependencies (for example, ArchiveName is often depends on ProjectName, therefore
// modules have to render them in ProjectName then ArchiveName order)
type TemplateData struct {
	// Arch defines target architecture
	Arch string
	// ArchiveName defines a URL where the resource will be remotely available
	ArchiveName string
	// OS defines target operating system
	OS string
	// ProjectName defines local filename of the resource
	ProjectName string
	// Version defines artifact's version
	Version string
	// Ext contains executable extension
	Ext string
}

// Parse parses a string based on TemplateData, and returns output in string format
func (td *TemplateData) Parse(name, text string) (string, error) {
	tmpl := template.New(name)
	_, err := tmpl.Parse(text)

	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, td); err != nil {
		return "", err
	}

	return out.String(), nil
}
