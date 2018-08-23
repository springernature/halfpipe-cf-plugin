package plan

import (
	"github.com/springernature/halfpipe-cf-plugin/manifest"
)

type push struct {
	appsGetter CliInterface
}

func NewPushPlanner(appsGetter CliInterface) Planner {
	return push{
		appsGetter: appsGetter,
	}
}

func (p push) GetPlan(application manifest.Application, request Request) (pl Plan, err error) {
	candidateName := createCandidateAppName(application.Name)
	candidateHost := createCandidateHostname(application.Name, request.Space)

	stateError := checkCFState(application.Name, p.appsGetter)
	if stateError != nil {
		err = stateError
		return
	}

	if application.NoRoute {
		pl = Plan{
			NewCfCommand(
				"push",
				candidateName,
				"-f", request.ManifestPath,
				"-p", request.AppPath,
			),
		}
	} else {
		pl = Plan{
			NewCfCommand(
				"push",
				candidateName,
				"-f", request.ManifestPath,
				"-p", request.AppPath,
				"--no-route",
				"--no-start",
			),
			NewCfCommand(
				"map-route",
				candidateName,
				request.TestDomain,
				"-n", candidateHost,
			),
			//NewCfCommand(
			//	"set-health-check",
			//	candidateName,
			//	"http",
			//),
			NewCfCommand(
				"start",
				candidateName,
			),
		}
	}
	return
}
