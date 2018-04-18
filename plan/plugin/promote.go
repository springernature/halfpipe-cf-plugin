package plugin

import (
	"strings"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
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

	plan = append(plan, renameOldAppToDelete(apps, createOldAppName(application.Name), application.Name)...)
	plan = append(plan, renameCurrentAppToOldAndStop(apps, application)...)
	plan = append(plan, renameCandidateAppToCurrent(application, candidateAppName))
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

func renameCandidateAppToCurrent(application manifest.Application, candidateAppName string) plan.Command {
	return plan.NewCfCommand("rename", candidateAppName, application.Name)
}

func renameCurrentAppToOldAndStop(apps []plugin_models.GetAppsModel, currentApp manifest.Application) (pl []plan.Command) {
	if appExists(apps, currentApp.Name) {
		oldAppName := createOldAppName(currentApp.Name)
		pl = append(pl, plan.NewCfCommand("rename", currentApp.Name, oldAppName))
		pl = append(pl, plan.NewCfCommand("stop", oldAppName))
	}
	return
}

func renameOldAppToDelete(apps []plugin_models.GetAppsModel, oldAppName string, appName string) (pl []plan.Command) {

	var getDeleteName func(int) string

	getDeleteName = func(index int) string {
		newName := createDeleteName(appName, index)
		if appExists(apps, newName) {
			return getDeleteName(index + 1)
		} else {
			return newName
		}
	}

	if appExists(apps, oldAppName) {
		deleteName := getDeleteName(0)
		pl = append(pl, plan.NewCfCommand("rename", oldAppName, deleteName))
	}

	return
}
