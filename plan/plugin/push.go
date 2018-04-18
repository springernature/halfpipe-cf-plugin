package plugin

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type push struct {
	check check
}

func NewPushPlanner(check check) push {
	return push{
		check: check,
	}
}

func (p push) GetPlan(application manifest.Application, request Request) (pl plan.Plan, err error) {
	candidateName := createCandidateAppName(application.Name)
	candidateHost := createCandidateHostname(application.Name, request.Space)

	if ok, stateError := p.check.IsCFInAGoodState(application.Name, request.TestDomain, candidateHost); !ok {
		err = stateError
		return
	}

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
			"-n", candidateHost,
			"-d", request.TestDomain,
		))
	}
	return
}
