package plan

import "time"

type Request struct {
	Command      string
	ManifestPath string
	AppPath      string
	TestDomain   string
	Space        string
	Timeout      time.Duration
}
