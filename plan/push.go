package plan

import (
	"github.com/springernature/halfpipe-cf-plugin/manifest"
)

type push struct {
	cliConnection CliInterface
}

func NewPushPlanner(cliConnection CliInterface) Planner {
	return push{
		cliConnection: cliConnection,
	}
}

func (p push) GetPlan(application manifest.Application, request Request) (pl Plan, err error) {
	currentSpace, err := p.cliConnection.GetCurrentSpace()
	if err != nil {
		return
	}

	candidateName := createCandidateAppName(application.Name)

	candidateHost := createCandidateHostname(application.Name, currentSpace.Name)

	stateError := checkCFState(application.Name, p.cliConnection)
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
