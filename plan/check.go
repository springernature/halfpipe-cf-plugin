package plan

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/plugin/models"
)

var (
	ErrAppNameExists = func(appName string) error {
		errorMessage := fmt.Sprintf("error! App name %s already exists, please delete it before retriggering this job! 'cf delete %s'", appName, appName)
		return errors.New(errorMessage)
	}
	ErrAppRunning = func(appName string) error {
		errorMessage := fmt.Sprintf("error! %s is already running, please stop it before retriggering this job! 'cf stop %s'", appName, appName)
		return errors.New(errorMessage)
	}
)

func checkCFState(appName string, testDomain string, candidateRoute string, appsGetter AppsGetter) error {
	apps, e := appsGetter.GetApps()

	deleteAppName := createDeleteName(appName, 0)
	oldAppName := createOldAppName(appName)

	if e != nil {
		return e
	}

	if appExists(apps, deleteAppName) {
		return ErrAppNameExists(deleteAppName)
	}

	if isAppRunning(apps, oldAppName) {
		return ErrAppRunning(oldAppName)
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

func isAppRunning(apps []plugin_models.GetAppsModel, appName string) bool {
	for _, app := range apps {
		if app.Name == appName && app.State == "running" {
			return true
		}
	}
	return false
}
