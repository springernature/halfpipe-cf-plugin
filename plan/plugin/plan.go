package plugin

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/plan"
)

var ErrUnknownCommand = func(cmd string) error {
	return fmt.Errorf("unknown command '%s'", cmd)
}

var ErrBadManifest = errors.New("Application manifest must contain exactly one application")

type ManifestReader func(pathToManifest string) ([]manifest.Application, error)

type Planner interface {
	GetPlan(application manifest.Application, request Request) (pl plan.Plan, err error)
}

type Plan interface {
	GetPlan(request Request) (commands plan.Plan, err error)
}

type pluginPlan struct {
	pushPlan       Planner
	promotePlan    Planner
	manifestReader ManifestReader
}

func NewPlanner(pushPlan Planner, promotePlan Planner, manifestReader ManifestReader) Plan {
	return pluginPlan{
		pushPlan:       pushPlan,
		promotePlan:    promotePlan,
		manifestReader: manifestReader,
	}
}

func (c pluginPlan) GetPlan(request Request) (commands plan.Plan, err error) {
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
