package ctx

import "fmt"

type OsArch struct {
	OS         string
	Arch       string
	ArmVersion int32
}

func (oa *OsArch) ArchName() string {
	if oa.Arch == "arm" && oa.ArmVersion > 0 {
		return fmt.Sprintf("%sv%d", oa.Arch, oa.ArmVersion)
	}

	return oa.Arch
}

func (oa *OsArch) String() string {
	return fmt.Sprintf("%s-%s", oa.OS, oa.ArchName())
}
