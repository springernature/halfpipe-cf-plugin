package plans

type PluginRequest struct {
	Command      string
	ManifestPath string
	AppPath      string
	TestDomain   string
}
