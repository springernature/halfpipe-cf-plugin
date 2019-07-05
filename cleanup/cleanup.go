package cleanup

import (
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/helpers"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"strings"
)

type cleanup struct {
	appGetter halfpipe_cf_plugin.CliInterface
}

func (p cleanup) GetPlan(application manifest.Application, request halfpipe_cf_plugin.Request) (pl plan.Plan, err error) {

	apps, err := p.appGetter.GetApps()
	if err != nil {
		err = halfpipe_cf_plugin.ErrGetApps(err)
		return
	}

	deleteNamePrefix := helpers.CreateDeleteName(application.Name, 0)
	for _, app := range apps {
		if strings.HasPrefix(app.Name, deleteNamePrefix) {
			pl = append(pl, command.NewCfShellCommand("delete", app.Name, "-f"))
		}
	}

	return
}

func NewCleanupPlanner(appGetter halfpipe_cf_plugin.CliInterface) plan.Planner {
	return cleanup{
		appGetter: appGetter,
	}
}
