package modules

import (
	"testing"

	"github.com/julian7/magelib/ctx"
)

func Test_errNumTargets(t *testing.T) {
	tests := []struct {
		name   string
		bad    string
		good   string
		builds map[string]*ctx.Artifacts
		errStr string
	}{
		{
			name:   "empty",
			bad:    "",
			good:   "",
			builds: map[string]*ctx.Artifacts{},
			errStr: "no builds found",
		},
		{
			name: "build not found",
			bad:  "bad",
			good: "good",
			builds: map[string]*ctx.Artifacts{
				"good": &ctx.Artifacts{
					&ctx.Artifact{Location: "dist/good-linux-amd64/good", Filename: "good", Format: 1, Name: "good", OS: "linux", Arch: "amd64"},
					&ctx.Artifact{Location: "dist/good-windows-amd64/good.exe", Filename: "good.exe", Format: 1, Name: "good", OS: "windows", Arch: "amd64"},
				},
			},
			errStr: "no targets found for builds linux-amd64, windows-amd64",
		},
		{
			name: "build missing",
			bad:  "bad",
			good: "good",
			builds: map[string]*ctx.Artifacts{
				"good": &ctx.Artifacts{
					&ctx.Artifact{Location: "dist/good-linux-amd64/good", Filename: "good", Format: 1, Name: "good", OS: "linux", Arch: "amd64"},
					&ctx.Artifact{Location: "dist/good-windows-amd64/good.exe", Filename: "good.exe", Format: 1, Name: "good", OS: "windows", Arch: "amd64"},
				},
				"bad": &ctx.Artifacts{
					&ctx.Artifact{Location: "dist/bad-linux-amd64/bad", Filename: "bad", Format: 1, Name: "bad", OS: "linux", Arch: "amd64"},
				},
			},
			errStr: "build bad is missing os-arch target windows-amd64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errNumTargets(tt.bad, tt.good, tt.builds)
			if err == nil {
				t.Errorf("unexpected succes")
			}
			if err.Error() != tt.errStr {
				t.Errorf("errNumTargets() error = %v, want error string %s", err, tt.errStr)
			}
		})
	}
}
