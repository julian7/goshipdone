package modules

import (
	"testing"
)

func TestResults_AddArtifact(t *testing.T) {
	type args struct {
		format   int
		name     string
		filename string
		os       string
		arch     string
	}
	tests := []struct {
		name       string
		artifacts  []Artifact
		args       args
		wantsCount int
	}{
		{
			"first",
			nil,
			args{1, "default", "dist/default", "linux", "amd64"},
			1,
		},
		{
			"second",
			[]Artifact{{"dist/default", 1, "default", "windows", "amd64"}},
			args{1, "default", "dist/default", "linux", "amd64"},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &Results{Artifacts: tt.artifacts}
			res.AddArtifact(tt.args.format, tt.args.name, tt.args.filename, tt.args.os, tt.args.arch)

			noArtifacts := len(res.Artifacts)
			if noArtifacts != tt.wantsCount {
				t.Errorf(
					"res.AddArtifact() yielded %d artifact%s, wants = %d",
					noArtifacts,
					map[bool]string{true: "", false: "s"}[noArtifacts == 1],
					tt.wantsCount,
				)
			}
		})
	}
}

func TestResults_ArtifactsByName(t *testing.T) {
	tests := []struct {
		name      string
		Artifacts []Artifact
		nameArg   string
		wantCount int
	}{
		{"empty", nil, "nonexisting", 0},
		{"notfound", []Artifact{
			{Filename: "a/b", Format: 1, Name: "a", OS: "linux", Arch: "386"},
		}, "b", 0},
		{"found", []Artifact{
			{Filename: "a/b", Format: 1, Name: "a", OS: "linux", Arch: "386"},
			{Filename: "a/b.exe", Format: 1, Name: "a", OS: "windows", Arch: "386"},
			{Filename: "a/c", Format: 1, Name: "c", OS: "linux", Arch: "386"},
		}, "a", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &Results{Artifacts: tt.Artifacts}
			if got := res.ArtifactsByName(tt.nameArg); len(got) != tt.wantCount {
				t.Errorf("Results.ArtifactsByName() = %v, want %v item(s)", got, tt.wantCount)
			}
		})
	}
}
