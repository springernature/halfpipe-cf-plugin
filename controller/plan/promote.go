package plan

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"fmt"
	"strings"
)

type promote struct {
	testDomain string
	appsGetter AppsGetter
}

func NewPromote(testDomain string) promote {
	return promote{
		testDomain: testDomain,
	}
}

func (p promote) GetPlan(application manifest.Application) (plan Plan, err error) {
	candidateAppName := fmt.Sprintf("%s-CANDIDATE", application.Name)

	plan = append(plan, addProdRoutes(application, candidateAppName)...)
	plan = append(plan, removeTestRoute(application, candidateAppName, p.testDomain))
	plan = append(plan,
		NewCfCommand("rename", candidateAppName, application.Name),
	)
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

func removeTestRoute(application manifest.Application, candidateAppName string, testDomain string) Command {
	return NewCfCommand(
		"unmap-route", candidateAppName,
		testDomain, "-n", candidateAppName)
}
