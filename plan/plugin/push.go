package plugin

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type push struct{}

func (p push) GetPlan(application manifest.Application, request Request) (pl plan.Plan, err error) {
	candidateName := createCandidateAppName(application.Name)

	if application.NoRoute {
		pl = append(pl, plan.NewCfCommand(
			"push",
			candidateName,
			"-f", request.ManifestPath,
			"-p", request.AppPath,
		))
	} else {
		pl = append(pl, plan.NewCfCommand(
			"push",
			candidateName,
			"-f", request.ManifestPath,
			"-p", request.AppPath,
			"-n", createCandidateHostname(application.Name, request.Space),
			"-d", request.TestDomain,
		))
	}

	return
}

func NewPushPlanner() Planner {
	return push{}
}
