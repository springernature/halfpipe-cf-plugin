package plugin

import (
	"fmt"

	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type push struct {
	appsGetter AppsGetter
}

func NewPushPlanner(appsGetter AppsGetter) push {
	return push{
		appsGetter: appsGetter,
	}
}

func (p push) GetPlan(application manifest.Application, request Request) (pl plan.Plan, err error) {
	candidateName := createCandidateAppName(application.Name)
	candidateHost := createCandidateHostname(application.Name, request.Space)

	if ok, stateError := p.IsCFInAGoodState(candidateName, request.TestDomain, candidateHost); !ok {
		err = stateError
		return
	}

	if application.NoRoute {
		pl = append(pl, plan.NewCfCommand(
			"push",
			candidateName,
			"-f", request.ManifestPath,
			"-p", request.AppPath,
		))
	} else {
		pl = append(pl, plan.NewCfCommand(
			"push",
			candidateName,
			"-f", request.ManifestPath,
			"-p", request.AppPath,
			"-n", candidateHost,
			"-d", request.TestDomain,
		))
	}

	return
}

func (p push) IsCFInAGoodState(candidateAppName string, testDomain string, candidateRoute string) (bool, error) {
	apps, e := p.appsGetter.GetApps()
	if e != nil {
		return false, e
	}

	if appExists(apps, candidateAppName) {
		return false, fmt.Errorf("candidate app name already exists %s", candidateAppName)
	}

	if routeExists(apps, testDomain, candidateRoute) {
		return false, fmt.Errorf("test route is already in use %s.%s", candidateRoute, testDomain)
	}

	return true, nil
}
