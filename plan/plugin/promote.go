package plugin

import (
	"strings"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
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
		domains, domainsErr := p.getDomainsInOrg()
		if domainsErr != nil {
			err = domainsErr
			return
		}

		plan = append(plan, addProdRoutes(application, candidateAppName, domains)...)
		plan = append(plan, removeTestRoute(application.Name, request.Space, request.TestDomain))
	}

	plan = append(plan, renameOldAppToDelete(apps, createOldAppName(application.Name), application.Name)...)
	plan = append(plan, renameCurrentAppToOldAndStop(apps, application)...)
	plan = append(plan, renameCandidateAppToCurrent(application, candidateAppName))
	return
}

func (p promote) getDomainsInOrg() (domains []string, err error) {
	output, err := p.appsGetter.CliCommandWithoutTerminalOutput("domains")
	if err != nil {
		return
	}

	// First two lines ar
	// Getting domains in org myOrg as myUser...
	//  name                   					status   type
	for _, domainLine := range output[2:] {
		domain := strings.Split(domainLine, " ")[0]
		domains = append(domains, strings.TrimSpace(domain))
	}
	return
}

func addProdRoutes(application manifest.Application, candidateAppName string, domains []string) (commands []plan.Command) {
	for _, route := range application.Routes {
		if routeIsDomain(route.Route, domains) {
			commands = append(commands,
				plan.NewCfCommand("map-route", candidateAppName, route.Route),
			)
		} else {
			parts := strings.Split(route.Route, ".")
			hostname := parts[0]
			domain := strings.Join(parts[1:], ".")

			commands = append(commands,
				plan.NewCfCommand("map-route", candidateAppName, domain, "-n", hostname),
			)
		}
	}
	return
}

func routeIsDomain(route string, domains []string) bool {
	for _, domain := range domains {
		if strings.TrimSpace(domain) == strings.TrimSpace(route) {
			return true
		}
	}
	return false
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
