package out

type Request struct {
	Params Params
}

type Params struct {
	Api string
	Org string
	Space string
	Username string
	Password string
	ManifestPath string
	AppPath string
}

