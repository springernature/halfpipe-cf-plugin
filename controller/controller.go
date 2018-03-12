package controller

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
)

var ErrUnknownCommand = func(cmd string) error {
	return errors.New(fmt.Sprintf("Unknown command '%s'", cmd))
}

type controller struct {
	command  string
	pushPlan plan.Planner
}

func NewController(command string, manifestPath string, appPath string) controller {
	return controller{
		command:  command,
		pushPlan: plan.NewPush(manifestPath, appPath),
	}
}

func (c controller) Run() (commands plan.Plan, err error) {
	switch c.command {
	case "halfpipe-push":
		commands, err = c.pushPlan.Commands()
	default:
		err = ErrUnknownCommand(c.command)
	}

	return
}
