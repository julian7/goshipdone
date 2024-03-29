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
			name:      "first",
			artifacts: nil,
			arg: &Artifact{
				ID:       "default",
				Location: "dist/default",
				Filename: "default",
				OsArch:   &OsArch{OS: "linux", Arch: "amd64"},
			},
			wantsCount: 1,
		},
		{
			name: "second",
			artifacts: Artifacts{
				&Artifact{
					Filename: "default",
					Location: "dist/default",
					ID:       "default",
					OsArch:   &OsArch{OS: "windows", Arch: "amd64"},
				},
			},
			arg: &Artifact{
				ID:       "default",
				Location: "dist/default",
				Filename: "default",
				OsArch:   &OsArch{OS: "linux", Arch: "amd64"},
			},
			wantsCount: 2,
		},
	}
	for _, tt := range tests {
		tt := tt
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

func TestArtifacts_ByID(t *testing.T) {
	tests := []struct {
		name      string
		Artifacts Artifacts
		nameArg   string
		wantCount int
	}{
		{"empty", nil, "nonexisting", 0},
		{"notfound", Artifacts{
			&Artifact{Location: "dist/a/b", Filename: "a/b", ID: "a", OsArch: &OsArch{OS: "linux", Arch: "386"}},
		}, "b", 0},
		{"found", Artifacts{
			&Artifact{Location: "dist/a/b", Filename: "a/b", ID: "a", OsArch: &OsArch{OS: "linux", Arch: "386"}},
			&Artifact{Location: "dist/a/b.exe", Filename: "a/b.exe", ID: "a", OsArch: &OsArch{OS: "windows", Arch: "386"}},
			&Artifact{Location: "dist/a/c", Filename: "a/c", ID: "c", OsArch: &OsArch{OS: "linux", Arch: "386"}},
		}, "a", 2},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.Artifacts.ByID(tt.nameArg); len(*got) != tt.wantCount {
				t.Errorf("Artifacts.ByID() = %v, want %v item(s)", got, tt.wantCount)
			}
		})
	}
}
