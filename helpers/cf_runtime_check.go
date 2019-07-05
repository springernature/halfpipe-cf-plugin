package helpers

import (
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/plugin/models"
	"fmt"
	"github.com/springernature/halfpipe-cf-plugin"
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

func CheckCFState(appName string, cliConnection halfpipe_cf_plugin.CliInterface) error {
	apps, err := cliConnection.GetApps()
	if err != nil {
		return halfpipe_cf_plugin.ErrGetApps(err)
	}

	oldAppName := CreateOldAppName(appName)
	if isAppRunning(apps, oldAppName) {
		return ErrAppRunning(oldAppName)
	}

	return nil
}

func isAppRunning(apps []plugin_models.GetAppsModel, appName string) bool {
	for _, app := range apps {
		if app.Name == appName && app.State == "running" {
			return true
		}
	}
	return false
}
