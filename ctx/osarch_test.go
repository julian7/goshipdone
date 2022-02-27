package ctx_test

import (
	"testing"

	"github.com/julian7/goshipdone/ctx"
)

func TestOsarch_ArchName(t *testing.T) {
	tests := []struct {
		name     string
		oa       ctx.OsArch
		archname string
	}{
		{
			name:     "conventional architecture",
			oa:       ctx.OsArch{OS: "conventional", Arch: "architecture"},
			archname: "architecture",
		},
		{
			name:     "fake arm version",
			oa:       ctx.OsArch{OS: "conventional", Arch: "architecture", ArmVersion: 2},
			archname: "architecture",
		},
		{
			name:     "real arm version",
			oa:       ctx.OsArch{OS: "conventional", Arch: "arm", ArmVersion: 2},
			archname: "armv2",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			archname := tt.oa.ArchName()
			if archname != tt.archname {
				t.Errorf("ArchName returned %q, wants %q", archname, tt.archname)
			}
		})
	}
}

func TestOsarch_String(t *testing.T) {
	tests := []struct {
		name     string
		oa       ctx.OsArch
		archname string
	}{
		{
			name:     "conventional architecture",
			oa:       ctx.OsArch{OS: "conventional", Arch: "architecture"},
			archname: "conventional-architecture",
		},
		{
			name:     "fake arm version",
			oa:       ctx.OsArch{OS: "conventional", Arch: "architecture", ArmVersion: 2},
			archname: "conventional-architecture",
		},
		{
			name:     "real arm version",
			oa:       ctx.OsArch{OS: "conventional", Arch: "arm", ArmVersion: 2},
			archname: "conventional-armv2",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			archname := tt.oa.String()
			if archname != tt.archname {
				t.Errorf("ArchName returned %q, wants %q", archname, tt.archname)
			}
		})
	}
}
