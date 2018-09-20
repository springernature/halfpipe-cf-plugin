package plan

import (
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"strings"
)

type cleanup struct {
	appGetter CliInterface
}

func (p cleanup) GetPlan(application manifest.Application, request Request) (pl Plan, err error) {

	apps, err := p.appGetter.GetApps()
	if err != nil {
		err = ErrGetApps(err)
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

func NewCleanupPlanner(appGetter CliInterface) Planner {
	return cleanup{
		appGetter: appGetter,
	}
}
