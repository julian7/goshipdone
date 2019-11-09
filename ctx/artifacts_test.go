package ctx

import (
	"testing"
)

func TestArtifacts_Add(t *testing.T) {
	type args struct {
		format   int
		name     string
		filename string
		os       string
		arch     string
	}
	tests := []struct {
		name       string
		artifacts  Artifacts
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
			Artifacts{&Artifact{"dist/default", 1, "default", "windows", "amd64"}},
			args{1, "default", "dist/default", "linux", "amd64"},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.artifacts.Add(tt.args.format, tt.args.name, tt.args.filename, tt.args.os, tt.args.arch)

			noArtifacts := len(tt.artifacts)
			if noArtifacts != tt.wantsCount {
				t.Errorf(
					"Artifacts.Add() yielded %d artifact%s, wants = %d",
					noArtifacts,
					map[bool]string{true: "", false: "s"}[noArtifacts == 1],
					tt.wantsCount,
				)
			}
		})
	}
}

func TestArtifacts_ByName(t *testing.T) {
	tests := []struct {
		name      string
		Artifacts Artifacts
		nameArg   string
		wantCount int
	}{
		{"empty", nil, "nonexisting", 0},
		{"notfound", Artifacts{
			&Artifact{Filename: "a/b", Format: 1, Name: "a", OS: "linux", Arch: "386"},
		}, "b", 0},
		{"found", Artifacts{
			&Artifact{Filename: "a/b", Format: 1, Name: "a", OS: "linux", Arch: "386"},
			&Artifact{Filename: "a/b.exe", Format: 1, Name: "a", OS: "windows", Arch: "386"},
			&Artifact{Filename: "a/c", Format: 1, Name: "c", OS: "linux", Arch: "386"},
		}, "a", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.Artifacts.ByName(tt.nameArg); len(got) != tt.wantCount {
				t.Errorf("Artifacts.ByName() = %v, want %v item(s)", got, tt.wantCount)
			}
		})
	}
}
