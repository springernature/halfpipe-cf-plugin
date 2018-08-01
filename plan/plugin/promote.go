package plugin

import (
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/plugin/models"
	"fmt"
	"strings"
)

var ErrCandidateNotRunning = errors.New("Canidate app is not running!")

type promote struct {
	appsGetter AppsGetter
}

func NewPromotePlanner(appsGetter AppsGetter) Planner {
	return promote{
		appsGetter: appsGetter,
	}
}

func (p promote) GetPlan(manifest manifest.Application, request Request) (plan plan.Plan, err error) {
	/*
		We must fetch the app under deployment with "cf app appName-CANDIDATE" as the call to "cf apps" in
		p.GetPreviousAppState does not include path information in the routes..
	*/
	candidateAppState, err := p.getAndVerifyCandidateAppState(manifest.Name)
	if err != nil {
		return
	}

	currentLiveApp, currentOldApp, currentDeleteApp, err := p.GetPreviousAppState(manifest.Name)
	if err != nil {
		return
	}

	domainsInOrg, err := p.getDomainsInOrg(manifest)
	if err != nil {
		return
	}

	plan = append(plan, addManifestRoutes(candidateAppState, manifest.Routes, domainsInOrg)...)
	plan = append(plan, removeTestRoute(candidateAppState, manifest.Name, request.TestDomain, request.Space)...)
	plan = append(plan, renameOldAppToDelete(currentLiveApp, currentOldApp, currentDeleteApp, manifest.Name)...)
	plan = append(plan, renameAndStopCurrentLiveApp(currentLiveApp, currentOldApp)...)
	plan = append(plan, renameCandidateAppToExpectedName(candidateAppState.Name, manifest.Name))

	return
}

func (p promote) getAndVerifyCandidateAppState(manifestAppName string) (app plugin_models.GetAppModel, err error) {
	app, err = p.appsGetter.GetApp(createCandidateAppName(manifestAppName))
	if err != nil {
		return
	}

	if app.State != "started" {
		err = ErrCandidateNotRunning
		return
	}
	return
}

func (p promote) GetPreviousAppState(manifestAppName string) (currentLive, currentOld, currentDelete plugin_models.GetAppsModel, err error) {
	appFinder := func(name string, apps []plugin_models.GetAppsModel) (app plugin_models.GetAppsModel) {
		for _, app := range apps {
			if app.Name == name {
				return app
			}
		}
		return
	}

	apps, err := p.appsGetter.GetApps()
	if err != nil {
		return
	}

	currentLive = appFinder(manifestAppName, apps)
	currentOld = appFinder(createOldAppName(manifestAppName), apps)
	currentDelete = appFinder(createDeleteName(manifestAppName, 0), apps)
	return
}

func (p promote) getDomainsInOrg(manifest manifest.Application) (domains []string, err error) {
	if !manifest.NoRoute && len(manifest.Routes) > 0 {
		output, getErr := p.appsGetter.CliCommandWithoutTerminalOutput("domains")
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

func addManifestRoutes(candidateAppState plugin_models.GetAppModel, routes []manifest.Route, domainsInOrg []string) (pl []plan.Command) {
	bindingToADomain := func(route string, domains []string) bool {
		for _, domain := range domains {
			if route == domain {
				return true
			}
		}
		return false
	}

	alreadyBoundToApp := func(route string, routes []plugin_models.GetApp_RouteSummary) bool {
		parts := strings.Split(route, "/")
		routeWithoutPath := parts[0]
		var path string
		if len(parts) > 1 {
			path = strings.Join(parts[1:], "/")
		}

		var hostname string
		var domain string
		if bindingToADomain(routeWithoutPath, domainsInOrg) {
			domain = routeWithoutPath
		} else {
			bits := strings.Split(routeWithoutPath, ".")
			hostname = bits[0]
			domain = strings.Join(bits[1:], ".")
		}

		for _, r := range routes {
			if r.Host == hostname && r.Path == path && r.Domain.Name == domain {
				return true
			}
		}

		return false
	}

	for _, route := range routes {
		if alreadyBoundToApp(route.Route, candidateAppState.Routes) {
			continue
		}
		parts := strings.Split(route.Route, "/")
		routeWithoutPath := parts[0]
		var path string
		if len(parts) > 1 {
			path = strings.Join(parts[1:], "/")
		}

		args := []string{"map-route"}
		if bindingToADomain(routeWithoutPath, domainsInOrg) {
			args = append(args, []string{candidateAppState.Name, routeWithoutPath}...)
		} else {
			bits := strings.Split(routeWithoutPath, ".")
			hostname := bits[0]
			domain := strings.Join(bits[1:], ".")

			args = append(args, []string{candidateAppState.Name, domain, "-n", hostname}...)
		}

		if path != "" {
			args = append(args, []string{"--path", path}...)
		}
		pl = append(pl, plan.NewCfCommand(args...))
	}
	return
}

func removeTestRoute(candidateAppState plugin_models.GetAppModel, manifestAppName string, testDomain string, space string) (pl []plan.Command) {
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
		pl = append(pl, plan.NewCfCommand("unmap-route", candidateAppState.Name, testDomain, "-n", testHostname))
	}

	return
}

func renameOldAppToDelete(currentLiveApp, oldApp, deleteApp plugin_models.GetAppsModel, manifestAppName string) (pl []plan.Command) {
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

	if deleteApp.Name == "" && oldApp.Name == "" {
		// If there is no old apps with the -DELETE and -OLD postfix.
		// We rust return
		return
	}

	if currentLiveApp.Name != "" && deleteApp.Name != "" && oldApp.Name == "" {
		// $ cf rename appName-OLD appName-DELETE <- Succeeded in a previous run
		// $ cf rename appName appName-OLD <- Failed in a previous run
		return
	}

	pl = append(pl, plan.NewCfCommand("rename", oldApp.Name, createDeleteName(manifestAppName, 0)))
	return
}

func renameAndStopCurrentLiveApp(currentLiveApp, currentOldApp plugin_models.GetAppsModel) (pl []plan.Command) {
	if currentLiveApp.Name == "" && currentOldApp.State == "started" {
		// See TestWorkerAppWithPreviousPromoteFailure.One previously running deployed version.previous promote failed at step [2]
		pl = append(pl, plan.NewCfCommand("stop", currentOldApp.Name))
		return
	}

	if currentLiveApp.Name == "" {
		return
	}

	pl = append(pl, plan.NewCfCommand("rename", currentLiveApp.Name, createOldAppName(currentLiveApp.Name)))

	if currentLiveApp.State == "started" {
		pl = append(pl, plan.NewCfCommand("stop", createOldAppName(currentLiveApp.Name)))
	}
	return
}

func renameCandidateAppToExpectedName(candidateAppName, expectedName string) plan.Command {
	return plan.NewCfCommand("rename", candidateAppName, expectedName)
}
