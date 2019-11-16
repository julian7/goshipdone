# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
{{- if .Unreleased.CommitGroups }}{{ range .Unreleased.CommitGroups }}

{{ .Title }}:
{{ range .Commits }}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{- end }}
{{- end }}
{{- else }}

No changes so far.
{{- end }}
{{- if .Versions }}
{{- range .Versions }}

## [{{ .Tag.Name }}] - {{ datetime "2006-01-02" .Tag.Date }}
{{- range .CommitGroups }}

{{ .Title }}:
{{ range .Commits }}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}

[Unreleased]: {{ .Info.RepositoryURL }}/compare/{{ $latest := index .Versions 0 }}{{ $latest.Tag.Name }}...HEAD
{{- if .Versions }}
  {{- range .Versions }}
    {{- if .Tag.Previous }}
[{{ .Tag.Name }}]: {{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}
    {{- else }}
[{{ .Tag.Name }}]: {{ $.Info.RepositoryURL }}/releases/tag/{{ .Tag.Name }}
    {{- end }}
  {{- end }}
{{- end }}
