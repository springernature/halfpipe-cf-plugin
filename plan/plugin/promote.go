package plugin

import (
	"strings"

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

	plan = append(plan, addProdRoutes(application, candidateAppName)...)
	plan = append(plan, removeTestRoute(candidateAppName, request.TestDomain))
	if len(apps) != 0 {
		plan = append(plan, renameOldApp(application.Name))
	}
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

func removeTestRoute(candidateAppName string, testDomain string) plan.Command {
	return plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", candidateAppName)
}

func renameCandidate(application manifest.Application, candidateAppName string) plan.Command {
	return plan.NewCfCommand("rename", candidateAppName, application.Name)
}

func renameOldApp(appName string) plan.Command {
	return plan.NewCfCommand("rename", appName, createOldAppName(appName))
}
