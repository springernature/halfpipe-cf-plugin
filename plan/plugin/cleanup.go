package plugin

import (
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"fmt"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
)

type cleanup struct {
	appGetter AppGetter
}

func (p cleanup) GetPlan(application manifest.Application, request Request) (pl plan.Plan, err error) {
	deleteName := createDeleteName(application.Name, 0)

	deletableApp, err := p.thereIsAnAppToBeDeleted(deleteName)
	if err != nil {
		return
	}

	if deletableApp {
		command := plan.NewCfCommand(
			"delete",
			deleteName,
			"-f",
		)
		pl = append(pl, command)
	}
	return
}

func (p cleanup) thereIsAnAppToBeDeleted(deleteName string) (delete bool, err error) {
	// This is messy.
	// cliConnection just errors if there is no app.
	// But it doesnt expose a error type for that case, soooooo, string match!
	app, err := p.appGetter.GetApp(deleteName)
	if err != nil && err.Error() == fmt.Sprintf("App %s not found", deleteName) {
		delete = false
		err = nil
		return
	} else if err != nil {
		return
	}

	delete = app.Name == deleteName
	return
}

func NewCleanupPlanner(appGetter AppGetter) Planner {
	return cleanup{
		appGetter: appGetter,
	}
}
