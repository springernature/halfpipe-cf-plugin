package plugin

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin/models"
)

type check struct {
	appsGetter AppsGetter
}

func NewCheck(appsGetter AppsGetter) check {
	return check{
		appsGetter: appsGetter,
	}
}

func (c check) IsCFInAGoodState(appName string, testDomain string, candidateRoute string) (bool, error) {
	apps, e := c.appsGetter.GetApps()
	candidateAppName := createCandidateAppName(appName)
	deleteAppName := createDeleteName(appName, 0)
	oldAppName := createOldAppName(appName)

	if e != nil {
		return false, e
	}

	if appExists(apps, candidateAppName) {
		return false, fmt.Errorf("error! Candidate app name already exists %s. Please make sure cf is in the right state", candidateAppName)
	}

	if appExists(apps, deleteAppName) {
		return false, fmt.Errorf("error! Delete app name already exists %s. Please make sure cf is in the right state", deleteAppName)
	}

	if routeExists(apps, testDomain, candidateRoute) {
		return false, fmt.Errorf("error! test route is already in use %s.%s. Please make sure cf is in the right state", candidateRoute, testDomain)
	}

	if isOldAppRunning(apps, oldAppName) {
		return false, fmt.Errorf("error! test route is already in use %s.%s. Please make sure cf is in the right state", candidateRoute, testDomain)
	}

	return true, nil
}

func appExists(apps []plugin_models.GetAppsModel, appName string) bool {
	for _, app := range apps {
		if app.Name == appName {
			return true
		}
	}
	return false
}

func routeExists(apps []plugin_models.GetAppsModel, domain string, host string) bool {
	for _, app := range apps {
		for _, route := range app.Routes {
			if route.Host == host && route.Domain.Name == domain {
				return true
			}

		}
	}
	return false
}

func isOldAppRunning(apps []plugin_models.GetAppsModel, oldAppName string) bool {
	for _, app := range apps {
		if app.Name == oldAppName && app.State == "running" {
			return true
		}
	}
	return false
}
