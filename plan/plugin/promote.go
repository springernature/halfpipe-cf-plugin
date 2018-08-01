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

	currentLiveApp, currentOldApp, err := p.GetPreviousAppState(manifest.Name)
	if err != nil {
		return
	}

	domainsInOrg, err := p.getDomainsInOrg(manifest)
	if err != nil {
		return
	}

	plan = append(plan, addManifestRoutes(candidateAppState, manifest.Routes, domainsInOrg)...)
	plan = append(plan, removeTestRoute(candidateAppState, manifest.Name, request.TestDomain, request.Space)...)
	plan = append(plan, renameOldAppToDelete(currentOldApp, manifest.Name)...)
	plan = append(plan, renameAndStopCurrentLiveApp(currentLiveApp)...)
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

func (p promote) GetPreviousAppState(manifestAppName string) (currentLive, currentOld plugin_models.GetAppsModel, err error) {
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

	for _, route := range routes {
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

func renameOldAppToDelete(oldApp plugin_models.GetAppsModel, manifestAppName string) (pl []plan.Command) {
	if oldApp.Name == "" {
		// Empty name means the app didn't exist
		return
	}
	pl = append(pl, plan.NewCfCommand("rename", oldApp.Name, createDeleteName(manifestAppName, 0)))
	return
}

func renameAndStopCurrentLiveApp(currentLiveApp plugin_models.GetAppsModel) (pl []plan.Command) {
	if currentLiveApp.Name == "" {
		// Empty name means the app didn't exist
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
