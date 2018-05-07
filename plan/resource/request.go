package resource

type Request struct {
	Source Source
	Params Params
}

type Source struct {
	API      string
	Org      string
	Space    string
	Username string
	Password string
}

type Params struct {
	Command      string
	ManifestPath string
	AppPath      string
	TestDomain   string
	Vars         map[string]string
}
