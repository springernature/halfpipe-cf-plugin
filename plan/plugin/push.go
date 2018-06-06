package plugin

import (
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
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
		pl = plan.Plan{
			plan.NewCfCommand(
				"push",
				candidateName,
				"-f", request.ManifestPath,
				"-p", request.AppPath,
			),
		}
	} else {
		pl = plan.Plan{
			plan.NewCfCommand(
				"push",
				candidateName,
				"-f", request.ManifestPath,
				"-p", request.AppPath,
				"--no-route",
				"--no-start",
			),
			plan.NewCfCommand(
				"map-route",
				candidateName,
				request.TestDomain,
				"-n", candidateHost,
			),
			//plan.NewCfCommand(
			//	"set-health-check",
			//	candidateName,
			//	"http",
			//),
			plan.NewCfCommand(
				"start",
				candidateName,
			),
		}
	}
	return
}
