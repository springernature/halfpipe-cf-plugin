package plan

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/plan/plans"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin"
)

var ErrUnknownCommand = func(cmd string) error {
	return errors.New(fmt.Sprintf("Unknown command '%s'", cmd))
}

var ErrBadManifest = errors.New("Application manifest must contain exactly one application")

type planner struct {
	pushPlan    plans.Planner
	promotePlan plans.Planner

	manifestPath   string
	manifestReader func(pathToManifest string) ([]manifest.Application, error)
	appsGetter     plans.AppsGetter
}

func NewPlanner(manifestPath string, appPath string, testDomain string, appsGetter plans.AppsGetter) planner {
	return planner{
		pushPlan:       plans.NewPush(manifestPath, appPath, testDomain),
		promotePlan:    plans.NewPromote(testDomain),
		manifestPath:   manifestPath,
		manifestReader: manifest.ReadAndMergeManifests,
	}
}

func (c planner) GetPlan(command string) (commands plans.Plan, err error) {
	apps, err := c.manifestReader(c.manifestPath)
	if err != nil {
		return
	}

	if len(apps) != 1 {
		err = ErrBadManifest
		return
	}

	switch command {
	case halfpipe_cf_plugin.PUSH:
		commands, err = c.pushPlan.GetPlan(apps[0])
	case halfpipe_cf_plugin.PROMOTE:
		commands, err = c.promotePlan.GetPlan(apps[0])
	default:
		err = ErrUnknownCommand(command)
	}

	return
}
