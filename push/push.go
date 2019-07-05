package push

import (
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	. "github.com/springernature/halfpipe-cf-plugin/helpers"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type push struct {
	cliConnection halfpipe_cf_plugin.CliInterface
}

func NewPushPlanner(cliConnection halfpipe_cf_plugin.CliInterface) plan.Planner {
	return push{
		cliConnection: cliConnection,
	}
}

func (p push) GetPlan(application manifest.Application, request halfpipe_cf_plugin.Request) (pl plan.Plan, err error) {
	currentSpace, err := p.cliConnection.GetCurrentSpace()
	if err != nil {
		err = halfpipe_cf_plugin.ErrGetCurrentSpace(err)
		return
	}

	candidateName := CreateCandidateAppName(application.Name)

	candidateHost := CreateCandidateHostname(application.Name, currentSpace.Name)

	stateError := CheckCFState(application.Name, p.cliConnection)
	if stateError != nil {
		err = stateError
		return
	}

	if application.NoRoute {
		pl = plan.Plan{
			command.NewCfShellCommand(
				"push",
				candidateName,
				"-f", request.ManifestPath,
				"-p", request.AppPath,
			),
		}
	} else {
		pl = plan.Plan{
			command.NewCfShellCommand(
				"push",
				candidateName,
				"-f", request.ManifestPath,
				"-p", request.AppPath,
				"--no-route",
				"--no-start",
			),
			command.NewCfShellCommand(
				"map-route",
				candidateName,
				request.TestDomain,
				"-n", candidateHost,
			),
			//NewCfShellCommand(
			//	"set-health-check",
			//	candidateName,
			//	"http",
			//),
			command.NewCfShellCommand(
				"start",
				candidateName,
			),
		}
	}
	return
}
