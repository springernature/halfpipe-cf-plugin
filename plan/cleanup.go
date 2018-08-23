package plan

import (
	"fmt"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"strings"
)

type cleanup struct {
	appGetter CliInterface
}

func (p cleanup) GetPlan(application manifest.Application, request Request) (pl Plan, err error) {

	apps, err := p.appGetter.GetApps()
	if err != nil {
		return
	}

	deleteNamePrefix := createDeleteName(application.Name, 0)
	for _, app := range apps {
		if strings.HasPrefix(app.Name, deleteNamePrefix) {
			pl = append(pl, NewCfCommand("delete", app.Name, "-f"))
		}
	}

	return
}

func (p cleanup) thereIsAnAppToBeDeleted(deleteName string) (delete bool, err error) {
	// This is messy.
	// cliConnection just errors if there is no app.
	// But it doesnt expose a error type for that case, soooooo, string match!
	app, err := p.appGetter.GetApp(deleteName)
	if err != nil && err.Error() == fmt.Sprintf("app %s not found", deleteName) {
		delete = false
		err = nil
		return
	} else if err != nil {
		return
	}

	delete = app.Name == deleteName
	return
}

func NewCleanupPlanner(appGetter CliInterface) Planner {
	return cleanup{
		appGetter: appGetter,
	}
}
