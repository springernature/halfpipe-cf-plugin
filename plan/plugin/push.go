package plugin

import (
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
)

type push struct {
	appsGetter AppsGetter
}

func NewPushPlanner(appsGetter AppsGetter) Planner {
	return push{
		appsGetter: appsGetter,
	}
}

func (p push) GetPlan(application manifest.Application, request Request) (pl plan.Plan, err error) {
	candidateName := createCandidateAppName(application.Name)
	candidateHost := createCandidateHostname(application.Name, request.Space)

	stateError := checkCFState(application.Name, request.TestDomain, candidateHost, p.appsGetter)
	if stateError != nil {
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
