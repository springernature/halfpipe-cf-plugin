package plans

import (
	"strings"

	"code.cloudfoundry.org/cli/util/manifest"
)

type promote struct{}

func NewPromote() Planner {
	return promote{}
}

func (p promote) GetPlan(application manifest.Application, request PluginRequest) (plan Plan, err error) {
	candidateAppName := createCandidateAppName(application.Name)

	plan = append(plan, addProdRoutes(application, candidateAppName)...)
	plan = append(plan, removeTestRoute(candidateAppName, request.TestDomain))
	plan = append(plan, renameCandidate(application, candidateAppName))
	return
}

func addProdRoutes(application manifest.Application, candidateAppName string) (commands []Command) {
	for _, route := range application.Routes {
		parts := strings.Split(route, ".")
		hostname := parts[0]
		domain := strings.Join(parts[1:], ".")

		commands = append(commands,
			NewCfCommand("map-route", candidateAppName, domain, "-n", hostname),
		)
	}

	return
}

func removeTestRoute(candidateAppName string, testDomain string) Command {
	return NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", candidateAppName)
}

func renameCandidate(application manifest.Application, candidateAppName string) Command {
	return NewCfCommand("rename", candidateAppName, application.Name)
}
