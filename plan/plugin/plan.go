package plugin

import (
	"fmt"

	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/springernature/halfpipe-cf-plugin/manifest"

	"errors"
)

var ErrUnknownCommand = func(cmd string) error {
	return fmt.Errorf("unknown command '%s'", cmd)
}

var ErrBadManifest = errors.New("Application manifest must contain exactly one application")

type Planner interface {
	GetPlan(application manifest.Application, request Request) (pl plan.Plan, err error)
}

type Plan interface {
	GetPlan(request Request) (commands plan.Plan, err error)
}

type pluginPlan struct {
	pushPlan             Planner
	promotePlan          Planner
	cleanupPlan          Planner
	manifestReaderWriter manifest.ManifestReaderWriter
}

func NewPlanner(pushPlan Planner, promotePlan Planner, cleanupPlan Planner, manifestReaderWriter manifest.ManifestReaderWriter) Plan {
	return pluginPlan{
		pushPlan:             pushPlan,
		promotePlan:          promotePlan,
		cleanupPlan:          cleanupPlan,
		manifestReaderWriter: manifestReaderWriter,
	}
}

func (c pluginPlan) GetPlan(request Request) (commands plan.Plan, err error) {
	man, err := c.manifestReaderWriter.ReadManifest(request.ManifestPath)
	if err != nil {
		return
	}

	if len(man.Applications) != 1 {
		err = ErrBadManifest
		return
	}

	app := man.Applications[0]
	switch request.Command {
	case config.PUSH:
		commands, err = c.pushPlan.GetPlan(app, request)
	case config.PROMOTE:
		commands, err = c.promotePlan.GetPlan(app, request)
	case config.DELETE, config.CLEANUP:
		commands, err = c.cleanupPlan.GetPlan(app, request)
	default:
		err = ErrUnknownCommand(request.Command)
	}

	return
}
