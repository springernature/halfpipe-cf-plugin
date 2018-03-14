package controller

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin"
)

var ErrUnknownCommand = func(cmd string) error {
	return errors.New(fmt.Sprintf("Unknown command '%s'", cmd))
}

var ErrBadManifest = errors.New("Application manifest must contain exactly one application")

type controller struct {
	command     string
	pushPlan    plan.Planner
	promotePlan plan.Planner

	manifestPath   string
	manifestReader func(pathToManifest string) ([]manifest.Application, error)
	appsGetter     plan.AppsGetter
}

func NewController(command string, manifestPath string, appPath string, testDomain string, appsGetter plan.AppsGetter) controller {
	return controller{
		command:        command,
		pushPlan:       plan.NewPush(manifestPath, appPath, testDomain),
		promotePlan:    plan.NewPromote(testDomain),
		manifestPath:   manifestPath,
		manifestReader: manifest.ReadAndMergeManifests,
	}
}

func (c controller) GetPlan() (commands plan.Plan, err error) {
	apps, err := c.manifestReader(c.manifestPath)
	if err != nil {
		return
	}

	if len(apps) != 1 {
		err = ErrBadManifest
		return
	}

	switch c.command {
	case halfpipe_cf_plugin.PUSH:
		commands, err = c.pushPlan.GetPlan(apps[0])
	case halfpipe_cf_plugin.PROMOTE:
		commands, err = c.promotePlan.GetPlan(apps[0])
	default:
		err = ErrUnknownCommand(c.command)
	}

	return
}
