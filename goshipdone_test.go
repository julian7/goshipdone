// Package goshipdone provides an extendable, build-to-publish pipeline for multiple
// OS-Architecture targets for go packages, built for magefile.
package goshipdone

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

func Test_detectFilename(t *testing.T) {
	tests := []struct {
		name  string
		env   string
		input string
		files []string
		want  string
	}{
		{name: "only env", env: "a", input: "", want: "a"},
		{name: "only input", env: "", input: "b", want: "b"},
		{name: "both env and input", env: "a", input: "b", want: "b"},
		{
			name:  "local file exists",
			env:   "",
			input: "",
			files: []string{".goshipdone.local.yml"},
			want:  ".goshipdone.local.yml",
		},
		{
			name:  "local file doesn't exists",
			env:   "",
			input: "",
			files: []string{},
			want:  ".goshipdone.yml",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != "" {
				os.Setenv("GOSHIPDONE_CONFIG", tt.env)
			}
			origFS := defaultFS
			defaultFS = afero.NewMemMapFs()
			for _, filename := range tt.files {
				_ = afero.WriteFile(defaultFS, filename, []byte(filename), 0o644)
			}

			if got := detectFilename(tt.input); got != tt.want {
				t.Errorf("detectFilename() = %q, want %q", got, tt.want)
			}
			defaultFS = origFS
			if tt.env != "" {
				os.Unsetenv("GOSHIPDONE_CONFIG")
			}
		})
	}
}
