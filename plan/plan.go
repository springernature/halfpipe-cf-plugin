package plan

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan/plans"
	"github.com/springernature/halfpipe-cf-plugin"
)

var ErrUnknownCommand = func(cmd string) error {
	return fmt.Errorf("unknown command '%s'", cmd)
}

var ErrBadManifest = errors.New("Application manifest must contain exactly one application")

type ManifestReader func(pathToManifest string) ([]manifest.Application, error)

type Plan interface {
	GetPlan(request plans.PluginRequest) (commands plans.Plan, err error)
}

type planner struct {
	pushPlan       plans.Planner
	promotePlan    plans.Planner
	manifestReader ManifestReader
}

func NewPlanner(pushPlan plans.Planner, promotePlan plans.Planner, manifestReader ManifestReader) Plan {
	return planner{
		pushPlan:       pushPlan,
		promotePlan:    promotePlan,
		manifestReader: manifestReader,
	}
}

func (c planner) GetPlan(request plans.PluginRequest) (commands plans.Plan, err error) {
	apps, err := c.manifestReader(request.ManifestPath)
	if err != nil {
		return
	}

	if len(apps) != 1 {
		err = ErrBadManifest
		return
	}

	switch request.Command {
	case types.PUSH:
		commands, err = c.pushPlan.GetPlan(apps[0], request)
	case types.PROMOTE:
		commands, err = c.promotePlan.GetPlan(apps[0], request)
	default:
		err = ErrUnknownCommand(request.Command)
	}

	return
}
