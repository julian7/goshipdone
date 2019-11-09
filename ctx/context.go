package ctx

// Context are a cumulative structure carried over to each module,
// to contain data later steps might require
type Context struct {
	ProjectName string
	TargetDir   string
	Version     string
	Artifacts   Artifacts
}
