package plugin

import (
	"strings"

	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"code.cloudfoundry.org/cli/plugin/models"
)

type promote struct {
	appsGetter AppsGetter
}

func NewPromotePlanner(appsGetter AppsGetter) Planner {
	return promote{
		appsGetter: appsGetter,
	}
}

func (p promote) GetPlan(application manifest.Application, request Request) (plan plan.Plan, err error) {
	apps, err := p.appsGetter.GetApps()

	if err != nil {
		return
	}

	candidateAppName := createCandidateAppName(application.Name)

	if !application.NoRoute {
		plan = append(plan, addProdRoutes(application, candidateAppName)...)
		plan = append(plan, removeTestRoute(application.Name, request.Space, request.TestDomain))
	}

	plan = append(plan, renameOlderApp(apps, createOldAppName(application.Name), application.Name)...)
	plan = append(plan, renameOldAppAndStopIt(apps, application)...)
	plan = append(plan, renameCandidate(application, candidateAppName))
	return
}

func addProdRoutes(application manifest.Application, candidateAppName string) (commands []plan.Command) {
	for _, route := range application.Routes {
		parts := strings.Split(route, ".")
		hostname := parts[0]
		domain := strings.Join(parts[1:], ".")

		commands = append(commands,
			plan.NewCfCommand("map-route", candidateAppName, domain, "-n", hostname),
		)
	}
	return
}

func removeTestRoute(appName string, space string, testDomain string) plan.Command {
	candidateAppName := createCandidateAppName(appName)
	candidateHostname := createCandidateHostname(appName, space)
	return plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", candidateHostname)
}

func renameCandidate(application manifest.Application, candidateAppName string) plan.Command {
	return plan.NewCfCommand("rename", candidateAppName, application.Name)
}

func renameOldAppAndStopIt(apps []plugin_models.GetAppsModel, currentApp manifest.Application) (pl []plan.Command) {
	for _, app := range apps {
		if app.Name == currentApp.Name {
			oldAppName := createOldAppName(currentApp.Name)
			pl = append(pl, plan.NewCfCommand("rename", currentApp.Name, oldAppName))
			pl = append(pl, plan.NewCfCommand("stop", oldAppName))
		}
	}
	return
}

func renameOlderApp(apps []plugin_models.GetAppsModel, oldAppName string, appName string) (pl []plan.Command) {
	for _, app := range apps {
		if app.Name == oldAppName {
			pl = append(pl, plan.NewCfCommand("rename", app.Name, createDeleteName(appName)))
		}
	}
	return
}
