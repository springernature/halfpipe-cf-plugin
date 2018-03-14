package controller

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
	"code.cloudfoundry.org/cli/util/manifest"
)

var ErrUnknownCommand = func(cmd string) error {
	return errors.New(fmt.Sprintf("Unknown command '%s'", cmd))
}

var ErrBadManifest = errors.New("Application manifest must contain exactly one application")

type controller struct {
	command  string
	pushPlan plan.Planner

	manifestPath   string
	manifestReader func(pathToManifest string) ([]manifest.Application, error)
}

func NewController(command string, manifestPath string, appPath string, testDomain string) controller {
	return controller{
		command:        command,
		pushPlan:       plan.NewPush(manifestPath, appPath, testDomain),
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
	case "halfpipe-push":
		commands, err = c.pushPlan.GetPlan(apps[0])
	default:
		err = ErrUnknownCommand(c.command)
	}

	return
}
