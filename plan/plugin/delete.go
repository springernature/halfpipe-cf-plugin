package plugin

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type delete struct {
	appGetter AppGetter
}

func (p delete) GetPlan(application manifest.Application, request Request) (pl plan.Plan, err error) {
	deleteName := createDeleteName(application.Name)
	_, err = p.appGetter.GetApp(deleteName)

	if err != nil {
		return
	}

	command := plan.NewCfCommand(
		"delete",
		deleteName,
		"-f",
	)
	pl = append(pl, command)
	return
}

func NewDeletePlanner(appGetter AppGetter) Planner {
	return delete{
		appGetter: appGetter,
	}
}
