package plans

import (
	"code.cloudfoundry.org/cli/util/manifest"
)

type push struct{}

func (p push) GetPlan(application manifest.Application, request PluginRequest) (plan Plan, err error) {
	candidateName := createCandidateAppName(application.Name)

	command := NewCfCommand(
		"push",
		candidateName,
		"-f", request.ManifestPath,
		"-p", request.AppPath,
		"-n", candidateName,
		"-d", request.TestDomain,
	)
	plan = append(plan, command)
	return
}

func NewPush() Planner {
	return push{}
}
