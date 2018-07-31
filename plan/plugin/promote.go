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
	candidateAppName := createCandidateAppName(application.Name)

	currentCandidateAppState, err := p.appsGetter.GetApp(candidateAppName)
	if err != nil {
		return
	}

	apps, err := p.appsGetter.GetApps()
	if err != nil {
		return
	}

	if !application.NoRoute {
		domains, domainsErr := p.getDomainsInOrg()
		if domainsErr != nil {
			err = domainsErr
			return
		}

		plan = append(plan, addProdRoutes(application, currentCandidateAppState.Routes, candidateAppName, domains)...)

		if routeAlreadyMapped(createCandidateHostname(application.Name, request.Space), request.TestDomain, "", currentCandidateAppState.Routes) {
			plan = append(plan, removeTestRoute(application.Name, request.Space, request.TestDomain))
		}
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

func addProdRoutes(application manifest.Application, currentCandidateAppRoutes []plugin_models.GetApp_RouteSummary, candidateAppName string, domains []string) (commands []plan.Command) {
	for _, route := range application.Routes {
		alreadyMapped := false
		routeSplitByPath := strings.Split(strings.TrimSpace(route.Route), "/")
		hostnameAndDomain := routeSplitByPath[0]
		var path string
		if len(routeSplitByPath) > 1 {
			path = strings.Join(routeSplitByPath[1:], "/")
		}

		args := []string{"map-route", candidateAppName}

		if routeIsDomain(hostnameAndDomain, domains) {
			alreadyMapped = routeAlreadyMapped("", hostnameAndDomain, path, currentCandidateAppRoutes)
			args = append(args, hostnameAndDomain)
		} else {
			parts := strings.Split(hostnameAndDomain, ".")
			hostname := parts[0]
			domain := strings.Join(parts[1:], ".")
			alreadyMapped = routeAlreadyMapped(hostname, domain, path, currentCandidateAppRoutes)
			args = append(args, domain, "-n", hostname)
		}

		if path != "" {
			args = append(args, "--path", path)
		}

		if !alreadyMapped {
			commands = append(commands, plan.NewCfCommand(args...))
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

func routeAlreadyMapped(hostname string, domain string, path string, routes []plugin_models.GetApp_RouteSummary) bool {
	for _, route := range routes {
		if route.Host == hostname && route.Domain.Name == domain && route.Path == path {
			return true
		}
	}
	return false
}
