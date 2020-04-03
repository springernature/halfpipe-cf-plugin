package push

import (
	"strconv"
	"strings"

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

	pushArgs := []string{
		"push",
		candidateName,
		"-f", request.ManifestPath,
		//"-p", request.AppPath,
	}
	if application.Docker.Image == "" {
		pushArgs = append(pushArgs, "-p", request.AppPath)
	}

	if application.NoRoute {
		pl = plan.Plan{
			command.NewCfShellCommand(pushArgs...),
		}
	} else {
		if request.Instances != 0 {
			pushArgs = append(pushArgs, "-i", strconv.Itoa(request.Instances))
		}
		pushArgs = append(pushArgs, "--no-route", "--no-start")

		pl = plan.Plan{
			command.NewCfShellCommand(pushArgs...),
			command.NewCfShellCommand(
				"map-route",
				candidateName,
				request.TestDomain,
				"-n", candidateHost,
			),
		}

		if len(request.PreStartCommand) > 0 {
			for _, preStartCommand := range strings.Split(request.PreStartCommand, ";") {
				trimmedCommand := strings.TrimSpace(preStartCommand)
				if strings.HasPrefix(trimmedCommand, "cf ") {
					pl = append(pl, command.NewCfShellCommand(strings.Split(trimmedCommand, " ")[1:]...))
				}
			}
		}

		pl = append(pl, command.NewCfShellCommand("start", candidateName))

	}
	return
}
