package plan

import (
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/plugin/models"
	"fmt"
	"strings"
)

var ErrCandidateNotRunning = errors.New("Canidate app is not running!")

type promote struct {
	cliConnection CliInterface
}

func NewPromotePlanner(cliConnection CliInterface) Planner {
	return promote{
		cliConnection: cliConnection,
	}
}

func (p promote) GetPlan(manifest manifest.Application, request Request) (plan Plan, err error) {
	currentSpace, err := p.cliConnection.GetCurrentSpace()
	if err != nil {
		return
	}

	/*
		We must fetch the app under deployment with "cf app appName-CANDIDATE" as the call to "cf apps" in
		p.GetPreviousAppState does not include path information in the routes..
	*/
	candidateAppState, err := p.getAndVerifyCandidateAppState(manifest.Name)
	if err != nil {
		return
	}

	currentLiveApp, currentOldApp, currentDeleteApps, err := p.GetPreviousAppState(manifest.Name)
	if err != nil {
		return
	}

	domainsInOrg, err := p.getDomainsInOrg(manifest)
	if err != nil {
		return
	}

	plan = append(plan, addManifestRoutes(candidateAppState, manifest.Routes, domainsInOrg)...)
	plan = append(plan, removeTestRoute(candidateAppState, manifest.Name, request.TestDomain, currentSpace.Name)...)
	plan = append(plan, renameOldAppToDelete(currentLiveApp, currentOldApp, currentDeleteApps, manifest.Name)...)
	plan = append(plan, renameAndStopCurrentLiveApp(currentLiveApp, currentOldApp)...)
	plan = append(plan, renameCandidateAppToExpectedName(candidateAppState.Name, manifest.Name))

	return
}

func (p promote) getAndVerifyCandidateAppState(manifestAppName string) (app plugin_models.GetAppModel, err error) {
	app, err = p.cliConnection.GetApp(createCandidateAppName(manifestAppName))
	if err != nil {
		return
	}

	if app.State != "started" {
		err = ErrCandidateNotRunning
		return
	}
	return
}

func (p promote) GetPreviousAppState(manifestAppName string) (currentLive, currentOld plugin_models.GetAppsModel, currentDeletes []plugin_models.GetAppsModel, err error) {
	appFinder := func(name string, apps []plugin_models.GetAppsModel) (app plugin_models.GetAppsModel) {
		for _, app := range apps {
			if app.Name == name {
				return app
			}
		}
		return
	}

	deleteAppFinder := func(name string, apps []plugin_models.GetAppsModel) (deleteApps []plugin_models.GetAppsModel) {
		for _, app := range apps {
			if strings.HasPrefix(app.Name, name) {
				deleteApps = append(deleteApps, app)
			}
		}
		return
	}

	apps, err := p.cliConnection.GetApps()
	if err != nil {
		return
	}

	currentLive = appFinder(manifestAppName, apps)
	currentOld = appFinder(createOldAppName(manifestAppName), apps)
	currentDeletes = deleteAppFinder(createDeleteName(manifestAppName, 0), apps)
	return
}

func (p promote) getDomainsInOrg(manifest manifest.Application) (domains []string, err error) {
	if !manifest.NoRoute && len(manifest.Routes) > 0 {
		output, getErr := p.cliConnection.CliCommandWithoutTerminalOutput("domains")
		if getErr != nil {
			err = getErr
			return
		}

		// First two lines are
		// Getting domains in org myOrg as myUser...
		//  name                   					status   type
		for _, domainLine := range output[2:] {
			domain := strings.Split(domainLine, " ")[0]
			domains = append(domains, strings.TrimSpace(domain))
		}
	}
	return
}

func addManifestRoutes(candidateAppState plugin_models.GetAppModel, routes []manifest.Route, domainsInOrg []string) (pl []Command) {
	for _, route := range routes {
		hostname, domain, path := parseRoute(route.Route, domainsInOrg)

		if routeIsBoundToApp(hostname, domain, path, candidateAppState.Routes) {
			continue
		}

		args := []string{
			"map-route",
			candidateAppState.Name,
			domain,
		}

		if hostname != "" {
			args = append(args, []string{"-n", hostname}...)
		}

		if path != "" {
			args = append(args, []string{"--path", path}...)
		}

		pl = append(pl, NewCfCommand(args...))
	}
	return
}

func parseRoute(route string, domainsInOrg []string) (hostname, domain, path string) {
	parts := strings.Split(route, "/")
	routeWithoutPath := parts[0]
	if len(parts) > 1 {
		path = strings.Join(parts[1:], "/")
	}

	if routeIsDomain(routeWithoutPath, domainsInOrg) {
		domain = routeWithoutPath
	} else {
		bits := strings.Split(routeWithoutPath, ".")
		hostname = bits[0]
		domain = strings.Join(bits[1:], ".")
	}
	return
}

func routeIsDomain(route string, domains []string) bool {
	for _, domain := range domains {
		if route == domain {
			return true
		}
	}
	return false
}

func routeIsBoundToApp(hostname, domain, path string, routes []plugin_models.GetApp_RouteSummary) bool {
	for _, r := range routes {
		if r.Host == hostname && r.Path == path && r.Domain.Name == domain {
			return true
		}
	}

	return false
}

func removeTestRoute(candidateAppState plugin_models.GetAppModel, manifestAppName string, testDomain string, space string) (pl []Command) {
	appHasRoute := func(hostname string, domain string, routes []plugin_models.GetApp_RouteSummary) bool {
		for _, route := range routes {
			if route.Host == hostname && route.Domain.Name == domain {
				return true
			}
		}
		return false
	}

	testHostname := fmt.Sprintf("%s-%s-CANDIDATE", manifestAppName, space)
	if appHasRoute(testHostname, testDomain, candidateAppState.Routes) {
		pl = append(pl, NewCfCommand("unmap-route", candidateAppState.Name, testDomain, "-n", testHostname))
	}

	return
}

func renameOldAppToDelete(currentLiveApp, oldApp plugin_models.GetAppsModel, deleteApps []plugin_models.GetAppsModel, manifestAppName string) (pl []Command) {
	/*
	Empty name means the app did not exist
	I.e for the app with name xyz there is no xyz-OLD and xyz-DELETE
	This function is confusing and complex, have a look at the tests cases
	* TestAppWithRouteWhenPreviousPromoteFailure
	* TestWorkerAppWithPreviousPromoteFailure
	 */
	if currentLiveApp.Name == "" && oldApp.Name != "" {
		// 		$ cf rename appName appName-OLD <- Succeeded in a previous run
		//		$ cf stop appName-OLD <- Failed in a previous run
		return
	}

	if len(deleteApps) == 0 && oldApp.Name == "" {
		// If there is no old apps with the -DELETE and -OLD postfix.
		// We rust return
		return
	}

	if currentLiveApp.Name != "" && len(deleteApps) > 0 && oldApp.Name == "" {
		// $ cf rename appName-OLD appName-DELETE <- Succeeded in a previous run
		// $ cf rename appName appName-OLD <- Failed in a previous run
		return
	}

	pl = append(pl, NewCfCommand("rename", oldApp.Name, createDeleteName(manifestAppName, len(deleteApps))))
	return
}

func renameAndStopCurrentLiveApp(currentLiveApp, currentOldApp plugin_models.GetAppsModel) (pl []Command) {
	if currentLiveApp.Name == "" && currentOldApp.State == "started" {
		// See TestWorkerAppWithPreviousPromoteFailure.One previously running deployed version.previous promote failed at step [2]
		pl = append(pl, NewCfCommand("stop", currentOldApp.Name))
		return
	}

	if currentLiveApp.Name == "" {
		return
	}

	pl = append(pl, NewCfCommand("rename", currentLiveApp.Name, createOldAppName(currentLiveApp.Name)))

	if currentLiveApp.State == "started" {
		pl = append(pl, NewCfCommand("stop", createOldAppName(currentLiveApp.Name)))
	}
	return
}

func renameCandidateAppToExpectedName(candidateAppName, expectedName string) Command {
	return NewCfCommand("rename", candidateAppName, expectedName)
}
