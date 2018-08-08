package plan

type Request struct {
	Command      string
	ManifestPath string
	AppPath      string
	TestDomain   string
	Space        string
}
