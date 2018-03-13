package out

type Request struct {
	Source Source
	Params Params
}

type Source struct {
	Api string
	Org string
	Space string
	Username string
	Password string
}

type Params struct {
	ManifestPath string
	AppPath string
}

