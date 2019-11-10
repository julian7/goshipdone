package ctx

import (
	"testing"
)

func TestArtifacts_Add(t *testing.T) {
	tests := []struct {
		name       string
		artifacts  Artifacts
		arg        *Artifact
		wantsCount int
	}{
		{
			name:       "first",
			artifacts:  nil,
			arg:        &Artifact{Format: 1, Name: "default", Location: "dist/default", Filename: "default", OS: "linux", Arch: "amd64"},
			wantsCount: 1,
		},
		{
			name: "second",
			artifacts: Artifacts{
				&Artifact{Filename: "default", Format: 1, Location: "dist/default", Name: "default", OS: "windows", Arch: "amd64"},
			},
			arg:        &Artifact{Format: 1, Name: "default", Location: "dist/default", Filename: "default", OS: "linux", Arch: "amd64"},
			wantsCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.artifacts.Add(tt.arg)

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
			&Artifact{Location: "dist/a/b", Filename: "a/b", Format: 1, Name: "a", OS: "linux", Arch: "386"},
		}, "b", 0},
		{"found", Artifacts{
			&Artifact{Location: "dist/a/b", Filename: "a/b", Format: 1, Name: "a", OS: "linux", Arch: "386"},
			&Artifact{Location: "dist/a/b.exe", Filename: "a/b.exe", Format: 1, Name: "a", OS: "windows", Arch: "386"},
			&Artifact{Location: "dist/a/c", Filename: "a/c", Format: 1, Name: "c", OS: "linux", Arch: "386"},
		}, "a", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.Artifacts.ByName(tt.nameArg); len(*got) != tt.wantCount {
				t.Errorf("Artifacts.ByName() = %v, want %v item(s)", got, tt.wantCount)
			}
		})
	}
}
