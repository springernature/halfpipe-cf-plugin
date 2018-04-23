package plugin

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/plugin/models"
)

var (
	ErrAppNameExists = func(appName string) error {
		errorMessage := fmt.Sprintf("error! App name %s already exists. Please make sure cf is in the right state", appName)
		return errors.New(errorMessage)
	}
	ErrRouteInUse = func(candidateRoute string, testDomain string) error {
		errorMessage := fmt.Sprintf("error! Route is already in use %s.%s. Please make sure cf is in the right state", candidateRoute, testDomain)
		return errors.New(errorMessage)
	}
	ErrAppRunning = func(appName string) error {
		errorMessage := fmt.Sprintf("error! %s is already running. Please make sure cf is in the right state", appName)
		return errors.New(errorMessage)
	}
)

func checkCFState(appName string, testDomain string, candidateRoute string, appsGetter AppsGetter) error {
	apps, e := appsGetter.GetApps()
	candidateAppName := createCandidateAppName(appName)
	deleteAppName := createDeleteName(appName, 0)
	oldAppName := createOldAppName(appName)

	if e != nil {
		return e
	}

	if appExists(apps, candidateAppName) {
		return ErrAppNameExists(candidateAppName)
	}

	if appExists(apps, deleteAppName) {
		return ErrAppNameExists(deleteAppName)
	}

	if routeInUse(apps, testDomain, candidateRoute) {
		return ErrRouteInUse(candidateRoute, testDomain)
	}

	if isAppRunning(apps, oldAppName) {
		return ErrAppRunning(oldAppName)
	}

	if isAppRunning(apps, candidateAppName) {
		return ErrAppRunning(candidateAppName)
	}

	return nil
}

func appExists(apps []plugin_models.GetAppsModel, appName string) bool {
	for _, app := range apps {
		if app.Name == appName {
			return true
		}
	}
	return false
}

func routeInUse(apps []plugin_models.GetAppsModel, domain string, host string) bool {
	for _, app := range apps {
		for _, route := range app.Routes {
			if route.Host == host && route.Domain.Name == domain {
				return true
			}

		}
	}
	return false
}

func isAppRunning(apps []plugin_models.GetAppsModel, appName string) bool {
	for _, app := range apps {
		if app.Name == appName && app.State == "running" {
			return true
		}
	}
	return false
}
