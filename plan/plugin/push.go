package plugin

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type push struct{}

func (p push) GetPlan(application manifest.Application, request Request) (pl plan.Plan, err error) {
	candidateName := createCandidateAppName(application.Name)

	command := plan.NewCfCommand(
		"push",
		candidateName,
		"-f", request.ManifestPath,
		"-p", request.AppPath,
		"-n", candidateName,
		"-d", request.TestDomain,
	)
	pl = append(pl, command)
	return
}

func NewPushPlanner() Planner {
	return push{}
}
