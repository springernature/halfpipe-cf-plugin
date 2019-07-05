package plan

import (
	"fmt"
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/springernature/halfpipe-cf-plugin/manifest"

	"errors"
)

var ErrUnknownCommand = func(cmd string) error {
	return fmt.Errorf("unknown command '%s'", cmd)
}

var ErrBadManifest = errors.New("application manifest must contain exactly one application")

type Planner interface {
	GetPlan(application manifest.Application, request halfpipe_cf_plugin.Request) (pl Plan, err error)
}

type PluginPlan interface {
	GetPlan(request halfpipe_cf_plugin.Request) (commands Plan, err error)
}

type pluginPlan struct {
	pushPlan             Planner
	checkPlan            Planner
	promotePlan          Planner
	cleanupPlan          Planner
	manifestReaderWriter manifest.ReaderWriter
}

func NewPlanner(pushPlan Planner, checkPlan Planner, promotePlan Planner, cleanupPlan Planner, manifestReaderWriter manifest.ReaderWriter) PluginPlan {
	return pluginPlan{
		pushPlan:             pushPlan,
		checkPlan:            checkPlan,
		promotePlan:          promotePlan,
		cleanupPlan:          cleanupPlan,
		manifestReaderWriter: manifestReaderWriter,
	}
}

func (c pluginPlan) GetPlan(request halfpipe_cf_plugin.Request) (commands Plan, err error) {
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
	case config.CHECK:
		commands, err = c.checkPlan.GetPlan(app, request)
	case config.PROMOTE:
		commands, err = c.promotePlan.GetPlan(app, request)
	case config.DELETE, config.CLEANUP:
		commands, err = c.cleanupPlan.GetPlan(app, request)
	default:
		err = ErrUnknownCommand(request.Command)
	}

	return
}
